package main

import (
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
