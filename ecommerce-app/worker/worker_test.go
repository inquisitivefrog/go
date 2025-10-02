package worker_test

import (
    "context"
    "encoding/json"
    "io"
    "sync"
    "testing"
    "time"

    "github.com/inquisitivefrog/ecommerce-app/config"
    "github.com/inquisitivefrog/ecommerce-app/models"
    "github.com/inquisitivefrog/ecommerce-app/repositories"
    "github.com/inquisitivefrog/ecommerce-app/worker"
    "github.com/rabbitmq/amqp091-go"
    "github.com/redis/go-redis/v9"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
)

// mockCartRepository mocks the CartRepository
type mockCartRepository struct {
    cartItems []models.Cart
    err       error
}

func (m *mockCartRepository) AddItem(cartItem *models.Cart) error {
    if m.err != nil {
        return m.err
    }
    m.cartItems = append(m.cartItems, *cartItem)
    return nil
}

func (m *mockCartRepository) GetCartByUserID(userID uint) ([]models.Cart, error) {
    if m.err != nil {
        return nil, m.err
    }
    var results []models.Cart
    for _, item := range m.cartItems {
        if item.UserID == userID {
            results = append(results, item)
        }
    }
    return results, nil
}

func (m *mockCartRepository) GetCartItemByID(id uint) (*models.Cart, error) {
    return nil, nil // Not used in worker
}

func (m *mockCartRepository) UpdateItem(cartItem *models.Cart) error {
    return nil // Not used in worker
}

func (m *mockCartRepository) DeleteItem(id uint) error {
    return nil // Not used in worker
}

// mockRedisClient mocks the Redis client
type mockRedisClient struct {
    cache map[string]string
}

func (m *mockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
    val, ok := m.cache[key]
    cmd := redis.NewStringCmd(ctx)
    if !ok {
        cmd.SetErr(redis.Nil)
    } else {
        cmd.SetVal(val)
    }
    return cmd
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
    m.cache[key] = value.(string)
    return redis.NewStatusCmd(ctx)
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
    for _, key := range keys {
        delete(m.cache, key)
    }
    return redis.NewIntCmd(ctx, len(keys))
}

// mockChannel mocks the RabbitMQ channel
type mockChannel struct {
    deliveries chan amqp091.Delivery
    acks       chan bool
    nacks      chan bool
}

func (m *mockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
    return m.deliveries, nil
}

func (m *mockChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
    return nil
}

func (m *mockChannel) Close() error {
    close(m.deliveries)
    return nil
}

func (m *mockChannel) Ack(tag uint64, multiple bool) error {
    m.acks <- true
    return nil
}

func (m *mockChannel) Nack(tag uint64, multiple, requeue bool) error {
    m.nacks <- requeue
    return nil
}

func TestRunCartWorker(t *testing.T) {
    // Setup mocks
    mockRepo := &mockCartRepository{}
    mockRedis := &mockRedisClient{cache: make(map[string]string)}
    config.RedisClient = mockRedis // Inject mock Redis client
    logrus.SetOutput(io.Discard)  // Suppress logs during testing

    // Create mock channel
    deliveries := make(chan amqp091.Delivery, 1)
    acks := make(chan bool, 1)
    nacks := make(chan bool, 1)
    ch := &mockChannel{
        deliveries: deliveries,
        acks:       acks,
        nacks:      nacks,
    }

    // Start worker in a goroutine
    var wg sync.WaitGroup
    wg.Add(1)
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        defer wg.Done()
        worker.RunCartWorker(mockRepo, ch)
    }()

    // Test case: Successful message processing
    t.Run("SuccessfulMessage", func(t *testing.T) {
        msg := worker.CartMessage{
            UserID:    1,
            ProductID: 1,
            Quantity:  2,
        }
        body, _ := json.Marshal(msg)
        delivery := amqp091.Delivery{Body: body, DeliveryTag: 1}
        deliveries <- delivery

        // Wait for processing or timeout
        select {
        case <-acks:
            // Verify cart item
            cartItems, err := mockRepo.GetCartByUserID(1)
            assert.NoError(t, err)
            assert.Len(t, cartItems, 1)
            assert.Equal(t, uint(1), cartItems[0].UserID)
            assert.Equal(t, uint(1), cartItems[0].ProductID)
            assert.Equal(t, 2, cartItems[0].Quantity)

            // Verify cache invalidation
            _, err = mockRedis.Get(context.Background(), "cart:1")
            assert.Equal(t, redis.Nil, err)
        case <-time.After(1 * time.Second):
            t.Fatal("Timeout waiting for message acknowledgment")
        }
    })

    // Test case: Invalid message
    t.Run("InvalidMessage", func(t *testing.T) {
        delivery := amqp091.Delivery{Body: []byte("invalid json"), DeliveryTag: 2}
        deliveries <- delivery

        // Wait for processing or timeout
        select {
        case requeue := <-nacks:
            assert.False(t, requeue, "Expected Nack without requeue")
            // Verify no cart items added
            cartItems, err := mockRepo.GetCartByUserID(1)
            assert.NoError(t, err)
            assert.Len(t, cartItems, 1) // Still only one from previous test
        case <-time.After(1 * time.Second):
            t.Fatal("Timeout waiting for message nack")
        }
    })

    // Test case: Repository error
    t.Run("RepositoryError", func(t *testing.T) {
        mockRepo.err = errors.New("database error")
        msg := worker.CartMessage{
            UserID:    1,
            ProductID: 2,
            Quantity:  3,
        }
        body, _ := json.Marshal(msg)
        delivery := amqp091.Delivery{Body: body, DeliveryTag: 3}
        deliveries <- delivery

        // Wait for processing or timeout
        select {
        case requeue := <-nacks:
            assert.True(t, requeue, "Expected Nack with requeue")
            // Verify no new cart items
            cartItems, err := mockRepo.GetCartByUserID(1)
            assert.NoError(t, err)
            assert.Len(t, cartItems, 1) // Still only one from first test
        case <-time.After(1 * time.Second):
            t.Fatal("Timeout waiting for message nack")
        }
    })

    // Clean up
    cancel()
    close(deliveries)
    wg.Wait()
}
