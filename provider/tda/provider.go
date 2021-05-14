package tda

import (
	"net/http"

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

func (t tdaProvider) Key() string {
	return "tda"
}

func (t tdaProvider) Authenticate(w http.ResponseWriter, req *http.Request) {
	url := t.oauth.AuthCodeURL("bloop")

	http.Redirect(w, req, url, http.StatusFound)
}

func (t tdaProvider) Callback(w http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	state := req.URL.Query().Get("state")

	if state != "bloop" {
		w.Write([]byte("invalid state"))
		return
	}

	if len(code) == 0 {
		w.Write([]byte("invalid code"))
		return
	}
}
