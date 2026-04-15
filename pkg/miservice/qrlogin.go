package miservice

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"micli/pkg/util"

	"github.com/mdp/qrterminal/v3"
	"github.com/pterm/pterm"
)

// QRLogin performs QR code login using the Mi Home app.
// Returns authentication data on success, error on failure.
func (s *Service) QRLogin() (*Tokens, error) {
	// Ensure token is initialized
	if s.token == nil {
		s.token = NewTokens()
		s.token.UserName = s.username
		s.token.DeviceId = strings.ToUpper(util.GetRandom(16))
	}

	// Try to load existing token from store
	if s.tokenStore != nil && s.token.PassToken == "" {
		if tokens, err := s.tokenStore.LoadToken(); err == nil {
			if tokens.UserName == s.username {
				s.token = tokens
			}
		}
	}

	// Step 1: Get location params from serviceLogin
	locationData, err := s.getLocation()
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}

	// If token is still valid, no need to re-login
	if locationData["code"] == "0" && locationData["message"] == "Token refresh successful" {
		pterm.Info.Println("Token is still valid, no need to re-login")
		// Ensure mijia SID exists when reusing a valid token
		if _, ok := s.token.Sids[MiioSid]; !ok {
			s.token.Sids[MiioSid] = SidToken{}
		}
		return s.token, nil
	}

	// Step 2: Build QR code URL and get QR data
	qrParams := url.Values{}
	qrParams.Set("theme", "")
	qrParams.Set("bizDeviceType", "")
	qrParams.Set("_hasLogo", "false")
	qrParams.Set("_qrsize", "240")
	qrParams.Set("_dc", fmt.Sprintf("%d", time.Now().UnixMilli()))

	// Copy location params to QR params
	for k, v := range locationData {
		if qrParams.Get(k) == "" {
			qrParams.Set(k, v)
		}
	}

	loginURL := fmt.Sprintf("https://account.xiaomi.com/longPolling/loginUrl?%s", qrParams.Encode())

	headers := http.Header{
		"User-Agent":   []string{s.buildUA()},
		"Content-Type": []string{"application/x-www-form-urlencoded"},
		"Connection":   []string{"keep-alive"},
	}

	req, err := http.NewRequest(http.MethodGet, loginURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR request: %w", err)
	}
	req.Header = headers

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get QR code URL: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var qrResp qrCodeResponse
	// Response starts with &&&START&&&
	jsonStr := strings.TrimPrefix(string(body), "&&&START&&&")
	if err := json.Unmarshal([]byte(jsonStr), &qrResp); err != nil {
		return nil, fmt.Errorf("failed to parse QR response: %w, body: %s", err, string(body))
	}

	if qrResp.Code != 0 {
		return nil, fmt.Errorf("QR code request failed: code=%d, desc=%s", qrResp.Code, qrResp.Desc)
	}

	// Step 3: Print QR code to terminal
	pterm.Info.Println("Please scan the QR code below using the Mi Home app:")
	qrterminal.GenerateHalfBlock(qrResp.LoginURL, qrterminal.L, os.Stdout)
	pterm.Info.Printf("Or visit this link to view the QR code: %s\n", qrResp.QR)

	// Step 4: Long polling for scan result
	pterm.Info.Println("Waiting for QR code scan...")
	client := &http.Client{Timeout: 125 * time.Second}

	pollReq, err := http.NewRequest(http.MethodGet, qrResp.LP, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create poll request: %w", err)
	}
	pollReq.Header = headers

	pollResp, err := client.Do(pollReq)
	if err != nil {
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
			return nil, fmt.Errorf("QR code scan timed out, please try again")
		}
		return nil, fmt.Errorf("poll request failed: %w", err)
	}
	defer pollResp.Body.Close()

	pollBody, err := io.ReadAll(pollResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read poll response: %w", err)
	}

	var pollData loginResp
	jsonStr = strings.TrimPrefix(string(pollBody), "&&&START&&&")
	if err := json.Unmarshal([]byte(jsonStr), &pollData); err != nil {
		return nil, fmt.Errorf("failed to parse poll response: %w, body: %s", err, string(pollBody))
	}

	if pollData.Code != 0 {
		return nil, fmt.Errorf("login failed: code=%d, desc=%s", pollData.Code, pollData.Desc)
	}

	// Step 5: Process login result
	s.token.SSecurity = pollData.Psecurity
	s.token.Nonce = fmt.Sprintf("%d", pollData.Nonce)
	s.token.Sids[MinaSid] = SidToken{SSecurity: pollData.Ssecurity}
	s.token.Sids[MiioSid] = SidToken{SSecurity: pollData.Ssecurity}
	s.token.PassToken = pollData.PassToken
	s.token.UserId = fmt.Sprint(pollData.UserID)
	s.token.CUserId = pollData.CUserID

	// Step 6: Visit callback URL to get serviceToken
	callbackURL := pollData.Location
	cbReq, err := http.NewRequest(http.MethodGet, callbackURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create callback request: %w", err)
	}
	cbReq.Header = headers

	cbResp, err := s.client.Do(cbReq)
	if err != nil {
		return nil, fmt.Errorf("callback request failed: %w", err)
	}
	defer cbResp.Body.Close()

	// Extract serviceToken from cookies
	var serviceToken string
	for _, cookie := range cbResp.Cookies() {
		if cookie.Name == "serviceToken" {
			serviceToken = cookie.Value
			break
		}
	}

	if serviceToken == "" {
		// Try to get service token via securityTokenService
		serviceToken, err = s.securityTokenService(pollData.Location, pollData.Ssecurity, pollData.Nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to get service token: %w", err)
		}
	}

	// Store service token for all services
	for sid := range s.token.Sids {
		sidToken := s.token.Sids[sid]
		sidToken.ServiceToken = serviceToken
		s.token.Sids[sid] = sidToken
	}

	s.token.LoginMode = "qr"
	s.token.SSecurity = pollData.Ssecurity

	pterm.Success.Println("QR code login successful!")
	return s.token, nil
}

// getLocation calls serviceLogin to get current login state and parameters.
func (s *Service) getLocation() (map[string]string, error) {
	sid := "mijia"
	serviceLoginURL := fmt.Sprintf("https://account.xiaomi.com/pass/serviceLogin?_json=true&sid=%s", sid)

	cookies := []*http.Cookie{
		{Name: "sdkVersion", Value: "3.9"},
		{Name: "deviceId", Value: s.token.DeviceId},
	}

	if s.token.PassToken != "" {
		cookies = append(cookies, &http.Cookie{Name: "userId", Value: s.token.UserId})
		cookies = append(cookies, &http.Cookie{Name: "passToken", Value: s.token.PassToken})
	}

	req, err := http.NewRequest(http.MethodGet, serviceLoginURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create service login request: %w", err)
	}
	req.Header = http.Header{
		"User-Agent": []string{s.buildUA()},
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("service login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var serviceData loginResp
	jsonStr := strings.TrimPrefix(string(body), "&&&START&&&")
	if err := json.Unmarshal([]byte(jsonStr), &serviceData); err != nil {
		return nil, fmt.Errorf("failed to parse service login response: %w, body: %s", err, string(body))
	}

	// If code is 0 and location is valid, token is still good
	if serviceData.Code == 0 && serviceData.Location != "" {
		// Try to refresh token
		locReq, err := http.NewRequest(http.MethodGet, serviceData.Location, nil)
		if err != nil {
			return nil, err
		}
		locReq.Header = http.Header{"User-Agent": []string{s.buildUA()}}

		locResp, err := s.client.Do(locReq)
		if err != nil {
			return nil, err
		}
		defer locResp.Body.Close()

		if locResp.StatusCode == 200 {
			locBody, _ := io.ReadAll(locResp.Body)
			if string(locBody) == "ok" {
				// Token refreshed, update cookies
				for _, cookie := range locResp.Cookies() {
					if cookie.Name == "serviceToken" {
						s.token.Sids[sid] = SidToken{
							SSecurity:    serviceData.Ssecurity,
							ServiceToken: cookie.Value,
						}
					}
				}
				return map[string]string{
					"code":    "0",
					"message": "Token refresh successful",
				}, nil
			}
		}
	}

	// Return location parameters for QR login
	result := map[string]string{
		"code":           fmt.Sprintf("%d", serviceData.Code),
		"qs":             serviceData.Qs,
		"sid":            serviceData.Sid,
		"_sign":          serviceData.Sign,
		"callback":       serviceData.Callback,
		"ssecurity":      serviceData.Ssecurity,
		"nonce":          fmt.Sprintf("%d", serviceData.Nonce),
		"location":       serviceData.Location,
		"passToken":      serviceData.PassToken,
		"userId":         fmt.Sprintf("%d", serviceData.UserID),
		"cUserId":        serviceData.CUserID,
		"securityStatus": fmt.Sprintf("%d", serviceData.SecurityStatus),
	}

	return result, nil
}

// buildUA builds a User-Agent string similar to the Python implementation.
func (s *Service) buildUA() string {
	if s.token.UA != "" {
		return s.token.UA
	}

	passO := s.token.PassO
	if passO == "" {
		passO = strings.ToLower(getRandomHex(16))
		s.token.PassO = passO
	}

	uaID1 := strings.ToUpper(getRandomHex(40))
	uaID2 := strings.ToUpper(getRandomHex(32))
	uaID3 := strings.ToUpper(getRandomHex(32))
	uaID4 := strings.ToUpper(getRandomHex(40))

	ua := fmt.Sprintf("Android-15-11.0.701-Xiaomi-23046RP50C-OS2.0.212.0.VMYCNXM-%s-CN-%s-%s-SmartHome-MI_APP_STORE-%s|%s|%s-64",
		uaID1, uaID3, uaID2, uaID1, uaID4, passO)

	s.token.UA = ua
	return ua
}

// getRandomHex generates a random hex string of given length.
func getRandomHex(length int) string {
	bytes := make([]byte, length/2)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		// Fallback to pseudo-random (shouldn't happen in practice)
		return strings.Repeat("0", length)
	}
	return fmt.Sprintf("%x", bytes)
}

// qrCodeResponse represents the response from the QR code URL request.
type qrCodeResponse struct {
	Code     int    `json:"code"`
	Desc     string `json:"desc"`
	LoginURL string `json:"loginUrl"`
	QR       string `json:"qr"`
	LP       string `json:"lp"`
}

// ensureSSecurity ensures ssecurity is available for encryption.
func (s *Service) ensureSSecurity() (string, error) {
	if s.token.SSecurity != "" {
		return s.token.SSecurity, nil
	}
	if sid, ok := s.token.Sids[MinaSid]; ok && sid.SSecurity != "" {
		s.token.SSecurity = sid.SSecurity
		return sid.SSecurity, nil
	}
	return "", fmt.Errorf("ssecurity not available, please login first")
}

// genSignedNonce generates a signed nonce using ssecurity and nonce.
func genSignedNonce(ssecurity, nonce string) string {
	decodedSSec, _ := base64.StdEncoding.DecodeString(ssecurity)
	decodedNonce, _ := base64.StdEncoding.DecodeString(nonce)

	hash := sha256.New()
	hash.Write(decodedSSec)
	hash.Write(decodedNonce)
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// genEncSignature generates the signature for encrypted requests.
func genEncSignature(uri, method, signedNonce string, params url.Values) string {
	sigParams := []string{
		strings.ToUpper(method),
		uri,
	}

	// Sort keys for deterministic order
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		sigParams = append(sigParams, fmt.Sprintf("%s=%s", key, params.Get(key)))
	}

	sigParams = append(sigParams, signedNonce)
	sigString := strings.Join(sigParams, "&")

	hash := sha1.Sum([]byte(sigString))
	return base64.StdEncoding.EncodeToString(hash[:])
}
