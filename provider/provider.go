package provider

import (
	"golang.org/x/oauth2"
)

type Provider interface {
	Key() string
	Name() string
	GetOAuthConfig() *oauth2.Config
}

type ProviderArray []Provider

func (a ProviderArray) Len() int {
	return len(a)
}

func (a ProviderArray) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ProviderArray) Less(i, j int) bool {
	return a[i].Name() < a[j].Name()
}
