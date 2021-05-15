package main

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite
	c     RouteController
	token string
}

type TestProvider struct {
}

func (p TestProvider) Key() string {
	return "test"
}

func (p TestProvider) Authenticate(w http.ResponseWriter, req *http.Request, jwtString string, userID string) {
	http.Redirect(w, req, "http://url.example", http.StatusFound)
}

func (p TestProvider) Callback(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "http://ok.ok", http.StatusFound)
}

func (s *ControllerTestSuite) SetupTest() {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	jwtService := JWTService{publicKey: &key.PublicKey}

	s.token, _ = jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user_id": "test id",
	}).SignedString(key)

	s.c = RouteController{
		jwtService: jwtService,
		providers:  []provider.Provider{TestProvider{}},
	}
}

func (s *ControllerTestSuite) TestRecognizesProvider() {
	provider := s.c.getProvider(url.Values{"provider": []string{"test"}})
	s.Require().NotNil(provider)
}

func (s *ControllerTestSuite) TestDoesNotRecognizeInvalidProvider() {
	provider := s.c.getProvider(url.Values{"provider": []string{"non"}})
	s.Require().Nil(provider)
}

func (s *ControllerTestSuite) TestCanAuthenticate() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=test&jwt="+s.token, nil)

	handler := http.HandlerFunc(s.c.Authenticate)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusFound, w.Code)
	s.Require().Equal("http://url.example", w.HeaderMap.Get("location"))
}

func (s *ControllerTestSuite) TestCannotAuthenticateWithInvalidJWT() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=test&jwt=test", nil)
	handler := http.HandlerFunc(s.c.Authenticate)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestCannotAuthenticateWithInvalidProvider() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=non&jwt="+s.token, nil)
	handler := http.HandlerFunc(s.c.Authenticate)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnprocessableEntity, w.Code)
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
