package provider

import (
	"golang.org/x/oauth2"
)

// enum representing types of providers

type Type int

const (
	UserType Type = iota
	MarketType
)

// a generic provider

type Provider interface {
	Key() string
	Name() string
	Types() []Type
	GetOAuthConfig() *oauth2.Config
}

// provider array type with the ability to sort

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
