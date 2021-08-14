package config

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/johanbrandhorst/certify"
	"github.com/johanbrandhorst/certify/issuers/vault"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
	logrusadapter "logur.dev/adapter/logrus"
)

func LoadConfig() {
	envPath := os.Getenv("ENV_PATH")
	fmt.Println(envPath)
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
	viper.SetDefault("grpc_port", "9001")
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

func CreateCertify() (*certify.Certify, error) {
	url, err := url.Parse(viper.GetString("vault_addr"))
	if err != nil {
		return nil, err
	}
	issuer := &vault.Issuer{
		URL:        url,
		Mount:      viper.GetString("vault_pki_mount"),
		AuthMethod: &vault.RenewingToken{Initial: loadVaultToken()},
		Role:       viper.GetString("vault_pki_role"),
		TimeToLive: time.Hour * 24,
	}

	log := logrus.New()
	log.Level = logrus.DebugLevel

	certify := &certify.Certify{
		Issuer:      issuer,
		CommonName:  viper.GetString("vault_pki_cn"),
		Cache:       certify.NewMemCache(),
		RenewBefore: time.Minute * 10,
		Logger:      logrusadapter.New(log),
	}
	return certify, nil
}

func loadVaultToken() string {
	tokenBytes, err := ioutil.ReadFile(viper.GetString("vault_token_path"))
	if err != nil {
		log.Println(err)
		return ""
	}

	return strings.TrimSpace(string(tokenBytes))
}
