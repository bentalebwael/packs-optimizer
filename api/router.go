package api

import (
	"github.com/gin-gonic/gin"
	"github.com/pack-calculator/config"
	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/pack-calculator/api/middleware"
	"github.com/pack-calculator/internal/order_calculations"
	"github.com/pack-calculator/internal/pack_configurations"
)

func SetupRouter(logger *zap.Logger, cfg *config.AppConfig, packCfgHandler *pack_configurations.Handler, calculationsHandler *order_calculations.Handler) *gin.Engine {
	// Create Gin router without default logging
	router := gin.New()

	// Use our custom logger and recovery middleware
	router.Use(middleware.Logger(logger))
	router.Use(gin.Recovery())

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimiter)

	// API routes with rate limiter
	apiGroup := router.Group("/api", rateLimiter.Middleware())
	{
		apiGroup.GET("/packs", packCfgHandler.GetActivePackConfiguration)
		apiGroup.POST("/packs", middleware.ValidatePacks(), packCfgHandler.CreatePackConfiguration)
		apiGroup.POST("/calculate", middleware.ValidateOrder(), calculationsHandler.CalculatePacksForOrder)
	}

	// Serve static files from /static URL path
	router.Static("/static", "./static")
	// Serve index.html for root path
	router.StaticFile("/", "./static/index.html")

	// Serve Swagger documentation
	router.StaticFile("/swagger.yaml", "./api/swagger.yaml")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.URL("/swagger.yaml")))

	return router
}
