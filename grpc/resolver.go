package grpc

import (
	"context"

	"github.com/bradenrayhorn/ledger-protos/market"
	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/service"
	"github.com/go-redis/redis/v8"
	vaultAPI "github.com/hashicorp/vault/api"
	"golang.org/x/oauth2"
)

type ProviderResolver struct {
	TokenDB     *redis.Client
	VaultClient *vaultAPI.Client
	Providers   []provider.Provider
}

func (r ProviderResolver) Resolve(ctx context.Context, marketRequestData *market.RequestData) (provider.Provider, *oauth2.Token, error) {
	var foundProvider provider.Provider
	for _, p := range r.Providers {
		if p.Key() != marketRequestData.ProviderID {
			continue
		}

		for _, t := range p.Types() {
			if t == provider.MarketType {
				foundProvider = p
				break
			}
		}
	}

	if foundProvider == nil {
		return nil, nil, nil
	}

	res := r.TokenDB.HGet(ctx, marketRequestData.UserID, marketRequestData.ProviderID)
	if res.Err() != nil {
		return nil, nil, res.Err()
	}

	token, err := service.DecryptOAuthToken(r.VaultClient, res.Val())
	if err != nil {
		return nil, nil, err
	}

	return foundProvider, token, nil
}
