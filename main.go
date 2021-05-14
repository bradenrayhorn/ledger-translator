package main

import (
	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/bradenrayhorn/ledger-translator/provider/tda"
)

func main() {
	println("ledger-translator initialization")

	var providers = []provider.Provider{}
	providers = append(providers, tda.TDAProvider{})

	for _, provider := range providers {
		println(provider.Key())
	}
}
