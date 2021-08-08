package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/bradenrayhorn/ledger-protos/session"
	"github.com/bradenrayhorn/ledger-translator/config"
	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-hclog"
	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/builtin/logical/transit"
	vaultHTTP "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var tokenURL = ""

type ControllerTestSuite struct {
	suite.Suite
	c             RouteController
	sessionID     string
	conn          *grpc.ClientConn
	sessionDB     *redis.Client
	tokenDB       *redis.Client
	vaultListener net.Listener
	vaultClient   *vaultAPI.Client
	router        *chi.Mux
}

type TestProvider struct {
}

func (p TestProvider) Key() string {
	return "test"
}

func (p TestProvider) Name() string {
	return "Test Provider"
}

func (p TestProvider) GetOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID: "my client id",
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenURL,
			AuthURL:  "https://example.com/auth",
		},
		RedirectURL: "https://example.com/callback",
	}
}

type TestProviderOther struct{}

func (p TestProviderOther) Key() string {
	return "test_other"
}
func (p TestProviderOther) Name() string {
	return "Other Test Provider"
}

func (p TestProviderOther) GetOAuthConfig() *oauth2.Config {
	return nil
}

func (s *ControllerTestSuite) SetupTest() {
	config.LoadConfig()
	lis := bufconn.Listen(1024 * 1024)
	sv := grpc.NewServer()
	session.RegisterSessionAuthenticatorServer(sv, SessionAuthenticatorServer{})
	go func() {
		if err := sv.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithInsecure(),
	)
	s.Require().Nil(err)
	s.conn = conn

	client := session.NewSessionAuthenticatorClient(conn)

	// configure vault
	vaultLogger := hclog.New(&hclog.LoggerOptions{
		Output: io.Discard,
	})
	coreConfig := &vault.CoreConfig{
		LogicalBackends: map[string]logical.Factory{
			"transit": transit.Factory,
		},
		Logger: vaultLogger,
	}
	core, _, rootToken := vault.TestCoreUnsealedWithConfig(s.T(), coreConfig)
	ln, addr := vaultHTTP.TestServer(s.T(), core)
	conf := vaultAPI.DefaultConfig()
	conf.Address = addr
	vaultClient, err := vaultAPI.NewClient(conf)
	s.Require().Nil(err)
	vaultClient.SetToken(rootToken)
	err = vaultClient.Sys().Mount("transit", &vaultAPI.MountInput{
		Type: "transit",
	})
	s.Require().Nil(err)
	_, err = vaultClient.Logical().Write("transit/keys/my-key", map[string]interface{}{})
	s.Require().Nil(err)

	s.vaultListener = ln
	s.sessionID = "good-id"
	s.sessionDB = config.NewRedisClient("oauth_sessions")
	s.tokenDB = config.NewRedisClient("oauth_tokens")
	s.vaultClient = vaultClient
	s.c = RouteController{
		providers:      []provider.Provider{TestProvider{}, TestProviderOther{}},
		sessionDB:      s.sessionDB,
		tokenDB:        s.tokenDB,
		sessionService: service.NewSessionService(client),
		vaultClient:    vaultClient,
	}
	s.router = CreateRouter(s.c)
}

func (s *ControllerTestSuite) TearDownTest() {
	s.conn.Close()
	s.vaultListener.Close()
}

func (s *ControllerTestSuite) TestRecognizesProvider() {
	provider := s.c.getProvider("test")
	s.Require().NotNil(provider)
}

func (s *ControllerTestSuite) TestDoesNotRecognizeInvalidProvider() {
	provider := s.c.getProvider("non")
	s.Require().Nil(provider)
}

func (s *ControllerTestSuite) TestCanAuthenticate() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=test", nil)
	req.Header.Add("Cookie", "session_id="+s.sessionID)

	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusFound, w.Code)
	s.Require().Regexp(regexp.MustCompile("^https://example\\.com/auth.+$"), w.HeaderMap.Get("location"))
}

func (s *ControllerTestSuite) TestCannotAuthenticateWithNoSession() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=test", nil)
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestCannotAuthenticateWithInvalidSession() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=test", nil)
	req.Header.Add("Cookie", "session_id=x")
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestCannotAuthenticateWithInvalidProvider() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=non", nil)
	req.Header.Add("Cookie", "session_id="+s.sessionID)
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnprocessableEntity, w.Code)
}

// Callback Tests
func (s *ControllerTestSuite) TestCannotCallbackWithNoSession() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/callback?provider=test", nil)
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestCannotCallbackWithInvalidSession() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/callback?provider=test", nil)
	req.Header.Add("Cookie", "session_id=x")
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestCannotCallbackWithInvalidProvider() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/callback?provider=non", nil)
	req.Header.Add("Cookie", "session_id="+s.sessionID)
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestCanCallback() {
	stateString, _ := json.Marshal(service.OAuthState{Random: "x", Provider: "test"})
	encodedState := base64.RawURLEncoding.EncodeToString(stateString)
	s.sessionDB.Set(context.Background(), s.sessionID, encodedState, 0)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/callback?code=newcode&state="+encodedState, nil)
	req.Header.Add("Cookie", "session_id="+s.sessionID)

	sv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.String() == "/token" {
			MockOAuthExchange(w, req)
		} else {
			w.WriteHeader(http.StatusGone)
		}
	}))
	sv.Start()
	tokenURL = sv.URL + "/token"
	defer sv.Close()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)
	savedToken := s.tokenDB.HGet(context.Background(), "good-user-id", "test")
	s.Require().Nil(savedToken.Err())
	var token oauth2.Token
	decrypted, err := s.vaultClient.Logical().Write("transit/decrypt/my-key", map[string]interface{}{
		"ciphertext": savedToken.Val(),
	})
	s.Require().Nil(err)
	savedTokenString, _ := base64.StdEncoding.DecodeString(decrypted.Data["plaintext"].(string))
	s.Require().Nil(json.Unmarshal(savedTokenString, &token))
	s.Require().Equal("access-token", token.AccessToken)
	s.Require().Equal("refresh-token", token.RefreshToken)
	s.Require().Equal("ok", token.TokenType)
}

// Get Provider Tests
func (s *ControllerTestSuite) TestCannotGetProvidersUnauthorized() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/providers", nil)
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

type GetProvidersResponse struct {
	Providers []GetProvidersResponseProvider `json:"data"`
}

type GetProvidersResponseProvider struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Authenticated bool   `json:"authenticated"`
}

func (s *ControllerTestSuite) TestCanGetProviders() {
	s.tokenDB.HSet(context.Background(), "good-user-id", "test", "some key here")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/providers", nil)
	req.Header.Add("Cookie", "session_id="+s.sessionID)
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)
	var body GetProvidersResponse
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	s.Require().Len(body.Providers, 2)
	s.Require().Equal(body.Providers[0].Name, "Other Test Provider")
	s.Require().Equal(body.Providers[1].Name, "Test Provider")
	s.Require().False(body.Providers[0].Authenticated)
	s.Require().True(body.Providers[1].Authenticated)
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

// mock session grpc service
type SessionAuthenticatorServer struct {
	session.UnimplementedSessionAuthenticatorServer
}

func (s SessionAuthenticatorServer) Authenticate(ctx context.Context, req *session.SessionAuthenticateRequest) (*session.SessionAuthenticateResponse, error) {
	response := &session.SessionAuthenticateResponse{}

	if req.SessionID != "good-id" {
		return response, errors.New("invalid session ID")
	}

	response.Session = &session.Session{
		SessionID: "good-id",
		UserID:    "good-user-id",
	}

	return response, nil
}

// mock oauth token exchange
func MockOAuthExchange(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":             "access-token",
		"refresh_token":            "refresh-token",
		"token_type":               "ok",
		"expires_in":               300,
		"scope":                    "all",
		"refresh_token_expires_in": 600,
	})
}
