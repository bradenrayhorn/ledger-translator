package main

import (
	"net/http"
	"net/url"

	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/dgrijalva/jwt-go"
)

type RouteController struct {
	providers  []provider.Provider
	jwtService JWTService
}

func (c RouteController) Authenticate(w http.ResponseWriter, req *http.Request) {
	jwtString := req.URL.Query().Get("jwt")
	token, err := c.jwtService.ParseToken(jwtString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid token"))
		return
	}

	userID := token.Claims.(jwt.MapClaims)["user_id"].(string)

	provider := c.getProvider(req.URL.Query())
	if provider == nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("invalid provider"))
		return
	}

	provider.Authenticate(w, req, jwtString, userID)
}

func (c RouteController) Callback(w http.ResponseWriter, req *http.Request) {
	provider := c.getProvider(req.URL.Query())

	if provider == nil {
		w.Write([]byte("invalid provider"))
		return
	}

	provider.Callback(w, req)
}

func (c RouteController) getProvider(queryValues url.Values) provider.Provider {
	providerKey := queryValues.Get("provider")
	var provider provider.Provider
	for _, p := range c.providers {
		if p.Key() == providerKey {
			provider = p
			break
		}
	}

	return provider
}
