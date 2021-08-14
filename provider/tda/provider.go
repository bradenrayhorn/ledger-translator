package tda

import (
	"net/url"

	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type tdaProvider struct {
	oauth  oauth2.Config
	client *provider.Client
}

func NewTDAProvider() tdaProvider {
	baseURL, _ := url.Parse("https://api.tdameritrade.com/v1/")
	return tdaProvider{
		client: provider.NewClient(baseURL),
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

func (t tdaProvider) Key() string {
	return "tda"
}

func (t tdaProvider) Name() string {
	return "TD Ameritrade"
}

func (t tdaProvider) Types() []provider.Type {
	return []provider.Type{provider.MarketType}
}

func (t tdaProvider) GetOAuthConfig() *oauth2.Config {
	return &t.oauth
}
