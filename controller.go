package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/service"
	"github.com/go-redis/redis/v8"
	vaultAPI "github.com/hashicorp/vault/api"
)

type RouteController struct {
	providers      []provider.Provider
	sessionService service.Session
	sessionDB      *redis.Client
	tokenDB        *redis.Client
	vaultClient    *vaultAPI.Client
}

func (c RouteController) Authenticate(w http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("session_id")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	sessionID := cookie.Value
	_, err = c.sessionService.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	provider := c.getProvider(req.URL.Query().Get("provider"))
	if provider == nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid provider")
		return
	}

	oauthService := service.NewOAuthService(*provider.GetOAuthConfig(), c.tokenDB, c.vaultClient)
	url, state, err := oauthService.Authenticate(provider.Key())

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start oauth")
		return
	}

	_, err = c.sessionDB.Set(context.Background(), sessionID, state, time.Minute*15).Result()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save state")
		return
	}

	http.Redirect(w, req, url, http.StatusFound)
}

func (c RouteController) Callback(w http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("session_id")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	sessionID := cookie.Value
	userID, err := c.sessionService.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	state, err := service.OAuthDecodeState(req.URL.Query().Get("state"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid state")
		return
	}

	savedStateString, err := c.sessionDB.Get(context.Background(), sessionID).Result()
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid state")
		return
	}
	err = c.sessionDB.Del(context.Background(), sessionID).Err()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove state")
		return
	}
	savedState, err := service.OAuthDecodeState(savedStateString)
	if err != nil || state.Random != savedState.Random || state.Provider != savedState.Provider {
		writeError(w, http.StatusUnauthorized, "invalid state")
		return
	}

	provider := c.getProvider(state.Provider)
	if provider == nil {
		writeError(w, http.StatusUnauthorized, "invalid state")
		return
	}

	oauthService := service.NewOAuthService(*provider.GetOAuthConfig(), c.tokenDB, c.vaultClient)
	err = oauthService.SaveToken(userID, state.Provider, req.URL.Query().Get("code"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save state")
		return
	}
	w.Write([]byte("yay"))
}

func (c RouteController) GetProviders(w http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("session_id")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	sessionID := cookie.Value
	userID, err := c.sessionService.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	res := c.tokenDB.HGetAll(context.Background(), userID)
	if res.Err() != nil {
		log.Printf("failed to get providers: %s", res.Err())
		writeError(w, http.StatusInternalServerError, "failed to get providers")
		return
	}

	userProviders := res.Val()

	var providers []map[string]interface{}
	sortedProviders := c.providers
	sort.Sort(provider.ProviderArray(sortedProviders))
	for _, provider := range c.providers {
		isAuthenticated := false
		if _, ok := userProviders[provider.Key()]; ok {
			isAuthenticated = true
		}
		providers = append(providers, map[string]interface{}{
			"id":            provider.Key(),
			"name":          provider.Name(),
			"authenticated": isAuthenticated,
		})
	}

	b, err := json.Marshal(providers)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to format output")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

func (c RouteController) getProvider(providerKey string) provider.Provider {
	var provider provider.Provider
	for _, p := range c.providers {
		if p.Key() == providerKey {
			provider = p
			break
		}
	}

	return provider
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
