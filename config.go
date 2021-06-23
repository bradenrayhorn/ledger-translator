package main

import (
	"crypto/x509"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func loadConfig() {
	envPath := os.Getenv("ENV_PATH")
	if len(envPath) == 0 {
		envPath = ".env"
	}
	err := godotenv.Load(envPath)
	if err != nil {
		log.Printf("failed to load .env: %e", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("http_port", "8080")
}

var certPool *x509.CertPool

func GetCACertPool() *x509.CertPool {
	if certPool != nil {
		return certPool
	}
	rootCertPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(viper.GetString("ca_cert_path"))
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	if ok := rootCertPool.AppendCertsFromPEM([]byte(strings.TrimSpace(string(pem)))); !ok {
		log.Println("failed to append pem")
		return nil
	}
	certPool = rootCertPool
	return certPool
}
