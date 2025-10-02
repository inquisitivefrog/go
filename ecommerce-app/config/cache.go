package config

import (
    "context"
    "fmt"

    "github.com/redis/go-redis/v9"
)

func InitCache(url string) (*redis.Client, error) {
    client := redis.NewClient(&redis.Options{Addr: url})
    if _, err := client.Ping(context.Background()).Result(); err != nil {
        return nil, fmt.Errorf("failed to connect to cache: %w", err)
    }
    return client, nil
}

