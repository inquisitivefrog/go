package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/inquisitivefrog/ecommerce-app/models"
    "github.com/inquisitivefrog/ecommerce-app/services"
    "github.com/inquisitivefrog/ecommerce-app/utils"
    "github.com/sirupsen/logrus" // Added import
)

// CartHandler handles HTTP requests for carts
type CartHandler struct {
    CartService *services.CartService
}

// NewCartHandler creates a new CartHandler
func NewCartHandler(cartService *services.CartService) *CartHandler {
    return &CartHandler{CartService: cartService}
}

// AddToCart handles POST /api/cart
func (h *CartHandler) AddToCart(c *gin.Context) {
    user, exists := c.Get("user")
    if !exists {
        h.CartService.Logger.WithFields(logrus.Fields{
            "path": c.Request.URL.Path,
        }).Warn("No user in context for POST /api/cart")
        utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }
    userID := user.(models.User).ID // Extract UserID from models.User

    var input struct {
        ProductID uint `json:"product_id" binding:"required"`
        Quantity  int  `json:"quantity" binding:"required,min=1"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        h.CartService.Logger.WithFields(logrus.Fields{
            "error": err,
        }).Warn("Invalid input for POST /api/cart")
        utils.RespondWithError(c, http.StatusBadRequest, err.Error())
        return
    }

    err := h.CartService.AddToCart(userID, input.ProductID, input.Quantity)
    if err != nil {
        h.CartService.Logger.WithFields(logrus.Fields{
            "user_id":    userID,
            "product_id": input.ProductID,
            "error":      err,
        }).Warn("Failed to add item to cart")
        utils.RespondWithError(c, http.StatusBadRequest, err.Error())
        return
    }

    h.CartService.Logger.WithFields(logrus.Fields{
        "user_id":    userID,
        "product_id": input.ProductID,
        "quantity":   input.Quantity,
    }).Info("Enqueued add to cart request")
    c.JSON(http.StatusAccepted, gin.H{"message": "Item enqueued for addition to cart"})
}

// GetCart handles GET /api/cart
func (h *CartHandler) GetCart(c *gin.Context) {
    user, exists := c.Get("user")
    if !exists {
        h.CartService.Logger.WithFields(logrus.Fields{
            "path": c.Request.URL.Path,
        }).Warn("No user in context for GET /api/cart")
        utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }
    userID := user.(models.User).ID // Extract UserID from models.User

    cartItems, err := h.CartService.GetCart(userID)
    if err != nil {
        h.CartService.Logger.WithFields(logrus.Fields{
            "user_id": userID,
            "error":   err,
        }).Error("Failed to fetch cart")
        utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
        return
    }

    // Format response to include product details
    type CartResponse struct {
        models.Cart
        ProductName  string  `json:"product_name"`
        ProductPrice float64 `json:"product_price"`
    }
    var response []CartResponse
    for _, item := range cartItems {
        response = append(response, CartResponse{
            Cart:         item,
            ProductName:  item.Product.Name,
            ProductPrice: item.Product.Price,
        })
    }

    h.CartService.Logger.WithFields(logrus.Fields{
        "user_id": userID,
        "count":   len(cartItems),
    }).Info("Fetched cart")
    c.JSON(http.StatusOK, response)
}

// GetCartItem handles GET /api/cart/:id
func (h *CartHandler) GetCartItem(c *gin.Context) {
    user, exists := c.Get("user")
    if !exists {
        h.CartService.Logger.WithFields(logrus.Fields{
            "path": c.Request.URL.Path,
        }).Warn("No user in context for GET /api/cart/:id")
        utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }
    userID := user.(models.User).ID // Extract UserID from models.User

    cartID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        h.CartService.Logger.WithFields(logrus.Fields{
            "cart_id": c.Param("id"),
            "error":   err,
        }).Warn("Invalid cart ID")
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid cart ID")
        return
    }

    cartItem, err := h.CartService.GetCartItemByID(uint(cartID))
    if err != nil {
        h.CartService.Logger.WithFields(logrus.Fields{
            "cart_id": cartID,
            "user_id": userID,
            "error":   err,
        }).Warn("Cart item not found")
        utils.RespondWithError(c, http.StatusNotFound, "Cart item not found")
        return
    }
    if cartItem.UserID != userID {
        h.CartService.Logger.WithFields(logrus.Fields{
            "cart_id": cartID,
            "user_id": userID,
        }).Warn("Unauthorized cart access")
        utils.RespondWithError(c, http.StatusForbidden, "Unauthorized")
        return
    }

    response := struct {
        models.Cart
        ProductName  string  `json:"product_name"`
        ProductPrice float64 `json:"product_price"`
    }{
        Cart:         *cartItem,
        ProductName:  cartItem.Product.Name,
        ProductPrice: cartItem.Product.Price,
    }

    h.CartService.Logger.WithFields(logrus.Fields{
        "cart_id": cartID,
        "user_id": userID,
    }).Info("Fetched cart item")
    c.JSON(http.StatusOK, response)
}

// UpdateCartItem handles PUT /api/cart/:id
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
    user, exists := c.Get("user")
    if !exists {
        h.CartService.Logger.WithFields(logrus.Fields{
            "path": c.Request.URL.Path,
        }).Warn("No user in context for PUT /api/cart/:id")
        utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }
    userID := user.(models.User).ID // Extract UserID from models.User

    cartID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        h.CartService.Logger.WithFields(logrus.Fields{
            "cart_id": c.Param("id"),
            "error":   err,
        }).Warn("Invalid cart ID")
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid cart ID")
        return
    }

    var input struct {
        Quantity int `json:"quantity" binding:"required,min=1"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        h.CartService.Logger.WithFields(logrus.Fields{
            "error": err,
        }).Warn("Invalid input for PUT /api/cart/:id")
        utils.RespondWithError(c, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.CartService.UpdateCartItem(uint(cartID), input.Quantity); err != nil {
        h.CartService.Logger.WithFields(logrus.Fields{
            "cart_id": cartID,
            "user_id": userID,
            "error":   err,
        }).Warn("Failed to update cart item")
        utils.RespondWithError(c, http.StatusBadRequest, err.Error())
        return
    }

    h.CartService.Logger.WithFields(logrus.Fields{
        "cart_id":  cartID,
        "user_id":  userID,
        "quantity": input.Quantity,
    }).Info("Updated cart item")
    c.JSON(http.StatusOK, gin.H{"message": "Cart item updated successfully"})
}

// DeleteCartItem handles DELETE /api/cart/:id
func (h *CartHandler) DeleteCartItem(c *gin.Context) {
    user, exists := c.Get("user")
    if !exists {
        h.CartService.Logger.WithFields(logrus.Fields{
            "path": c.Request.URL.Path,
        }).Warn("No user in context for DELETE /api/cart/:id")
        utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }
    userID := user.(models.User).ID // Extract UserID from models.User

    cartID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        h.CartService.Logger.WithFields(logrus.Fields{
            "cart_id": c.Param("id"),
            "error":   err,
        }).Warn("Invalid cart ID")
        utils.RespondWithError(c, http.StatusBadRequest, "Invalid cart ID")
        return
    }

    if err := h.CartService.DeleteCartItem(uint(cartID)); err != nil {
        h.CartService.Logger.WithFields(logrus.Fields{
            "cart_id": cartID,
            "user_id": userID,
            "error":   err,
        }).Warn("Cart item not found")
        utils.RespondWithError(c, http.StatusNotFound, "Cart item not found")
        return
    }

    h.CartService.Logger.WithFields(logrus.Fields{
        "cart_id": cartID,
        "user_id": userID,
    }).Info("Deleted cart item")
    c.JSON(http.StatusOK, gin.H{"message": "Cart item deleted successfully"})
}
