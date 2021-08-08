package grpc

import (
	"context"

	pbProvider "github.com/bradenrayhorn/ledger-protos/provider"
	"github.com/go-redis/redis/v8"
)

type ProviderServiceServer struct {
	tokenRedisClient *redis.Client
	pbProvider.UnimplementedProviderServiceServer
}

func NewProviderServiceServer(tokenRedisClient *redis.Client) ProviderServiceServer {
	return ProviderServiceServer{
		tokenRedisClient: tokenRedisClient,
	}
}

func (s ProviderServiceServer) GetUserProviders(ctx context.Context, req *pbProvider.GetUserProvidersRequest) (*pbProvider.GetUserProvidersResponse, error) {
	response := &pbProvider.GetUserProvidersResponse{}
	res := s.tokenRedisClient.HGetAll(ctx, req.UserID)
	if res.Err() != nil {
		return response, res.Err()
	}

	response.ProviderIDs = make([]string, len(res.Val()))
	i := 0
	for k := range res.Val() {
		response.ProviderIDs[i] = k
		i++
	}

	return response, nil
}
