package services

import (
    "context"
    "encoding/json"
    "strconv"
    "time"

    "github.com/inquisitivefrog/ecommerce-app/models"
    "github.com/inquisitivefrog/ecommerce-app/repositories"
    "github.com/pkg/errors"
    "github.com/rabbitmq/amqp091-go"
    "github.com/redis/go-redis/v9"
    "github.com/sirupsen/logrus"
)

// Custom error types
var (
    ErrInvalidQuantity    = errors.New("quantity must be positive")
    ErrInsufficientStock  = errors.New("insufficient stock")
    ErrProductNotFound    = errors.New("product not found")
    ErrCartItemNotFound   = errors.New("cart item not found")
    ErrMarshalFailed      = errors.New("failed to marshal message")
    ErrPublishFailed      = errors.New("failed to publish to RabbitMQ")
    ErrCacheUnmarshal     = errors.New("failed to unmarshal cached cart")
    ErrCacheFailed        = errors.New("failed to cache cart")
    ErrCacheInvalidate    = errors.New("failed to invalidate cache")
    ErrFetchCartFailed    = errors.New("failed to fetch cart")
    ErrUpdateCartFailed   = errors.New("failed to update cart item")
    ErrDeleteCartFailed   = errors.New("failed to delete cart item")
)

type CartService struct {
    CartRepo    repositories.CartRepository
    ProductRepo repositories.ProductRepository
    RabbitMQ    *amqp091.Channel
    RedisClient *redis.Client
    Logger      *logrus.Logger
}

func NewCartService(cartRepo repositories.CartRepository, productRepo repositories.ProductRepository, rabbitMQ *amqp091.Channel, redisClient *redis.Client, logger *logrus.Logger) *CartService {
    return &CartService{
        CartRepo:    cartRepo,
        ProductRepo: productRepo,
        RabbitMQ:    rabbitMQ,
        RedisClient: redisClient,
        Logger:      logger,
    }
}

func (s *CartService) AddToCart(userID uint, productID uint, quantity int) error {
    product, err := s.ProductRepo.GetProductByID(productID)
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "product_id": productID,
            "error":      err,
            "error_code": "PRODUCT_NOT_FOUND",
        }).Warn("Product not found")
        return errors.Wrap(ErrProductNotFound, err.Error())
    }
    if product.Stock < quantity {
        s.Logger.WithFields(logrus.Fields{
            "product_id": productID,
            "stock":      product.Stock,
            "quantity":   quantity,
            "error_code": "INSUFFICIENT_STOCK",
        }).Warn("Insufficient stock")
        return ErrInsufficientStock
    }
    if quantity <= 0 {
        s.Logger.WithFields(logrus.Fields{
            "quantity":   quantity,
            "error_code": "INVALID_QUANTITY",
        }).Warn("Invalid quantity")
        return ErrInvalidQuantity
    }

    message := struct {
        UserID    uint `json:"user_id"`
        ProductID uint `json:"product_id"`
        Quantity  int  `json:"quantity"`
    }{
        UserID:    userID,
        ProductID: productID,
        Quantity:  quantity,
    }
    body, err := json.Marshal(message)
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "MARSHAL_FAILED",
        }).Error("Failed to marshal RabbitMQ message")
        return errors.Wrap(ErrMarshalFailed, err.Error())
    }

    err = s.RabbitMQ.PublishWithContext(
        context.Background(),
        "",
        "cart_queue",
        false,
        false,
        amqp091.Publishing{
            ContentType: "application/json",
            Body:        body,
        },
    )
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "PUBLISH_FAILED",
        }).Error("Failed to publish to RabbitMQ")
        return errors.Wrap(ErrPublishFailed, err.Error())
    }

    s.Logger.WithFields(logrus.Fields{
        "user_id":    userID,
        "product_id": productID,
        "quantity":   quantity,
    }).Info("Published add to cart message to RabbitMQ")
    return nil
}

func (s *CartService) GetCart(userID uint) ([]models.Cart, error) {
    cacheKey := "cart:" + strconv.FormatUint(uint64(userID), 10)
    cached, err := s.RedisClient.Get(context.Background(), cacheKey).Result()
    if err == nil {
        var cartItems []models.Cart
        if err := json.Unmarshal([]byte(cached), &cartItems); err == nil {
            s.Logger.WithFields(logrus.Fields{
                "user_id": userID,
                "count":   len(cartItems),
            }).Info("Fetched cart from cache")
            return cartItems, nil
        }
        s.Logger.WithFields(logrus.Fields{
            "user_id":    userID,
            "error":      err,
            "error_code": "CACHE_UNMARSHAL",
        }).Warn("Failed to unmarshal cached cart")
    }

    cartItems, err := s.CartRepo.GetCartByUserID(userID)
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "user_id":    userID,
            "error":      err,
            "error_code": "FETCH_CART_FAILED",
        }).Error("Failed to fetch cart")
        return nil, errors.Wrap(ErrFetchCartFailed, err.Error())
    }

    if len(cartItems) > 0 {
        data, err := json.Marshal(cartItems)
        if err == nil {
            s.RedisClient.Set(context.Background(), cacheKey, data, 10*time.Minute)
        } else {
            s.Logger.WithFields(logrus.Fields{
                "user_id":    userID,
                "error":      err,
                "error_code": "CACHE_FAILED",
            }).Warn("Failed to cache cart")
        }
    }

    s.Logger.WithFields(logrus.Fields{
        "user_id": userID,
        "count":   len(cartItems),
    }).Info("Fetched cart")
    return cartItems, nil
}

func (s *CartService) GetCartItemByID(id uint) (*models.Cart, error) {
    cartItem, err := s.CartRepo.GetCartItemByID(id)
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "cart_id":    id,
            "error":      err,
            "error_code": "CART_ITEM_NOT_FOUND",
        }).Warn("Cart item not found")
        return nil, errors.Wrap(ErrCartItemNotFound, err.Error())
    }
    return cartItem, nil
}

func (s *CartService) UpdateCartItem(id uint, quantity int) error {
    cartItem, err := s.CartRepo.GetCartItemByID(id)
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "cart_id":    id,
            "error":      err,
            "error_code": "CART_ITEM_NOT_FOUND",
        }).Warn("Cart item not found")
        return errors.Wrap(ErrCartItemNotFound, err.Error())
    }
    if quantity <= 0 {
        s.Logger.WithFields(logrus.Fields{
            "quantity":   quantity,
            "error_code": "INVALID_QUANTITY",
        }).Warn("Invalid quantity")
        return ErrInvalidQuantity
    }
    product, err := s.ProductRepo.GetProductByID(cartItem.ProductID)
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "product_id": cartItem.ProductID,
            "error":      err,
            "error_code": "PRODUCT_NOT_FOUND",
        }).Warn("Product not found")
        return errors.Wrap(ErrProductNotFound, err.Error())
    }
    if product.Stock < quantity {
        s.Logger.WithFields(logrus.Fields{
            "product_id": cartItem.ProductID,
            "stock":      product.Stock,
            "quantity":   quantity,
            "error_code": "INSUFFICIENT_STOCK",
        }).Warn("Insufficient stock")
        return ErrInsufficientStock
    }
    cartItem.Quantity = quantity
    if err := s.CartRepo.UpdateItem(cartItem); err != nil {
        s.Logger.WithFields(logrus.Fields{
            "cart_id":    id,
            "error":      err,
            "error_code": "UPDATE_CART_FAILED",
        }).Error("Failed to update cart item")
        return errors.Wrap(ErrUpdateCartFailed, err.Error())
    }

    cacheKey := "cart:" + strconv.FormatUint(uint64(cartItem.UserID), 10)
    if err := s.RedisClient.Del(context.Background(), cacheKey).Err(); err != nil {
        s.Logger.WithFields(logrus.Fields{
            "user_id":    cartItem.UserID,
            "error":      err,
            "error_code": "CACHE_INVALIDATE",
        }).Warn("Failed to invalidate cache")
    }

    s.Logger.WithFields(logrus.Fields{
        "cart_id":  id,
        "user_id":  cartItem.UserID,
        "quantity": quantity,
    }).Info("Updated cart item")
    return nil
}

func (s *CartService) DeleteCartItem(id uint) error {
    cartItem, err := s.CartRepo.GetCartItemByID(id)
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "cart_id":    id,
            "error":      err,
            "error_code": "CART_ITEM_NOT_FOUND",
        }).Warn("Cart item not found")
        return errors.Wrap(ErrCartItemNotFound, err.Error())
    }
    if err := s.CartRepo.DeleteItem(id); err != nil {
        s.Logger.WithFields(logrus.Fields{
            "cart_id":    id,
            "error":      err,
            "error_code": "DELETE_CART_FAILED",
        }).Error("Failed to delete cart item")
        return errors.Wrap(ErrDeleteCartFailed, err.Error())
    }

    cacheKey := "cart:" + strconv.FormatUint(uint64(cartItem.UserID), 10)
    if err := s.RedisClient.Del(context.Background(), cacheKey).Err(); err != nil {
        s.Logger.WithFields(logrus.Fields{
            "user_id":    cartItem.UserID,
            "error":      err,
            "error_code": "CACHE_INVALIDATE",
        }).Warn("Failed to invalidate cache")
    }

    s.Logger.WithFields(logrus.Fields{
        "cart_id": id,
        "user_id": cartItem.UserID,
    }).Info("Deleted cart item")
    return nil
}
