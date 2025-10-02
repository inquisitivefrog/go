package routes

import (
    "github.com/gin-gonic/gin"
    "github.com/inquisitivefrog/ecommerce-app/api/handlers"
    "github.com/inquisitivefrog/ecommerce-app/config"
    "github.com/inquisitivefrog/ecommerce-app/middleware"
)

func SetupCartRoutes(r *gin.RouterGroup, cartHandler *handlers.CartHandler, cfg *config.Config) {
    protected := r.Group("/api").Use(middleware.AuthMiddleware(cfg))
    {
        protected.POST("/cart", cartHandler.AddToCart)
        protected.GET("/cart", cartHandler.GetCart)
        protected.GET("/cart/:id", cartHandler.GetCartItem)
        protected.DELETE("/cart/:id", cartHandler.DeleteCartItem)
        protected.PUT("/cart/:id", cartHandler.UpdateCartItem)
    }
}
