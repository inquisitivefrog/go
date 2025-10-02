package routes

import (
    "github.com/gin-gonic/gin"
    "github.com/inquisitivefrog/ecommerce-app/api/handlers"
    "github.com/inquisitivefrog/ecommerce-app/config"
    "github.com/inquisitivefrog/ecommerce-app/middleware"
)

func SetupProductRoutes(r *gin.Engine, handler *handlers.ProductHandler, cfg *config.Config) {
    api := r.Group("/api/v1")
    {
        products := api.Group("/products")
        products.Use(middleware.AuthMiddleware(cfg))
        {
            products.GET("", handler.GetAllProducts)
            products.GET("/:id", handler.GetProduct)
            products.GET("/search", handler.SearchProducts)
            products.POST("", middleware.AdminMiddleware(cfg), handler.CreateProduct)
            products.PUT("/:id", middleware.AdminMiddleware(cfg), handler.UpdateProduct)
            products.DELETE("/:id", middleware.AdminMiddleware(cfg), handler.DeleteProduct)
        }
    }
}
