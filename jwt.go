package main

import (
	"crypto/rsa"
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

type JWTService struct {
	publicKey *rsa.PublicKey
}

func (j JWTService) ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})
}
