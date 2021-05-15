package main

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/suite"
)

type JWTTestSuite struct {
	suite.Suite
	jwtService JWTService
	validToken string
}

func (s *JWTTestSuite) SetupTest() {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	s.jwtService = JWTService{publicKey: &key.PublicKey}

	s.validToken, _ = jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user_id": "test id",
	}).SignedString(key)
}

func (s *JWTTestSuite) TestCanParseValidToken() {
	token, err := s.jwtService.ParseToken(s.validToken)

	s.Require().Nil(err)
	s.Require().Nil(token.Claims.Valid())
	s.Require().Equal("test id", token.Claims.(jwt.MapClaims)["user_id"])
}

func (s *JWTTestSuite) TestFailsToParseEmptyToken() {
	token, err := s.jwtService.ParseToken("")

	var vErr *jwt.ValidationError
	s.Require().ErrorAs(err, &vErr)
	s.Require().Equal(jwt.ValidationErrorMalformed, vErr.Errors)
	s.Require().Nil(token)
}

func (s *JWTTestSuite) TestFailsToParseWithOtherSigningMethod() {
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "test id",
	}).SignedString([]byte("x"))

	s.Require().Nil(err)
	_, err = s.jwtService.ParseToken(tokenString)

	var vErr *jwt.ValidationError
	s.Require().ErrorAs(err, &vErr)
	s.Require().Equal(jwt.ValidationErrorUnverifiable, vErr.Errors)
}

func TestJWTSuite(t *testing.T) {
	suite.Run(t, new(JWTTestSuite))
}
