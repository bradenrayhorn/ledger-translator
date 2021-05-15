package tda

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
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

type oauthState struct {
	Random   string
	JWT      string
	Provider string
}

func (t tdaProvider) Key() string {
	return "tda"
}

func (t tdaProvider) Authenticate(w http.ResponseWriter, req *http.Request, jwtString string, userID string) {
	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to generate state: %e", err)
		return
	}

	state := oauthState{
		Random:   string(b),
		JWT:      jwtString,
		Provider: t.Key(),
	}
	data, err := json.Marshal(state)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to marshal state: %e", err)
		return
	}

	url := t.oauth.AuthCodeURL(base64.RawURLEncoding.EncodeToString(data))

	http.Redirect(w, req, url, http.StatusFound)
}

func (t tdaProvider) Callback(w http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	state := req.URL.Query().Get("state")

	if len(state) == 0 || state != "bloop" {
		w.Write([]byte("invalid state"))
		return
	}

	if len(code) == 0 {
		w.Write([]byte("invalid code"))
		return
	}
}
