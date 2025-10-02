package services_test

import (
    "errors"
    "strings"
    "testing"

    "github.com/inquisitivefrog/ecommerce-app/models"
    "github.com/inquisitivefrog/ecommerce-app/repositories"
    "github.com/inquisitivefrog/ecommerce-app/services"
    "github.com/stretchr/testify/assert"
    "gorm.io/gorm"
)

// mockProductRepository mocks the ProductRepository interface
type mockProductRepository struct {
    products []models.Product
    err      error
}

// NewMockProductRepository creates a new mock repository
func NewMockProductRepository(products []models.Product) *mockProductRepository {
    return &mockProductRepository{products: products}
}

// CreateProduct mocks creating a product
func (m *mockProductRepository) CreateProduct(product *models.Product) error {
    if m.err != nil {
        return m.err
    }
    m.products = append(m.products, *product)
    return nil
}

// GetProducts mocks retrieving all products
func (m *mockProductRepository) GetProducts() ([]models.Product, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.products, nil
}

// GetProductByID mocks retrieving a product by ID
func (m *mockProductRepository) GetProductByID(id uint) (*models.Product, error) {
    if m.err != nil {
        return nil, m.err
    }
    for _, product := range m.products {
        if product.ID == id {
            return &product, nil
        }
    }
    return nil, gorm.ErrRecordNotFound
}

// SearchProducts mocks searching products by name or description
func (m *mockProductRepository) SearchProducts(query string) ([]models.Product, error) {
    if m.err != nil {
        return nil, m.err
    }
    var results []models.Product
    for _, product := range m.products {
        if contains(product.Name, query) || contains(product.Description, query) {
            results = append(results, product)
        }
    }
    return results, nil
}

// contains checks if a string contains a substring (case-insensitive)
func contains(str, substr string) bool {
    return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

func TestProductService_CreateProduct(t *testing.T) {
    mockRepo := NewMockProductRepository([]models.Product{})
    service := services.NewProductService(mockRepo)

    // Test valid product
    product := &models.Product{
        Name:        "Shirt",
        Description: "Blue cotton shirt",
        Price:       29.99,
        Stock:       10,
    }
    err := service.CreateProduct(product)
    assert.NoError(t, err)
    assert.Len(t, mockRepo.products, 1)
    assert.Equal(t, "Shirt", mockRepo.products[0].Name)

    // Test invalid product (empty name)
    invalidProduct := &models.Product{
        Name:        "",
        Description: "Invalid product",
        Price:       19.99,
        Stock:       5,
    }
    err = service.CreateProduct(invalidProduct)
    assert.Error(t, err)
    assert.Equal(t, "invalid product data", err.Error())
    assert.Len(t, mockRepo.products, 1) // No new product added

    // Test invalid product (negative price)
    invalidProduct = &models.Product{
        Name:        "Pants",
        Description: "Black jeans",
        Price:       -10.0,
        Stock:       5,
    }
    err = service.CreateProduct(invalidProduct)
    assert.Error(t, err)
    assert.Equal(t, "invalid product data", err.Error())
    assert.Len(t, mockRepo.products, 1) // No new product added
}

func TestProductService_GetProducts(t *testing.T) {
    mockRepo := NewMockProductRepository([]models.Product{
        {ID: 1, Name: "Shirt", Description: "Blue cotton shirt", Price: 29.99, Stock: 10},
        {ID: 2, Name: "Pants", Description: "Black jeans", Price: 49.99, Stock: 5},
    })
    service := services.NewProductService(mockRepo)

    products, err := service.GetProducts()
    assert.NoError(t, err)
    assert.Len(t, products, 2)
    assert.Equal(t, "Shirt", products[0].Name)
    assert.Equal(t, "Pants", products[1].Name)

    // Test repository error
    mockRepo.err = errors.New("database error")
    _, err = service.GetProducts()
    assert.Error(t, err)
    assert.Equal(t, "database error", err.Error())
}

func TestProductService_GetProductByID(t *testing.T) {
    mockRepo := NewMockProductRepository([]models.Product{
        {ID: 1, Name: "Shirt", Description: "Blue cotton shirt", Price: 29.99, Stock: 10},
    })
    service := services.NewProductService(mockRepo)

    // Test existing product
    product, err := service.GetProductByID(1)
    assert.NoError(t, err)
    assert.Equal(t, "Shirt", product.Name)

    // Test non-existent product
    _, err = service.GetProductByID(2)
    assert.Error(t, err)
    assert.Equal(t, gorm.ErrRecordNotFound, err)

    // Test repository error
    mockRepo.err = errors.New("database error")
    _, err = service.GetProductByID(1)
    assert.Error(t, err)
    assert.Equal(t, "database error", err.Error())
}

func TestProductService_SearchProducts(t *testing.T) {
    mockRepo := NewMockProductRepository([]models.Product{
        {ID: 1, Name: "Shirt", Description: "Blue cotton shirt", Price: 29.99, Stock: 10},
        {ID: 2, Name: "Pants", Description: "Black jeans", Price: 49.99, Stock: 5},
    })
    service := services.NewProductService(mockRepo)

    // Test valid search
    products, err := service.SearchProducts("shirt")
    assert.NoError(t, err)
    assert.Len(t, products, 1)
    assert.Equal(t, "Shirt", products[0].Name)

    // Test search with no results
    products, err = service.SearchProducts("jacket")
    assert.NoError(t, err)
    assert.Len(t, products, 0)

    // Test empty query
    _, err = service.SearchProducts("")
    assert.Error(t, err)
    assert.Equal(t, "search query cannot be empty", err.Error())

    // Test repository error
    mockRepo.err = errors.New("database error")
    _, err = service.SearchProducts("shirt")
    assert.Error(t, err)
    assert.Equal(t, "database error", err.Error())
}
