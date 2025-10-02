package routes

import (
    "github.com/gin-gonic/gin"
    "github.com/inquisitivefrog/ecommerce-app/api/handlers"
    "github.com/inquisitivefrog/ecommerce-app/config"
    "github.com/inquisitivefrog/ecommerce-app/middleware"
)

func SetupUserRoutes(r *gin.RouterGroup, handler *handlers.UserHandler, cfg *config.Config) {
    r.POST("/register", handler.Register)
    r.POST("/login", handler.Login)
    protected := r.Group("/users").Use(middleware.AuthMiddleware(cfg))
    {
        protected.GET("/profile", handler.GetProfile)
    }
}
