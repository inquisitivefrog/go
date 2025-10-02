package services_test

import (
    "errors"
    "testing"

    "github.com/inquisitivefrog/ecommerce-app/models"
    "github.com/inquisitivefrog/ecommerce-app/repositories"
    "github.com/inquisitivefrog/ecommerce-app/services"
    "github.com/stretchr/testify/assert"
    "gorm.io/gorm"
)

type mockCartRepository struct {
    cartItems []models.Cart
}

func (m *mockCartRepository) AddItem(cartItem *models.Cart) error {
    m.cartItems = append(m.cartItems, *cartItem)
    return nil
}

func (m *mockCartRepository) GetCartByUserID(userID uint) ([]models.Cart, error) {
    var result []models.Cart
    for _, item := range m.cartItems {
        if item.UserID == userID {
            result = append(result, item)
        }
    }
    return result, nil
}

func TestCartService_AddToCart(t *testing.T) {
    mockCartRepo := &mockCartRepository{}
    mockProductRepo := &repositories.ProductRepository{DB: &gorm.DB{}} // Mocked, assumes product exists
    service := services.NewCartService(mockCartRepo, mockProductRepo)

    err := service.AddToCart(1, 1, 2)
    assert.NoError(t, err)

    cartItems, err := service.GetCart(1)
    assert.NoError(t, err)
    assert.Len(t, cartItems, 1)
    assert.Equal(t, uint(1), cartItems[0].UserID)
    assert.Equal(t, uint(1), cartItems[0].ProductID)
    assert.Equal(t, 2, cartItems[0].Quantity)
}

func TestCartService_AddToCart_Async(t *testing.T) {
    db := setupTestDB(t) // Assume helper to setup in-memory DB
    repo := repositories.NewCartRepository(db)
    productRepo := repositories.NewProductRepository(db)
    ch := setupTestRabbitMQ(t) // Assume helper for test RabbitMQ
    service := services.NewCartService(repo, productRepo, ch)
    config.RedisClient = &mockRedisClient{cache: make(map[string]string)}

    // Start worker in a goroutine
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        worker.RunCartWorker(repo, ch)
    }()

    // Add product to DB
    err := productRepo.CreateProduct(&models.Product{ID: 1, Name: "Shirt", Price: 29.99, Stock: 10})
    assert.NoError(t, err)

    // Test async add to cart
    err = service.AddToCart(1, 1, 2)
    assert.NoError(t, err)

    // Wait for worker to process
    time.Sleep(100 * time.Millisecond)

    // Verify cart item
    cartItems, err := repo.GetCartByUserID(1)
    assert.NoError(t, err)
    assert.Len(t, cartItems, 1)
    assert.Equal(t, uint(1), cartItems[0].ProductID)
    assert.Equal(t, 2, cartItems[0].Quantity)

    // Clean up
    close(ch.(*mockChannel).deliveries)
    wg.Wait()
}
