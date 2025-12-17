package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Addr             string
	UsersServiceURL  string
	OrdersServiceURL string
	JWTSecret        string
}

func Load() *Config {
	_ = godotenv.Load()
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env not found, using environment variables: %v", err)
	}

	cfg := &Config{
		Addr:             ":" + getEnv("GATEWAY_PORT", "8080"),
		UsersServiceURL:  getEnv("USERS_SERVICE_URL", "http://service-users:8082"),
		OrdersServiceURL: getEnv("ORDERS_SERVICE_URL", "http://service-orders:8081"),
		JWTSecret:        getEnv("TOKEN_SECRET", "your-default-secret"),
	}

	if cfg.Addr == ":" || cfg.UsersServiceURL == "" || cfg.OrdersServiceURL == "" {
		log.Fatalf("Missing required configuration: Addr=%s, UsersServiceURL=%s, OrdersServiceURL=%s",
			cfg.Addr, cfg.UsersServiceURL, cfg.OrdersServiceURL)
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	value := viper.GetString(key)
	if value == "" {
		value = os.Getenv(key)
	}
	if value == "" {
		return defaultValue
	}
	return value
}
