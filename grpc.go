package main

import (
	"crypto/tls"
	"log"

	"github.com/bradenrayhorn/ledger-protos/session"
	"github.com/bradenrayhorn/ledger-translator/config"
	"github.com/johanbrandhorst/certify"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GRPCClient struct {
	session session.SessionAuthenticatorClient
}

func NewGRPCClient(certify *certify.Certify) GRPCClient {
	tlsConfig := &tls.Config{
		GetClientCertificate: certify.GetClientCertificate,
		RootCAs:              config.GetCACertPool(),
	}
	conn, err := grpc.Dial(viper.GetString("grpc_host_auth"), grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		log.Fatalf("could not connect to grpc auth: %v", err)
	}

	sessionClient := session.NewSessionAuthenticatorClient(conn)
	return GRPCClient{
		session: sessionClient,
	}
}
