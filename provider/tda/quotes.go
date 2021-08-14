package tda

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bradenrayhorn/ledger-protos/quotes"
	"github.com/bradenrayhorn/ledger-translator/provider"
)

type TDQuote struct {
	AskPrice  float64 `json:"askPrice"`
	BidPrice  float64 `json:"bidPrice"`
	LastPrice float64 `json:"lastPrice"`
}

func (t tdaProvider) GetQuote(ctx context.Context, auth provider.AuthData, symbol string) (*quotes.Quote, error) {
	symbol = strings.ToUpper(symbol)
	quote := quotes.Quote{}
	tokenSrc := t.GetOAuthConfig().TokenSource(ctx, auth.Token)
	req, err := t.client.CreateRequest(tokenSrc, "GET", fmt.Sprintf("marketdata/%s/quotes", symbol), nil)
	if err != nil {
		return nil, err
	}

	var resp map[string]json.RawMessage
	err = t.client.Do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	var tdQuote TDQuote
	err = json.Unmarshal(resp[symbol], &tdQuote)
	if err != nil {
		return nil, err
	}

	quote.Symbol = symbol
	quote.AskPrice = tdQuote.AskPrice
	quote.BidPrice = tdQuote.BidPrice
	quote.LastPrice = tdQuote.LastPrice

	return &quote, err
}
