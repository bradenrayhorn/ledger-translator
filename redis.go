package main

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

func NewRedisClient(name string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     viper.GetString(name + "_redis_addr"),
		Password: viper.GetString(name + "_redis_password"),
		DB:       viper.GetInt(name + "_redis_db"),
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Println(err)
	}

	return rdb
}
