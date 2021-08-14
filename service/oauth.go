package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type OAuth struct {
	oauthConfig oauth2.Config
	tokenDB     *redis.Client
	vaultClient *vaultAPI.Client
}

type OAuthState struct {
	Random   string
	Provider string
}

type TokenKey struct {
	UserID   string
	Provider string
}

func NewOAuthService(config oauth2.Config, tokenDB *redis.Client, vaultClient *vaultAPI.Client) OAuth {
	return OAuth{
		oauthConfig: config,
		tokenDB:     tokenDB,
		vaultClient: vaultClient,
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
	token, err := o.oauthConfig.Exchange(context.Background(), code, oauth2.AccessTypeOffline)
	if err != nil {
		return err
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return err
	}
	tokenString := base64.StdEncoding.EncodeToString(tokenBytes)
	s, err := o.vaultClient.Logical().Write(getVaultPath("encrypt"), map[string]interface{}{
		"plaintext": tokenString,
	})
	if err != nil {
		return err
	}
	encryptedToken := s.Data["ciphertext"]

	return o.tokenDB.HSet(context.Background(), userID, provider, encryptedToken).Err()
}

func getVaultPath(action string) string {
	return fmt.Sprintf("%s/%s/%s", viper.GetString("vault_transit_path"), action, viper.GetString("vault_transit_key"))
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

func DecryptOAuthToken(vaultClient *vaultAPI.Client, encryptedToken string) (*oauth2.Token, error) {
	s, err := vaultClient.Logical().Write(getVaultPath("decrypt"), map[string]interface{}{
		"ciphertext": encryptedToken,
	})
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	tokenString, err := base64.StdEncoding.DecodeString(s.Data["plaintext"].(string))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(tokenString, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}
