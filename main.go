package main

import (
	"log"
	"net/http"

	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/provider/tda"
)

func main() {
	println("ledger-translator initialization")

	var providers []provider.Provider
	providers = append(providers, tda.NewTDAProvider())

	for _, provider := range providers {
		println(provider.Key())
	}

	controller := RouteController{providers: providers}

	http.HandleFunc("/authenticate", controller.Authenticate)
	http.HandleFunc("/callback", controller.Callback)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
