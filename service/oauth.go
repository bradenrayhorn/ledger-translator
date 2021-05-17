package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"golang.org/x/oauth2"
)

type OAuth struct {
	oauthConfig oauth2.Config
	tokenDB     *redis.Client
}

type OAuthState struct {
	Random   string
	Provider string
}

type TokenKey struct {
	UserID   string
	Provider string
}

func NewOAuthService(config oauth2.Config, tokenDB *redis.Client) OAuth {
	return OAuth{
		oauthConfig: config,
		tokenDB:     tokenDB,
	}
}

func (o OAuth) Authenticate(provider string) (string, string, error) {
	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}

	state := OAuthState{
		Random:   string(b),
		Provider: provider,
	}
	data, err := json.Marshal(state)
	if err != nil {
		return "", "", err
	}

	encodedState := base64.RawURLEncoding.EncodeToString(data)
	url := o.oauthConfig.AuthCodeURL(encodedState)
	return url, encodedState, nil
}

func (o OAuth) SaveToken(userID string, provider string, code string) error {
	token, err := o.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	tokenString, err := json.Marshal(token)
	if err != nil {
		return err
	}
	tokenKey, err := json.Marshal(TokenKey{UserID: userID, Provider: provider})
	if err != nil {
		return err
	}
	tokenKeyString := base64.RawURLEncoding.EncodeToString(tokenKey)

	_, err = o.tokenDB.Set(context.Background(), tokenKeyString, base64.RawURLEncoding.EncodeToString(tokenString), 0).Result()
	return err
}

func OAuthDecodeState(stateString string) (*OAuthState, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(stateString)
	if err != nil {
		return nil, err
	}

	var state OAuthState
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}
