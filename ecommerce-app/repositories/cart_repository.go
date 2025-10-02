package repositories

import (
    "github.com/inquisitivefrog/ecommerce-app/models"
    "gorm.io/gorm"
)

// CartRepository interface defines methods for cart database operations
type CartRepository interface {
    AddItem(cartItem *models.Cart) error
    GetCartByUserID(userID uint) ([]models.Cart, error)
    GetCartItemByID(id uint) (*models.Cart, error)
    UpdateItem(cartItem *models.Cart) error
    DeleteItem(id uint) error
}

// cartRepository struct implements CartRepository
type cartRepository struct {
    DB *gorm.DB
}

// NewCartRepository creates a new CartRepository
func NewCartRepository(db *gorm.DB) CartRepository {
    return &cartRepository{DB: db}
}

// AddItem adds a cart item to the database
func (r *cartRepository) AddItem(cartItem *models.Cart) error {
    return r.DB.Create(cartItem).Error
}

// GetCartByUserID retrieves a user's cart
func (r *cartRepository) GetCartByUserID(userID uint) ([]models.Cart, error) {
    var cartItems []models.Cart
    err := r.DB.Where("user_id = ?", userID).Preload("Product").Find(&cartItems).Error
    return cartItems, err
}

// GetCartItemByID retrieves a specific cart item
func (r *cartRepository) GetCartItemByID(id uint) (*models.Cart, error) {
    var cartItem models.Cart
    err := r.DB.Preload("Product").First(&cartItem, id).Error
    return &cartItem, err
}

// UpdateItem updates a cart item
func (r *cartRepository) UpdateItem(cartItem *models.Cart) error {
    return r.DB.Save(cartItem).Error
}

// DeleteItem deletes a cart item
func (r *cartRepository) DeleteItem(id uint) error {
    return r.DB.Delete(&models.Cart{}, id).Error
}
