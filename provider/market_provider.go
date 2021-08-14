package provider

import (
	"context"

	"github.com/bradenrayhorn/ledger-protos/quotes"
	"golang.org/x/oauth2"
)

type AuthData struct {
	UserID string
	Token  *oauth2.Token
}

type MarketProvider interface {
	GetQuote(ctx context.Context, auth AuthData, symbol string) (*quotes.Quote, error)
}
