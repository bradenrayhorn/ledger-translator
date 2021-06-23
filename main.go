package main

import (
	"log"
	"net/http"

	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/provider/tda"
	"github.com/bradenrayhorn/ledger-translator/service"
	"github.com/spf13/viper"
)

func main() {
	println("ledger-translator initializing...")

	loadConfig()

	println("connecting to databases...")
	oauthSessionDB := NewRedisClient("oauth_sessions")
	tokenDB := NewRedisClient("oauth_tokens")

	println("creating vault client...")
	vaultClient := createVaultClient()
	testVaultConnection(vaultClient)

	grpcClient := NewGRPCClient()

	var providers []provider.Provider
	providers = append(providers, tda.NewTDAProvider())

	controller := RouteController{
		providers:      providers,
		sessionService: service.NewSessionService(grpcClient.session),
		sessionDB:      oauthSessionDB,
		tokenDB:        tokenDB,
		vaultClient:    vaultClient,
	}

	http.HandleFunc("/authenticate", controller.Authenticate)
	http.HandleFunc("/callback", controller.Callback)
	http.HandleFunc("/health-check", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("ok"))
	})

	port := viper.GetString("http_port")
	log.Printf("listening for http requests on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
