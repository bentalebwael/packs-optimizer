package main

import (
	"fmt"
	"log"

	"go.uber.org/zap"

	"github.com/pack-calculator/api"
	"github.com/pack-calculator/config"
	"github.com/pack-calculator/internal/order_calculations"
	"github.com/pack-calculator/internal/pack_configurations"
	"github.com/pack-calculator/pkg/logger"
	"github.com/pack-calculator/pkg/postgres"
)

func main() {
	// Initialize logger
	l, err := logger.New()
	if err != nil {
		log.Fatalf("Failed to intialize the logger: %v", err)
	}
	defer l.Sync()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		l.Fatal("Failed to load configuration", zap.Error(err))
	}
	l.Info("Configuration loaded")

	// Connect to database
	db, err := postgres.NewConnection(cfg.Database.URL)
	if err != nil {
		l.Fatal("Failed to connect to database", zap.Error(err))
	}
	l.Info("Database loaded")

	// Migrate database
	if err := postgres.Migrate(cfg.Database.URL); err != nil {
		l.Fatal("Failed to migrate database", zap.Error(err))
	}
	l.Info("Migrations executed")

	// Initialize repositories
	packsCfgRepo := pack_configurations.NewRepository(db)
	calculationsCfgRepo := order_calculations.NewRepository(db)
	l.Info("database repositories initialized")

	// Initialize services
	packsService := pack_configurations.NewService(l, packsCfgRepo)
	calculationsService := order_calculations.NewService(l, calculationsCfgRepo, packsCfgRepo)
	l.Info("services initialized")

	// Initialize handlers
	packsHandler := pack_configurations.NewHandler(l, packsService)
	calculationsHandler := order_calculations.NewHandler(l, calculationsService)
	l.Info("handlers initialized")

	// Setup router
	router := api.SetupRouter(l, cfg, packsHandler, calculationsHandler)
	l.Info("router initialized")

	// Start server
	l.Info(fmt.Sprintf("Server listening on port %s", cfg.Server.Port))
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		l.Fatal("Failed to start server", zap.Error(err))
	}
}
