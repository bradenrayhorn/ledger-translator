package grpc

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/bradenrayhorn/ledger-protos/provider"
	"github.com/bradenrayhorn/ledger-translator/config"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type ProviderSuite struct {
	lis     *bufconn.Listener
	tokenDB *redis.Client
	suite.Suite
}

func TestProviderSuite(t *testing.T) {
	suite.Run(t, new(ProviderSuite))
}

func (s *ProviderSuite) SetupTest() {
	s.lis = bufconn.Listen(1024 * 1024)
	sv := grpc.NewServer()
	s.tokenDB = config.NewRedisClient("oauth_tokens")
	provider.RegisterProviderServiceServer(sv, NewProviderServiceServer(s.tokenDB))

	go func() {
		if err := sv.Serve(s.lis); err != nil {
			log.Fatalf("server exited with error: %s", err)
		}
	}()
}

func (s *ProviderSuite) bufDialer(context.Context, string) (net.Conn, error) {
	return s.lis.Dial()
}

func (s *ProviderSuite) TearDownTest() {
	s.tokenDB.FlushAll(context.Background())
}

func (s *ProviderSuite) TestGetProviderGetsOnlyAuthenticatedProviders() {
	ctx := context.Background()
	s.tokenDB.HSet(ctx, "user-id", "tda", "my-value")
	s.tokenDB.HSet(ctx, "user-id", "other-provider", "my-value")
	s.tokenDB.HSet(ctx, "user-id-two", "another-provider", "my-value")

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(s.bufDialer), grpc.WithInsecure())
	s.Require().Nil(err)
	defer conn.Close()

	client := provider.NewProviderServiceClient(conn)
	resp, err := client.GetForUser(ctx, &provider.GetUserProvidersRequest{UserID: "user-id"})
	s.Require().Nil(err)
	providerIDs := resp.GetProviderIDs()
	s.Require().Len(providerIDs, 2)
	s.Require().Contains(providerIDs, "tda")
	s.Require().Contains(providerIDs, "other-provider")
}
