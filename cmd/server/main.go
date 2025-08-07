package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/kk7453603/RybakovTestGo/internal/adapters/grpc"
	"github.com/kk7453603/RybakovTestGo/internal/adapters/repository"
	"github.com/kk7453603/RybakovTestGo/internal/config"
	"github.com/kk7453603/RybakovTestGo/internal/core/services"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключаемся к базе данных
	db, err := connectDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Создаем репозитории
	currencyRepo, priceRepo := repository.NewPostgresRepository(db)

	// Создаем внешний провайдер цен
	priceProvider := repository.NewExternalPriceProvider()

	// Создаем сервис
	currencyService := services.NewCurrencyService(currencyRepo, priceRepo, priceProvider)

	// Создаем и запускаем gRPC сервер
	grpcServer := grpc.NewGRPCServer(currencyService, cfg.GRPC.Port, cfg.Server.Port)

	// Запускаем сервер в горутине
	go func() {
		if err := grpcServer.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Обработка graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	grpcServer.Stop()
}

// connectDatabase подключается к базе данных PostgreSQL
func connectDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode, cfg.Timezone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Проверяем соединение
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
