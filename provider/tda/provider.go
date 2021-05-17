package tda

import (
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type tdaProvider struct {
	oauth oauth2.Config
}

func NewTDAProvider() tdaProvider {
	return tdaProvider{
		oauth: oauth2.Config{
			ClientID: viper.GetString("client_id") + "@AMER.OAUTHAP",
			Endpoint: oauth2.Endpoint{
				TokenURL: "https://api.tdameritrade.com/v1/oauth2/token",
				AuthURL:  "https://auth.tdameritrade.com/auth",
			},
			RedirectURL: "https://localhost:8080/callback",
		},
	}
}

type oauthState struct {
	Random   string
	JWT      string
	Provider string
}

func (t tdaProvider) Key() string {
	return "tda"
}

func (t tdaProvider) GetOAuthConfig() *oauth2.Config {
	return &t.oauth
}
