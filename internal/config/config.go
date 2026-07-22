package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppPort       string
	JWTSecret     string
	TokenLifetime time.Duration
	DatabaseURL   string
}

func Load() Config {
	lifetimeStr := os.Getenv("TOKEN_LIFETIME")
	lifetime, _ := strconv.Atoi(lifetimeStr)
	return Config{
		AppPort:       os.Getenv("APP_PORT"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		TokenLifetime: time.Duration(lifetime) * time.Minute,
		DatabaseURL:   os.Getenv("DATABASE_URL"),
	}
}
