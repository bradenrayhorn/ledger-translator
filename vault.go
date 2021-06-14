package main

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

func createVaultClient() *api.Client {
	vaultClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Println(err)
		return nil
	}

	tokenBytes, err := ioutil.ReadFile(viper.GetString("vault_token_path"))
	if err != nil {
		log.Println(err)
		return nil
	}

	vaultClient.SetToken(strings.TrimSpace(string(tokenBytes)))
	return vaultClient
}

func testVaultConnection(client *api.Client) {
	if client == nil {
		log.Println("vault client does not exist")
		return
	}
	_, err := client.Logical().Read("auth/token/lookup-self")
	if err != nil {
		log.Println(err)
	}
}
