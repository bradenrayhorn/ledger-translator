package grpc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net"
	"testing"

	"github.com/bradenrayhorn/ledger-protos/market"
	"github.com/bradenrayhorn/ledger-protos/quotes"
	"github.com/bradenrayhorn/ledger-translator/config"
	"github.com/bradenrayhorn/ledger-translator/internal/testutils"
	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/provider/tda"
	"github.com/go-redis/redis/v8"
	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type TDQuotesSuite struct {
	lis           *bufconn.Listener
	tokenDB       *redis.Client
	vaultClient   *vaultAPI.Client
	vaultListener net.Listener
	suite.Suite
}

var tdBase = "https://api.tdameritrade.com/v1/"

func TestTDQuoteSuite(t *testing.T) {
	suite.Run(t, new(TDQuotesSuite))
}

func (s *TDQuotesSuite) SetupTest() {
	s.lis = bufconn.Listen(1024 * 1024)
	sv := grpcLib.NewServer()

	tokenDB := config.NewRedisClient("oauth_tokens")
	s.tokenDB = tokenDB
	providers := []provider.Provider{tda.NewTDAProvider()}
	vaultLn, vaultClient := testutils.SetupVault(s.T())
	s.vaultListener = vaultLn
	s.vaultClient = vaultClient

	quotes.RegisterQuotesServiceServer(sv, NewQuotesServiceServer(&ProviderResolver{
		TokenDB:     tokenDB,
		Providers:   providers,
		VaultClient: vaultClient,
	}))

	go func() {
		if err := sv.Serve(s.lis); err != nil {
			log.Fatalf("Exited with error %s", err)
		}
	}()
}

func (s *TDQuotesSuite) TearDownTest() {
	s.tokenDB.FlushAll(context.Background())
	s.vaultListener.Close()
}

func (s *TDQuotesSuite) bufDialer(context.Context, string) (net.Conn, error) {
	return s.lis.Dial()
}

func (s *TDQuotesSuite) TestCanGetQuote() {
	ctx := context.Background()
	conn, err := grpcLib.DialContext(ctx, "bufnet", grpcLib.WithContextDialer(s.bufDialer), grpcLib.WithInsecure())
	defer conn.Close()
	s.Require().Nil(err)

	client := quotes.NewQuotesServiceClient(conn)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockResp, err := httpmock.NewJsonResponder(200, map[string]interface{}{
		"AAPL": map[string]interface{}{
			"askPrice":  19.25,
			"bidPrice":  20.25,
			"lastPrice": 20,
		},
	})
	s.Require().Nil(err)
	httpmock.RegisterResponder("GET", tdBase+"marketdata/AAPL/quotes", mockResp)

	token := oauth2.Token{
		AccessToken: "my-token",
	}
	tokenBytes, err := json.Marshal(token)
	s.Require().Nil(err)
	tokenString := base64.StdEncoding.EncodeToString(tokenBytes)
	vs, err := s.vaultClient.Logical().Write("transit/encrypt/ledger_translator", map[string]interface{}{
		"plaintext": tokenString,
	})
	s.Require().Nil(err)
	encryptedToken := vs.Data["ciphertext"]
	s.tokenDB.HSet(ctx, "1", "tda", encryptedToken)

	resp, err := client.GetQuote(ctx, &quotes.GetQuoteRequest{Symbol: "AAPL", MarketRequestData: &market.RequestData{
		UserID:     "1",
		ProviderID: "tda",
	}})

	s.Require().Nil(err)
	s.Require().Equal("AAPL", resp.Quote.Symbol)
	s.Require().Equal(19.25, resp.Quote.AskPrice)
	s.Require().Equal(20.25, resp.Quote.BidPrice)
	s.Require().Equal(float64(20), resp.Quote.LastPrice)
}
