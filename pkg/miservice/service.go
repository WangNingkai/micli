package miservice

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"micli/pkg/util"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Service struct {
	client     *http.Client
	username   string
	password   string
	tokenStore TokenStore
	token      *Tokens
	region     string
}

type loginResp struct {
	Qs             string      `json:"qs"`
	Ssecurity      string      `json:"ssecurity"`
	Code           int         `json:"code"`
	PassToken      string      `json:"passToken"`
	Description    string      `json:"description"`
	SecurityStatus int         `json:"securityStatus"`
	Nonce          int64       `json:"nonce"`
	UserID         int         `json:"userId"`
	CUserID        string      `json:"cUserId"`
	Result         string      `json:"result"`
	Psecurity      string      `json:"psecurity"`
	CaptchaURL     interface{} `json:"captchaUrl"`
	Location       string      `json:"location"`
	Pwd            int         `json:"pwd"`
	Child          int         `json:"child"`
	Desc           string      `json:"desc"`
	ServiceParam   string      `json:"serviceParam"`
	Sign           string      `json:"_sign"`
	Sid            string      `json:"sid"`
	Callback       string      `json:"callback"`
}

type DataCb func(tokens *Tokens, cookie map[string]string) url.Values

func New(username, password, region string, tokenStore TokenStore) *Service {
	j, _ := cookiejar.New(nil)
	return &Service{
		client: &http.Client{
			Jar: j,
		},
		username:   username,
		password:   password,
		tokenStore: tokenStore,
		region:     region,
	}
}

// Login 米家服务登录
func (s *Service) login(sid string) error {
	var err error
	defer func() {
		if err != nil && s.tokenStore != nil {
			_ = s.tokenStore.SaveToken(nil)
		}
	}()
	if s.token == nil {
		if s.tokenStore != nil {
			var tokens *Tokens
			tokens, err = s.tokenStore.LoadToken()
			if err == nil {
				if tokens.UserName != s.username {
					_ = s.tokenStore.SaveToken(nil)
				} else {
					s.token = tokens
				}
			}
		}
	}
	if s.token == nil {
		s.token = NewTokens()
		s.token.UserName = s.username
		s.token.DeviceId = strings.ToUpper(util.GetRandom(16))
	}

	cookies := []*http.Cookie{
		{Name: "sdkVersion", Value: "3.9"},
		{Name: "deviceId", Value: s.token.DeviceId},
	}

	if s.token.PassToken != "" {
		cookies = append(cookies, &http.Cookie{Name: "userId", Value: s.token.UserId})
		cookies = append(cookies, &http.Cookie{Name: "passToken", Value: s.token.PassToken})
	}

	var resp *loginResp
	resp, err = s.serviceLogin(fmt.Sprintf("serviceLogin?sid=%s&_json=true", sid), nil, cookies)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		data := url.Values{
			"_json":    {"true"},
			"qs":       {resp.Qs},
			"sid":      {resp.Sid},
			"_sign":    {resp.Sign},
			"callback": {resp.Callback},
			"user":     {s.username},
			"hash":     {strings.ToUpper(fmt.Sprintf("%x", md5.Sum([]byte(s.password))))},
		}
		resp, err = s.serviceLogin("serviceLoginAuth2", data, cookies)
		if err != nil {
			//log.Println("serviceLoginAuth2 error", err)
			return err
		}
		if resp.Code != 0 {
			return fmt.Errorf("code Error: %v", resp)
		}
	}
	s.token.UserId = fmt.Sprint(resp.UserID)
	s.token.PassToken = resp.PassToken

	var serviceToken string
	serviceToken, err = s.securityTokenService(resp.Location, resp.Ssecurity, resp.Nonce)
	if err != nil {
		//log.Println("securityTokenService error", err)
		return err
	}
	s.token.Sids[sid] = SidToken{
		SSecurity:    resp.Ssecurity,
		ServiceToken: serviceToken,
	}

	if s.tokenStore != nil {
		_ = s.tokenStore.SaveToken(s.token)
	}

	return nil
}

// Request 请求
func (s *Service) Request(sid, u string, data url.Values, cb DataCb, headers http.Header, reLogin bool, output any) error {
	if !s.existSid(sid) {
		err := s.login(sid)
		if err != nil {
			return err
		}
	}
	//log.Println("request token done")
	req := s.buildRequest(sid, u, data, cb, headers)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode == http.StatusOK {
		type _result struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		var rs []byte
		rs, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		//pterm.Println("response", u, string(rs))
		var result *_result
		err = json.Unmarshal(rs, &result)
		if err != nil {
			return err
		}
		if result.Code == 0 {
			err = json.Unmarshal(rs, output)
			return err
		}

		if strings.Contains(strings.ToLower(result.Message), "auth") {
			resp.StatusCode = http.StatusUnauthorized
		}
	}
	if resp.StatusCode == http.StatusUnauthorized && reLogin {
		s.token = nil
		if s.tokenStore != nil {
			_ = s.tokenStore.SaveToken(nil)
		}
		return s.Request(sid, u, data, cb, headers, false, output)
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("error %s: %s", u, string(body))
}

// NewRequest 构造请求
func (s *Service) buildRequest(sid, u string, data url.Values, cb DataCb, headers http.Header) *http.Request {
	var req *http.Request
	var body io.Reader
	cookies := []*http.Cookie{
		{Name: "userId", Value: s.token.UserId},
		{Name: "serviceToken", Value: s.token.Sids[sid].ServiceToken},
		{Name: "yetAnotherServiceToken", Value: s.token.Sids[sid].ServiceToken},
		{Name: "channel", Value: "MI_APP_STORE"},
	}
	//pterm.Println("tokens", s.token)
	method := http.MethodGet
	if data != nil || cb != nil {
		var values url.Values
		if cb != nil {
			var cookieMap = make(map[string]string)
			values = cb(s.token, cookieMap)
			for k, v := range cookieMap {
				cookies = append(cookies, &http.Cookie{Name: k, Value: v})
			}
		} else if data != nil {
			values = data
		}
		if values != nil {
			method = http.MethodPost
			//pterm.Println(values)
			body = strings.NewReader(values.Encode())
			headers.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	req, _ = http.NewRequest(method, u, body)
	if headers != nil {
		req.Header = headers
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	/*for k, v := range cookies {
		pterm.Println("request cookie", k, v)
	}*/
	return req
}

// serviceLogin 服务登录
func (s *Service) serviceLogin(uri string, data url.Values, cookies []*http.Cookie) (*loginResp, error) {
	headers := http.Header{
		"User-Agent": []string{"APP/com.xiaomi.mihome APPV/6.0.103 iosPassportSDK/3.9.0 iOS/14.4 miHSTS"},
	}
	var reqBody io.Reader
	method := http.MethodGet
	if data != nil {
		reqBody = strings.NewReader(data.Encode())
		method = http.MethodPost
		headers.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req, _ := http.NewRequest(method, fmt.Sprintf("https://account.xiaomi.com/pass/%s", uri), reqBody)
	req.Header = headers

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	//log.Println("service login", req.URL.String())
	resp, err := s.client.Do(req)
	if err != nil {
		//log.Println("http do request error", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	//log.Println("service login return", resp.StatusCode)
	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//log.Println("body", string(body))
	var jsonResponse loginResp
	err = json.Unmarshal(body[11:], &jsonResponse)
	if err != nil {
		//log.Println("json unmarshal error", err, string(body))
		return nil, err
	}
	//log.Println("service login success", jsonResponse)
	return &jsonResponse, nil
}

// secureUrl 生成安全链接
func (s *Service) secureUrl(location, sSecurity string, nonce int64) string {
	sNonce := fmt.Sprintf("nonce=%d&%s", nonce, sSecurity)
	sum := sha1.Sum([]byte(sNonce))
	clientSign := base64.StdEncoding.EncodeToString(sum[:])
	es := url.QueryEscape(clientSign)
	//es = strings.ReplaceAll(es, "%2F", "/")
	requestUrl := fmt.Sprintf("%s&clientSign=%s", location, es)
	return requestUrl
}

// securityTokenService 获取安全令牌
func (s *Service) securityTokenService(location, sSecurity string, nonce int64) (string, error) {
	requestUrl := s.secureUrl(location, sSecurity, nonce)
	//log.Println("securityTokenService", requestUrl)
	req, _ := http.NewRequest(http.MethodGet, requestUrl, nil)
	headers := http.Header{
		"User-Agent": []string{"APP/com.xiaomi.mihome APPV/6.0.103 iosPassportSDK/3.9.0 iOS/14.4 miHSTS"},
	}
	req.Header = headers

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	cookies := resp.Cookies()
	var serviceToken string

	for _, cookie := range cookies {
		if cookie.Name == "serviceToken" {
			serviceToken = cookie.Value
			break
		}
	}

	if serviceToken == "" {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.New(string(body))
	}

	return serviceToken, nil
}

// existSid 判断是否有sid
func (s *Service) existSid(sid string) bool {
	if s.token == nil {
		return false
	}
	_, ok := s.token.Sids[sid]
	return ok
}
