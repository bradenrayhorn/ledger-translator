package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/bradenrayhorn/ledger-protos/session"
	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/service"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var tokenURL = ""

type ControllerTestSuite struct {
	suite.Suite
	c         RouteController
	sessionID string
	conn      *grpc.ClientConn
	sessionDB *redis.Client
	tokenDB   *redis.Client
}

type TestProvider struct {
}

func (p TestProvider) Key() string {
	return "test"
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

func (s *ControllerTestSuite) SetupTest() {
	loadConfig()
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

	s.sessionID = "good-id"
	s.sessionDB = NewRedisClient("oauth_sessions")
	s.tokenDB = NewRedisClient("oauth_tokens")
	s.c = RouteController{
		providers:      []provider.Provider{TestProvider{}},
		sessionDB:      s.sessionDB,
		tokenDB:        s.tokenDB,
		sessionService: service.NewSessionService(client),
	}
}

func (s *ControllerTestSuite) TearDownTest() {
	s.conn.Close()
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

	handler := http.HandlerFunc(s.c.Authenticate)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusFound, w.Code)
	s.Require().Regexp(regexp.MustCompile("^https://example\\.com/auth.+$"), w.HeaderMap.Get("location"))
}

func (s *ControllerTestSuite) TestCannotAuthenticateWithNoSession() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=test", nil)
	handler := http.HandlerFunc(s.c.Authenticate)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestCannotAuthenticateWithInvalidSession() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=test", nil)
	req.Header.Add("Cookie", "session_id=x")
	handler := http.HandlerFunc(s.c.Authenticate)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestCannotAuthenticateWithInvalidProvider() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=non", nil)
	req.Header.Add("Cookie", "session_id="+s.sessionID)
	handler := http.HandlerFunc(s.c.Authenticate)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnprocessableEntity, w.Code)
}

// Callback Tests
func (s *ControllerTestSuite) TestCannotCallbackWithNoSession() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/callback?provider=test", nil)
	handler := http.HandlerFunc(s.c.Callback)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestCannotCallbackWithInvalidSession() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/callback?provider=test", nil)
	req.Header.Add("Cookie", "session_id=x")
	handler := http.HandlerFunc(s.c.Callback)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestCannotCallbackWithInvalidProvider() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/callback?provider=non", nil)
	req.Header.Add("Cookie", "session_id="+s.sessionID)
	handler := http.HandlerFunc(s.c.Callback)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestCanCallback() {
	stateString, _ := json.Marshal(service.OAuthState{Random: "x", Provider: "test"})
	encodedState := base64.RawURLEncoding.EncodeToString(stateString)
	s.sessionDB.Set(context.Background(), s.sessionID, encodedState, 0)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://callback?code=newcode&state="+encodedState, nil)
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
	handler := http.HandlerFunc(s.c.Callback)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)
	tokenKey, _ := json.Marshal(service.TokenKey{UserID: "good-user-id", Provider: "test"})
	savedToken := s.tokenDB.Get(context.Background(), base64.RawURLEncoding.EncodeToString(tokenKey))
	s.Require().Nil(savedToken.Err())
	var token oauth2.Token
	savedTokenString, _ := base64.RawStdEncoding.DecodeString(savedToken.Val())
	s.Require().Nil(json.Unmarshal(savedTokenString, &token))
	s.Require().Equal("access-token", token.AccessToken)
	s.Require().Equal("refresh-token", token.RefreshToken)
	s.Require().Equal("ok", token.TokenType)
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
