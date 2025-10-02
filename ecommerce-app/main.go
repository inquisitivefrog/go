package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "github.com/inquisitivefrog/ecommerce-app/config"
    "github.com/inquisitivefrog/ecommerce-app/docs"
    "github.com/inquisitivefrog/ecommerce-app/handlers"
    "github.com/inquisitivefrog/ecommerce-app/middleware"
    "github.com/inquisitivefrog/ecommerce-app/repositories"
    "github.com/inquisitivefrog/ecommerce-app/routes"
    "github.com/inquisitivefrog/ecommerce-app/services"

    ginSwagger "github.com/swaggo/gin-swagger"
    swaggerFiles "github.com/swaggo/files"
)

func main() {
    // --- Load config ---
    cfg, err := config.NewConfig()
    if err != nil {
        log.Fatalf("failed to initialize config: %v", err)
    }
    defer func() {
        if cfg.QueueConn != nil {
            cfg.QueueConn.Close()
        }
        if cfg.QueueChan != nil {
            cfg.QueueChan.Close()
        }
        if cfg.Cache != nil {
            cfg.Cache.Close()
        }
        // DB close (gorm.DB needs sqlDB extracted)
        if cfg.DB != nil {
            if sqlDB, err := cfg.DB.DB(); err == nil {
                sqlDB.Close()
            }
        }
    }()

    // --- Setup Gin ---
    r := gin.Default()
    r.Use(middleware.Logger(cfg.Logger), middleware.RateLimiter())

    // --- Swagger setup ---
    docs.SwaggerInfo.BasePath = "/api/v1"
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // --- Prometheus metrics ---
    r.GET("/metrics", middleware.PrometheusHandler())

    // --- Repositories ---
    userRepo := repositories.NewUserRepository(cfg.DB)
    productRepo := repositories.NewProductRepository(cfg.DB)
    orderRepo := repositories.NewOrderRepository(cfg.DB)
    cartRepo := repositories.NewCartRepository(cfg.DB)

    // --- Services ---
    userService := services.NewUserService(userRepo, cfg.JWTSecret, cfg.Logger)
    productService := services.NewProductService(productRepo, cfg.Cache, cfg.Logger)
    orderService := services.NewOrderService(orderRepo, productRepo, cfg.QueueChan, cfg.Logger)
    cartService := services.NewCartService(cartRepo, productRepo, cfg.QueueChan, cfg.Cache, cfg.Logger)

    // --- Handlers ---
    userHandler := handlers.NewUserHandler(userService)
    productHandler := handlers.NewProductHandler(productService)
    orderHandler := handlers.NewOrderHandler(orderService)
    cartHandler := handlers.NewCartHandler(cartService)

    // --- Routes ---
    api := r.Group("/api/v1")
    routes.SetupUserRoutes(api, userHandler)
    routes.SetupOrderRoutes(api, orderHandler)
    routes.SetupCartRoutes(api, cartHandler)
    routes.SetupProductRoutes(api, productHandler, cfg) // maybe pass just logger + cache instead of whole cfg

    // --- Start server ---
    if err := r.Run(cfg.ServerPort); err != nil {
        cfg.Logger.Fatalf("failed to run server: %v", err)
    }
}
