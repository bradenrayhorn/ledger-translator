package main

import (
	"log"

	"github.com/bradenrayhorn/ledger-protos/session"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type GRPCClient struct {
	session session.SessionAuthenticatorClient
}

func NewGRPCClient() GRPCClient {
	conn, err := grpc.Dial(viper.GetString("grpc_host_auth"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to grpc auth: %v", err)
	}

	sessionClient := session.NewSessionAuthenticatorClient(conn)
	return GRPCClient{
		session: sessionClient,
	}
}
