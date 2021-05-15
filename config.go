package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func loadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to load .env: %e", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("http_port", "8080")
	viper.SetDefault("rsa_public_path", "jwt_rsa.pub")

}

func readKey(public bool, reader io.Reader) (interface{}, error) {
	keyBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var rsaKey interface{}
	if public {
		rsaKey, err = jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	} else {
		rsaKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	}
	return rsaKey, err
}

func readKeyFromFile(public bool, filePath string) (interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return readKey(public, file)
}
