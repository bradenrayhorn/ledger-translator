package grpc

import (
	"context"

	"github.com/bradenrayhorn/ledger-protos/quotes"
	"github.com/bradenrayhorn/ledger-translator/provider"
)

type QuotesServiceServer struct {
	quotes.UnimplementedQuotesServiceServer
	resolver *ProviderResolver
}

func NewQuotesServiceServer(resolver *ProviderResolver) QuotesServiceServer {
	return QuotesServiceServer{
		resolver: resolver,
	}
}

func (s QuotesServiceServer) GetQuote(ctx context.Context, req *quotes.GetQuoteRequest) (*quotes.GetQuoteResponse, error) {
	response := &quotes.GetQuoteResponse{}

	ledgerProvider, oauthToken, err := s.resolver.Resolve(ctx, req.MarketRequestData)
	if err != nil {
		return response, err
	}
	quote, err := ledgerProvider.(provider.MarketProvider).GetQuote(ctx, provider.AuthData{UserID: req.MarketRequestData.UserID, Token: oauthToken}, req.Symbol)
	if err != nil {
		return response, err
	}
	response.Quote = quote

	return response, nil
}
