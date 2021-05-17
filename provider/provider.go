package provider

import (
	"golang.org/x/oauth2"
)

type Provider interface {
	Key() string
	GetOAuthConfig() *oauth2.Config
}
