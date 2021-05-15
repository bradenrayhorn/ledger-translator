package main

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/provider/tda"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite
	c     RouteController
	token string
}

func (s *ControllerTestSuite) SetupTest() {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	jwtService := JWTService{publicKey: &key.PublicKey}

	s.token, _ = jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user_id": "test id",
	}).SignedString(key)

	s.c = RouteController{
		jwtService: jwtService,
		providers:  []provider.Provider{tda.NewTDAProvider()},
	}
}

func (s *ControllerTestSuite) TestRecognizesProvider() {
	provider := s.c.getProvider(url.Values{"provider": []string{"tda"}})
	s.Require().NotNil(provider)
}

func (s *ControllerTestSuite) TestDoesNotRecognizeInvalidProvider() {
	provider := s.c.getProvider(url.Values{"provider": []string{"non"}})
	s.Require().Nil(provider)
}

func (s *ControllerTestSuite) TestCanAuthenticate() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=tda&jwt="+s.token, nil)

	handler := http.HandlerFunc(s.c.Authenticate)
	handler.ServeHTTP(w, req)

	s.Require().Equal(http.StatusFound, w.Code)
}

func (s *ControllerTestSuite) TestCannotAuthenticateWithInvalidJWT() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/authenticate?provider=tda&jwt=test", nil)
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
