package main

import (
	"log"
	"net/http"

	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/provider/tda"
	"github.com/spf13/viper"
)

func main() {
	println("ledger-translator initializing...")

	loadConfig()

	var providers []provider.Provider
	providers = append(providers, tda.NewTDAProvider())

	controller := RouteController{providers: providers}

	http.HandleFunc("/authenticate", controller.Authenticate)
	http.HandleFunc("/callback", controller.Callback)

	port := viper.GetString("http_port")
	log.Printf("listening for http requests on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
