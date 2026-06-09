package config

import "os"

type Config struct {
	AppPort     string
	JWTSecret   string
	DatabaseURL string
	Timeout     string
}

func Load() Config {
	return Config{
		AppPort:     os.Getenv("APP_PORT"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Timeout:     os.Getenv("TIMEOUT"),
	}
}
