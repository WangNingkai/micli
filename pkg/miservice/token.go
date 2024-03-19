package miservice

import (
	"os"
)

type SidToken struct {
	SSecurity    string `json:"ssecurity"`
	ServiceToken string `json:"service_token"`
}

type Tokens struct {
	UserName  string              `json:"user_name"`
	DeviceId  string              `json:"device_id"`
	UserId    string              `json:"user_id"`
	PassToken string              `json:"pass_token"`
	Sids      map[string]SidToken `json:"sids"`
}

func NewTokens() *Tokens {
	return &Tokens{
		Sids: make(map[string]SidToken),
	}
}

type TokenStore interface {
	LoadToken() (*Tokens, error)
	SaveToken(tokens *Tokens) error
}

type FileTokenStore struct {
	tokenPath string
}

func NewTokenStore(tokenPath string) *FileTokenStore {
	return &FileTokenStore{tokenPath: tokenPath}
}

func (mts *FileTokenStore) LoadToken() (*Tokens, error) {
	var tokens Tokens
	if _, err := os.Stat(mts.tokenPath); os.IsNotExist(err) {
		return nil, err
	}
	data, err := os.ReadFile(mts.tokenPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &tokens)
	return &tokens, err
}

func (mts *FileTokenStore) SaveToken(tokens *Tokens) (err error) {
	if tokens != nil {
		var data []byte
		data, err = json.MarshalIndent(tokens, "", "  ")
		if err != nil {
			return
		}
		err = os.WriteFile(mts.tokenPath, data, 0o644)
		if err != nil {
			return
		}
	} else {
		err = os.Remove(mts.tokenPath)
		if os.IsNotExist(err) {
			err = nil
		}
	}
	return
}
