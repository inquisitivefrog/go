package services

import (
    "github.com/inquisitivefrog/ecommerce-app/models"
    "github.com/inquisitivefrog/ecommerce-app/repositories"
    "github.com/pkg/errors"
    "github.com/sirupsen/logrus"
)

// ProductService handles business logic for products
type ProductService struct {
    ProductRepo repositories.ProductRepository
    Logger      *logrus.Logger
}

// NewProductService creates a new ProductService
func NewProductService(productRepo repositories.ProductRepository, logger *logrus.Logger) *ProductService {
    return &ProductService{
        ProductRepo: productRepo,
        Logger:      logger,
    }
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(product *models.Product) error {
    if product.Name == "" || product.Price <= 0 || product.Stock < 0 {
        s.Logger.WithFields(logrus.Fields{
            "error_code": "INVALID_PRODUCT_DATA",
        }).Warn("Invalid product data")
        return errors.New("invalid product data")
    }
    return s.ProductRepo.CreateProduct(product)
}

// GetProducts retrieves all products
func (s *ProductService) GetProducts() ([]models.Product, error) {
    products, err := s.ProductRepo.GetProducts()
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "FETCH_PRODUCTS_FAILED",
        }).Error("Failed to fetch products")
        return nil, err
    }
    s.Logger.WithFields(logrus.Fields{
        "count": len(products),
    }).Info("Fetched products")
    return products, nil
}

// GetProductByID retrieves a product by ID
func (s *ProductService) GetProductByID(id uint) (*models.Product, error) {
    product, err := s.ProductRepo.GetProductByID(id)
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "product_id": id,
            "error":      err,
            "error_code": "PRODUCT_NOT_FOUND",
        }).Warn("Product not found")
        return nil, err
    }
    s.Logger.WithFields(logrus.Fields{
        "product_id": id,
    }).Info("Fetched product")
    return product, nil
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(product *models.Product) error {
    if product.Name == "" || product.Price <= 0 || product.Stock < 0 {
        s.Logger.WithFields(logrus.Fields{
            "product_id": product.ID,
            "error_code": "INVALID_PRODUCT_DATA",
        }).Warn("Invalid product data")
        return errors.New("invalid product data")
    }
    err := s.ProductRepo.UpdateProduct(product)
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "product_id": product.ID,
            "error":      err,
            "error_code": "UPDATE_PRODUCT_FAILED",
        }).Warn("Failed to update product")
        return err
    }
    s.Logger.WithFields(logrus.Fields{
        "product_id": product.ID,
    }).Info("Updated product")
    return nil
}

// DeleteProduct deletes a product by ID
func (s *ProductService) DeleteProduct(id uint) error {
    err := s.ProductRepo.DeleteProduct(id)
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "product_id": id,
            "error":      err,
            "error_code": "DELETE_PRODUCT_FAILED",
        }).Warn("Product not found")
        return err
    }
    s.Logger.WithFields(logrus.Fields{
        "product_id": id,
    }).Info("Deleted product")
    return nil
}

// SearchProducts searches products by query with pagination
func (s *ProductService) SearchProducts(query string, page, limit int) ([]models.Product, error) {
    if query == "" {
        s.Logger.WithFields(logrus.Fields{
            "query":      query,
            "error_code": "INVALID_QUERY",
        }).Warn("Empty search query")
        return nil, errors.New("search query cannot be empty")
    }
    products, err := s.ProductRepo.SearchProducts(query, page, limit)
    if err != nil {
        s.Logger.WithFields(logrus.Fields{
            "query":      query,
            "page":       page,
            "limit":      limit,
            "error":      err,
            "error_code": "SEARCH_PRODUCTS_FAILED",
        }).Warn("Failed to search products")
        return nil, err
    }
    s.Logger.WithFields(logrus.Fields{
        "query": query,
        "page":  page,
        "limit": limit,
        "count": len(products),
    }).Info("Searched products")
    return products, nil
}
