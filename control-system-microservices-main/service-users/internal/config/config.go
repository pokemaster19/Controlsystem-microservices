package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	UsersPort                string
	DBHost                   string
	DBName                   string
	PostgresUser             string
	PostgresPassword         string
	PostgresPort             string
	DBSSLMode                string
	TokenSecret              string
	RefreshTokenSecret       string
	TokenMinuteLifespan      string
	RefreshTokenHourLifespan string
	SuperadminPassword       string
}

func Load() *Config {

	_ = godotenv.Load()
	viper.AutomaticEnv()

	cfg := &Config{
		UsersPort:                getEnv("USERS_PORT", "8082"),
		DBHost:                   getEnv("DB_HOST", "localhost"),
		DBName:                   getEnv("DB_NAME_USERS", "postgres"),
		PostgresUser:             getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:         getEnv("POSTGRES_PASSWORD", "password"),
		PostgresPort:             getEnv("POSTGRES_PORT", "5432"),
		DBSSLMode:                getEnv("DB_SSLMODE", "disable"),
		TokenSecret:              getEnv("TOKEN_SECRET", "default-token-secret"),
		RefreshTokenSecret:       getEnv("REFRESH_TOKEN_SECRET", "default-refresh-token-secret"),
		TokenMinuteLifespan:      getEnv("TOKEN_MINUTE_LIFESPAN", "15"),
		RefreshTokenHourLifespan: getEnv("REFRESH_TOKEN_HOUR_LIFESPAN", "24"),
		SuperadminPassword:       getEnv("SUPERADMIN_PASSWORD", "default-superadmin-password"),
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
