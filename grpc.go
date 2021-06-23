package main

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/bradenrayhorn/ledger-protos/session"
	"github.com/johanbrandhorst/certify"
	"github.com/johanbrandhorst/certify/issuers/vault"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	logrusadapter "logur.dev/adapter/logrus"
)

type GRPCClient struct {
	session session.SessionAuthenticatorClient
}

func NewGRPCClient() GRPCClient {
	certify, _ := createCertify()
	tlsConfig := &tls.Config{
		GetClientCertificate: certify.GetClientCertificate,
		RootCAs:              GetCACertPool(),
	}
	conn, err := grpc.Dial(viper.GetString("grpc_host_auth"), grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		log.Fatalf("could not connect to grpc auth: %v", err)
	}

	sessionClient := session.NewSessionAuthenticatorClient(conn)
	return GRPCClient{
		session: sessionClient,
	}
}

func loadVaultToken() string {
	tokenBytes, err := ioutil.ReadFile(viper.GetString("vault_token_path"))
	if err != nil {
		log.Println(err)
		return ""
	}

	return strings.TrimSpace(string(tokenBytes))
}

func createCertify() (*certify.Certify, error) {
	url, err := url.Parse(viper.GetString("vault_url"))
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
