package main

import (
	"net/http"

	"github.com/bradenrayhorn/ledger-translator/provider"
)

type RouteController struct {
	providers []provider.Provider
}

func (c RouteController) Authenticate(w http.ResponseWriter, req *http.Request) {
	provider := c.getProvider(req)
	if provider == nil {
		w.Write([]byte("invalid provider"))
		return
	}

	provider.Authenticate(w, req)
}

func (c RouteController) Callback(w http.ResponseWriter, req *http.Request) {
	provider := c.getProvider(req)

	if provider == nil {
		w.Write([]byte("invalid provider"))
	}

	provider.Callback(w, req)
}

func (c RouteController) getProvider(req *http.Request) provider.Provider {
	providerKey := req.URL.Query().Get("provider")
	var provider provider.Provider
	for _, p := range c.providers {
		if p.Key() == providerKey {
			provider = p
			break
		}
	}

	return provider
}
