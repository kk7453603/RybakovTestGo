package config

import (
	"os"
	"strconv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	GRPC     GRPCConfig
	APIToken string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	Timezone string
}

type ServerConfig struct {
	Port int
	Host string
}

type GRPCConfig struct {
	Port int
	Host string
}

func Load() (*Config, error) {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	serverPort, _ := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	grpcPort, _ := strconv.Atoi(getEnv("GRPC_PORT", "9090"))

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "crypto_currency_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			Timezone: getEnv("DB_TIMEZONE", "UTC"),
		},
		Server: ServerConfig{
			Port: serverPort,
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		GRPC: GRPCConfig{
			Port: grpcPort,
			Host: getEnv("GRPC_HOST", "0.0.0.0"),
		},
		APIToken: getEnv("API_TOKEN", ""),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
