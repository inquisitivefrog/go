package repositories

import (
    "fmt"

    "github.com/inquisitivefrog/ecommerce-app/models"
    "gorm.io/gorm"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
    CreateProduct(product *models.Product) error
    GetProducts() ([]models.Product, error)
    GetProductByID(id uint) (*models.Product, error)
    SearchProducts(query string, page, limit int) ([]models.Product, error)
    UpdateProduct(product *models.Product) error
    DeleteProduct(id uint) error
}

// productRepository implements ProductRepository
type productRepository struct {
    db *gorm.DB
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository(db *gorm.DB) ProductRepository {
    return &productRepository{db: db}
}

func (r *productRepository) CreateProduct(product *models.Product) error {
    return r.db.Create(product).Error
}

func (r *productRepository) GetProducts() ([]models.Product, error) {
    var products []models.Product
    err := r.db.Where("deleted_at IS NULL").Find(&products).Error // Add deleted_at filter
    return products, err
}

func (r *productRepository) GetProductByID(id uint) (*models.Product, error) {
    var product models.Product
    err := r.db.Where("deleted_at IS NULL").First(&product, id).Error // Add deleted_at filter
    if err != nil {
        return nil, err
    }
    return &product, nil
}

func (r *productRepository) SearchProducts(query string, page, limit int) ([]models.Product, error) {
    var products []models.Product
    offset := (page - 1) * limit
    query = "%" + query + "%"
    err := r.db.Where("name ILIKE ? OR description ILIKE ?", query, query).
        Where("deleted_at IS NULL").
        Offset(offset).
        Limit(limit).
        Find(&products).Error
    if err != nil {
        return nil, fmt.Errorf("failed to search products: %w", err)
    }
    return products, nil
}

func (r *productRepository) UpdateProduct(product *models.Product) error {
    return r.db.Save(product).Error
}

func (r *productRepository) DeleteProduct(id uint) error {
    return r.db.Delete(&models.Product{}, id).Error
}
