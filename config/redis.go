package config

import (
	"context"
	"crypto/tls"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

func NewRedisClient(name string) *redis.Client {
	options := &redis.Options{
		Addr:     viper.GetString(name + "_redis_addr"),
		Username: viper.GetString(name + "_redis_username"),
		Password: viper.GetString(name + "_redis_password"),
		DB:       viper.GetInt(name + "_redis_db"),
	}

	if viper.GetBool("redis_enable_tls") {
		options.TLSConfig = &tls.Config{
			RootCAs: GetCACertPool(),
		}
	}

	rdb := redis.NewClient(options)

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Println(err)
	}

	return rdb
}
