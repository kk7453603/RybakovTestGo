package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/kk7453603/RybakovTestGo/internal/adapters/grpc"
	"github.com/kk7453603/RybakovTestGo/internal/adapters/repository"
	"github.com/kk7453603/RybakovTestGo/internal/config"
	"github.com/kk7453603/RybakovTestGo/internal/core/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := connectDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	currencyRepo, priceRepo := repository.NewPostgresRepository(db)

	priceProvider := repository.NewExternalPriceProvider(*cfg)

	currencyService := services.NewCurrencyService(currencyRepo, priceRepo, priceProvider)

	grpcServer := grpc.NewGRPCServer(currencyService, cfg.GRPC.Port, cfg.Server.Port)

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- grpcServer.Start()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Fatalf("❌ Server error: %v", err)
	case sig := <-quit:
		log.Printf("🔄 Received signal: %v", sig)
	}

	log.Println("🛑 Shutting down server...")
	grpcServer.Stop()

	time.Sleep(2 * time.Second)
	log.Println("✅ Server stopped successfully")
}

func connectDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode, cfg.Timezone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to database successfully!")
	return db, nil
}
