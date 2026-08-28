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
	Timeout       string
	SMTPHost      string
	SMTPPort      int
	SMTPUserName  string
	SMTPPassword  string
	SMTPFrom      string
}

func Load() Config {
	lifetimeStr := os.Getenv("TOKEN_LIFETIME")
	lifetime, _ := strconv.Atoi(lifetimeStr)
	smtpPortStr := os.Getenv("SMTPPORT")
	smtpPort, _ := strconv.Atoi(smtpPortStr)

	return Config{
		AppPort:       os.Getenv("APP_PORT"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		TokenLifetime: time.Duration(lifetime) * time.Minute,
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Timeout:       os.Getenv("TIMEOUT"),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      smtpPort,
		SMTPUserName:  os.Getenv("SMTP_USERNAME"),
		SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:      os.Getenv("SMTP_FROM"),
	}
}
