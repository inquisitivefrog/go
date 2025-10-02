package worker

import (
    "context"
    "encoding/json"
    "strconv"

    "github.com/inquisitivefrog/ecommerce-app/models"
    "github.com/inquisitivefrog/ecommerce-app/repositories"
    "github.com/rabbitmq/amqp091-go"
    "github.com/redis/go-redis/v9"
    "github.com/sirupsen/logrus"
)

// CartMessage represents a message from the cart queue
type CartMessage struct {
    UserID    uint `json:"user_id"`
    ProductID uint `json:"product_id"`
    Quantity  int  `json:"quantity"`
}

// RunCartWorker runs the cart worker to process RabbitMQ messages
func RunCartWorker(repo repositories.CartRepository, ch *amqp091.Channel, redisClient *redis.Client) {
    msgs, err := ch.Consume(
        "cart_queue", // Queue
        "",           // Consumer
        false,        // Auto-ack
        false,        // Exclusive
        false,        // No-local
        false,        // No-wait
        nil,          // Args
    )
    if err != nil {
        logrus.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "CONSUME_FAILED",
        }).Fatal("Failed to consume RabbitMQ queue")
    }

    logrus.Info("Worker started, waiting for cart messages")

    for msg := range msgs {
        var cartMsg CartMessage
        if err := json.Unmarshal(msg.Body, &cartMsg); err != nil {
            logrus.WithFields(logrus.Fields{
                "error":      err,
                "error_code": "UNMARSHAL_FAILED",
            }).Warn("Failed to unmarshal cart message")
            msg.Nack(false, false) // multiple=false, requeue=false
            continue
        }

        cartItem := &models.Cart{
            UserID:    cartMsg.UserID,
            ProductID: cartMsg.ProductID,
            Quantity:  cartMsg.Quantity,
        }
        if err := repo.AddItem(cartItem); err != nil { // Changed back to AddItem
            logrus.WithFields(logrus.Fields{
                "user_id":    cartMsg.UserID,
                "product_id": cartMsg.ProductID,
                "error":      err,
                "error_code": "ADD_ITEM_FAILED",
            }).Error("Failed to add cart item")
            msg.Nack(false, true) // multiple=false, requeue=true
            continue
        }

        // Invalidate cache
        cacheKey := "cart:" + strconv.FormatUint(uint64(cartMsg.UserID), 10)
        if err := redisClient.Del(context.Background(), cacheKey).Err(); err != nil {
            logrus.WithFields(logrus.Fields{
                "user_id":    cartMsg.UserID,
                "error":      err,
                "error_code": "CACHE_INVALIDATE",
            }).Warn("Failed to invalidate cache")
        }

        logrus.WithFields(logrus.Fields{
            "user_id":    cartMsg.UserID,
            "product_id": cartMsg.ProductID,
            "quantity":   cartMsg.Quantity,
        }).Info("Processed cart item")
        msg.Ack(false)
    }
}
