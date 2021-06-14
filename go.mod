module github.com/bradenrayhorn/ledger-translator

go 1.16

require (
	github.com/bradenrayhorn/ledger-protos v0.2.0
	github.com/go-redis/redis/v8 v8.8.2
	github.com/hashicorp/go-hclog v0.16.1 // indirect
	github.com/hashicorp/vault v1.7.2
	github.com/hashicorp/vault/api v1.1.0
	github.com/hashicorp/vault/sdk v0.2.1-0.20210519002511-48c5544c77f4
	github.com/joho/godotenv v1.3.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v0.20.0 // indirect
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c
	google.golang.org/grpc v1.37.1
)
