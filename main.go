package main

import (
	"log"
	"net/http"

	"github.com/bradenrayhorn/ledger-translator/config"
	"github.com/bradenrayhorn/ledger-translator/grpc"
	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/provider/tda"
	"github.com/bradenrayhorn/ledger-translator/service"
	"github.com/spf13/viper"
)

func main() {
	println("ledger-translator initializing...")

	config.LoadConfig()

	println("connecting to databases...")
	oauthSessionDB := config.NewRedisClient("oauth_sessions")
	tokenDB := config.NewRedisClient("oauth_tokens")

	println("creating vault client...")
	vaultClient := createVaultClient()
	testVaultConnection(vaultClient)

	certify, err := config.CreateCertify()
	if err != nil {
		log.Printf("failed to create certify %s", err)
	}

	var providers []provider.Provider
	providers = append(providers, tda.NewTDAProvider())

	grpcServer := grpc.NewGRPCServer(tokenDB, vaultClient, providers, certify, config.GetCACertPool())
	grpcClient := NewGRPCClient(certify)

	go grpcServer.Start()

	controller := RouteController{
		providers:      providers,
		sessionService: service.NewSessionService(grpcClient.session),
		sessionDB:      oauthSessionDB,
		tokenDB:        tokenDB,
		vaultClient:    vaultClient,
	}

	router := CreateRouter(controller)

	port := viper.GetString("http_port")
	log.Printf("listening for http requests on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
