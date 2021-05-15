package main

import (
	"crypto/rsa"
	"log"
	"net/http"

	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/provider/tda"
	"github.com/spf13/viper"
)

func main() {
	println("ledger-translator initializing...")

	loadConfig()

	publicKey, err := readKeyFromFile(true, viper.GetString("rsa_public_path"))
	if err != nil {
		log.Printf("failed to load public key: %e", err)
	}

	jwtService := JWTService{publicKey: publicKey.(*rsa.PublicKey)}

	var providers []provider.Provider
	providers = append(providers, tda.NewTDAProvider())

	controller := RouteController{providers: providers, jwtService: jwtService}

	http.HandleFunc("/authenticate", controller.Authenticate)
	http.HandleFunc("/callback", controller.Callback)

	port := viper.GetString("http_port")
	log.Printf("listening for http requests on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
