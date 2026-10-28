package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct{ AppPort, DBHost, DBPort, DBName, DBUser, DBPassword, LogLevel, JWTSecret, ProductServiceURL, ConsulAddr string }

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Load() Config {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	return Config{
		AppPort: get("APP_PORT", "8080"), DBHost: get("DB_HOST", "localhost"), DBPort: get("DB_PORT", "5432"),
		DBName: get("DB_NAME", "orderdb"), DBUser: get("DB_USER", "postgres"), DBPassword: get("DB_PASSWORD", ""),
		LogLevel: get("LOG_LEVEL", "info"), JWTSecret: get("JWT_SECRET", "change_me"),
		ProductServiceURL: get("PRODUCT_SERVICE_URL", "http://localhost:8081"), ConsulAddr: get("CONSUL_ADDR", "http://localhost:8500"),
	}
}
