package config

import (
    "fmt"

    "github.com/rabbitmq/amqp091-go"
    "github.com/redis/go-redis/v9"
    "github.com/sirupsen/logrus"
    "github.com/spf13/viper"
    "gorm.io/gorm"
)

type Config struct {
    DB         *gorm.DB
    Cache      *redis.Client
    QueueConn  *amqp091.Connection
    QueueChan  *amqp091.Channel
    Logger     *logrus.Logger
    ServerPort string
    JWTSecret  string
}

func NewConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.AddConfigPath(".")
    viper.AutomaticEnv()

    viper.SetDefault("SERVER_PORT", ":8080")
    viper.SetDefault("DATABASE_URL", "postgres://postgres:secret@postgres:5432/ecommerce?sslmode=disable")
    viper.SetDefault("CACHE_URL", "redis:6379")
    viper.SetDefault("QUEUE_URL", "amqp://guest:guest@rabbitmq:5672/")

    logger := InitLogger()

    db, err := InitDB(viper.GetString("DATABASE_URL"), logger)
    if err != nil {
        return nil, err
    }

    cache, err := InitCache(viper.GetString("CACHE_URL"))
    if err != nil {
        return nil, err
    }

    queueConn, queueChan, err := InitQueue(viper.GetString("QUEUE_URL"))
    if err != nil {
        return nil, err
    }

    secret := viper.GetString("JWT_SECRET")
    if secret == "" {
        return nil, fmt.Errorf("JWT_SECRET environment variable is required")
    }

    return &Config{
        DB:         db,
        Cache:      cache,
        QueueConn:  queueConn,
        QueueChan:  queueChan,
        Logger:     logger,
        ServerPort: viper.GetString("SERVER_PORT"),
        JWTSecret:  secret,
    }, nil
}

