package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"

	pbProvider "github.com/bradenrayhorn/ledger-protos/provider"
	pbQuotes "github.com/bradenrayhorn/ledger-protos/quotes"
	"github.com/bradenrayhorn/ledger-translator/provider"
	"github.com/go-redis/redis/v8"
	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/johanbrandhorst/certify"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	tokenRedisClient *redis.Client
	resolver         *ProviderResolver
	certify          *certify.Certify
	certPool         *x509.CertPool
}

func NewGRPCServer(tokenRedisClient *redis.Client, vaultClient *vaultAPI.Client, providers []provider.Provider, certify *certify.Certify, certPool *x509.CertPool) GRPCServer {
	return GRPCServer{
		tokenRedisClient: tokenRedisClient,
		resolver: &ProviderResolver{
			VaultClient: vaultClient,
			TokenDB:     tokenRedisClient,
			Providers:   providers,
		},
		certify:  certify,
		certPool: certPool,
	}
}

func (s GRPCServer) Start() {
	requestedPort := viper.GetString("grpc_port")
	log.Printf("starting grpc server on port %s", requestedPort)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", requestedPort))
	if err != nil {
		log.Println(err)
		return
	}

	tlsConfig := &tls.Config{
		GetCertificate: s.certify.GetCertificate,
		ClientCAs:      s.certPool,
		ClientAuth:     tls.RequireAndVerifyClientCert,
	}
	fmt.Println(tlsConfig.ClientCAs)

	//grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	grpcServer := grpc.NewServer()

	pbProvider.RegisterProviderServiceServer(grpcServer, NewProviderServiceServer(s.tokenRedisClient))
	pbQuotes.RegisterQuotesServiceServer(grpcServer, NewQuotesServiceServer(s.resolver))

	if err := grpcServer.Serve(lis); err != nil {
		log.Println(err)
		return
	}
}
