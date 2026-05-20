package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	JWTSecret string `json:"jwt_secret"`
	DBurl     string `json:"db_url"`
	Port      string `json:"port"`
}

func Load() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, fmt.Errorf("config open error: %v", err)
	}
	defer file.Close()

	cfg := &Config{}
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config decode error: %v", err)
	}

	if cfg.DBurl == "" {
		return nil, fmt.Errorf("DBurl is empty")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWTSecret is empty")
	}
	return cfg, nil
}
