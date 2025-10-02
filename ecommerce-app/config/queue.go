package config

import (
    "fmt"

    "github.com/rabbitmq/amqp091-go"
)

func InitQueue(url string) (*amqp091.Connection, *amqp091.Channel, error) {
    conn, err := amqp091.Dial(url)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to connect queue: %w", err)
    }
    ch, err := conn.Channel()
    if err != nil {
        return nil, nil, fmt.Errorf("failed to open queue channel: %w", err)
    }
    return conn, ch, nil
}

