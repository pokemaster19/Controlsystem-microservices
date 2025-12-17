package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	UsersPort        string
	DBHost           string
	DBName           string
	PostgresUser     string
	PostgresPassword string
	PostgresPort     string
	DBSSLMode        string
}

func Load() *Config {

	_ = godotenv.Load()
	viper.AutomaticEnv()

	cfg := &Config{
		UsersPort:        getEnv("ORDERS_PORT", "8081"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBName:           getEnv("DB_NAME_ORDERS", "postgres"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "password"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
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

func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.PostgresPort,
		c.PostgresUser,
		c.PostgresPassword,
		c.DBName,
		c.DBSSLMode,
	)
}
