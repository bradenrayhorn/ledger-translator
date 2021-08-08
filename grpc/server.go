package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"

	"github.com/bradenrayhorn/ledger-protos/provider"
	"github.com/go-redis/redis/v8"
	"github.com/johanbrandhorst/certify"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GRPCServer struct {
	tokenRedisClient *redis.Client
	certify          *certify.Certify
	certPool         *x509.CertPool
}

func NewGRPCServer(tokenRedisClient *redis.Client, certify *certify.Certify, certPool *x509.CertPool) GRPCServer {
	return GRPCServer{
		tokenRedisClient: tokenRedisClient,
		certify:          certify,
		certPool:         certPool,
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

	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))

	provider.RegisterProviderServiceServer(grpcServer, NewProviderServiceServer(s.tokenRedisClient))

	if err := grpcServer.Serve(lis); err != nil {
		log.Println(err)
		return
	}
}
