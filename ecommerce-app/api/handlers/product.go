package handlers

import (
    "net/http"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/inquisitivefrog/ecommerce-app/models"
    "github.com/inquisitivefrog/ecommerce-app/services"
    "github.com/inquisitivefrog/ecommerce-app/utils"
    "github.com/sirupsen/logrus"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
    ProductService *services.ProductService
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(productService *services.ProductService) *ProductHandler {
    return &ProductHandler{ProductService: productService}
}

// CreateProduct handles POST /api/v1/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
    var product models.Product
    if err := c.ShouldBindJSON(&product); err != nil {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "INVALID_INPUT",
        }).Warn("Invalid input for POST /api/v1/products")
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid input")
        return
    }

    if err := h.ProductService.CreateProduct(&product); err != nil {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "CREATE_PRODUCT_FAILED",
        }).Warn("Failed to create product")
        utils.RespondWithError(c, http.StatusBadRequest, "Failed to create product")
        return
    }

    h.ProductService.Logger.WithFields(logrus.Fields{
        "product_id": product.ID,
    }).Info("Created product")
    c.JSON(http.StatusCreated, product)
}

// GetAllProducts handles GET /api/v1/products
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
    products, err := h.ProductService.GetProducts()
    if err != nil {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "FETCH_PRODUCTS_FAILED",
        }).Error("Failed to fetch products")
        utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch products")
        return
    }
    h.ProductService.Logger.WithFields(logrus.Fields{
        "count": len(products),
    }).Info("Fetched products")
    c.JSON(http.StatusOK, products)
}

// GetProduct handles GET /api/v1/products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "id":         c.Param("id"),
            "error":      err,
            "error_code": "INVALID_PRODUCT_ID",
        }).Warn("Invalid product ID")
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid product ID")
        return
    }

    product, err := h.ProductService.GetProductByID(uint(id))
    if err != nil {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "id":         id,
            "error":      err,
            "error_code": "PRODUCT_NOT_FOUND",
        }).Warn("Product not found")
        utils.RespondWithError(c, http.StatusNotFound, "Product not found")
        return
    }
    h.ProductService.Logger.WithFields(logrus.Fields{
        "product_id": id,
    }).Info("Fetched product")
    c.JSON(http.StatusOK, product)
}

// UpdateProduct handles PUT /api/v1/products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "id":         c.Param("id"),
            "error":      err,
            "error_code": "INVALID_PRODUCT_ID",
        }).Warn("Invalid product ID")
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid product ID")
        return
    }

    var input models.Product
    if err := c.ShouldBindJSON(&input); err != nil {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "INVALID_INPUT",
        }).Warn("Invalid input for PUT /api/v1/products/:id")
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid input")
        return
    }

    input.ID = uint(id)
    if err := h.ProductService.UpdateProduct(&input); err != nil {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "product_id": id,
            "error":      err,
            "error_code": "UPDATE_PRODUCT_FAILED",
        }).Warn("Failed to update product")
        utils.RespondWithError(c, http.StatusBadRequest, "Failed to update product")
        return
    }

    h.ProductService.Logger.WithFields(logrus.Fields{
        "product_id": id,
    }).Info("Updated product")
    c.JSON(http.StatusOK, input)
}

// DeleteProduct handles DELETE /api/v1/products/:id
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "id":         c.Param("id"),
            "error":      err,
            "error_code": "INVALID_PRODUCT_ID",
        }).Warn("Invalid product ID")
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid product ID")
        return
    }

    if err := h.ProductService.DeleteProduct(uint(id)); err != nil {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "product_id": id,
            "error":      err,
            "error_code": "DELETE_PRODUCT_FAILED",
        }).Warn("Product not found")
        utils.RespondWithError(c, http.StatusNotFound, "Product not found")
        return
    }

    h.ProductService.Logger.WithFields(logrus.Fields{
        "product_id": id,
    }).Info("Deleted product")
    c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// SearchProducts handles GET /api/v1/products/search?q=term&page=1&limit=10
func (h *ProductHandler) SearchProducts(c *gin.Context) {
    query := strings.TrimSpace(c.Query("q"))
    if len(query) < 3 {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "query":      query,
            "error_code": "INVALID_QUERY",
        }).Warn("Invalid search query")
        utils.RespondWithError(c, http.StatusBadRequest, "Search query must be at least 3 characters")
        return
    }

    page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
    if err != nil || page < 1 {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "page":       c.Query("page"),
            "error":      err,
            "error_code": "INVALID_PAGE",
        }).Warn("Invalid page number")
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid page number")
        return
    }

    limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
    if err != nil || limit < 1 || limit > 100 {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "limit":      c.Query("limit"),
            "error":      err,
            "error_code": "INVALID_LIMIT",
        }).Warn("Invalid limit")
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid limit, must be between 1 and 100")
        return
    }

    products, err := h.ProductService.SearchProducts(query, page, limit)
    if err != nil {
        h.ProductService.Logger.WithFields(logrus.Fields{
            "query":      query,
            "error":      err,
            "error_code": "SEARCH_PRODUCTS_FAILED",
        }).Warn("Failed to search products")
        utils.RespondWithError(c, http.StatusInternalServerError, "Failed to search products")
        return
    }

    h.ProductService.Logger.WithFields(logrus.Fields{
        "query": query,
        "count": len(products),
        "page":  page,
        "limit": limit,
    }).Info("Searched products")
    c.JSON(http.StatusOK, gin.H{
        "products": products,
        "page":     page,
        "limit":    limit,
    })
}
