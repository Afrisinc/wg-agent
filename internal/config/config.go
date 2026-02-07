package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Host       string
	Port       int
	APIKey     string
	WgInterface string
	PublicKeyPath string
	LogLevel   string
}

func Load() (*Config, error) {
	cfg := &Config{
		Host:       getEnv("HOST", "0.0.0.0"),
		Port:       getEnvInt("PORT", 9999),
		APIKey:     getEnv("API_KEY", ""),
		WgInterface: getEnv("WG_INTERFACE", "wg0"),
		PublicKeyPath: getEnv("PUBLIC_KEY_PATH", "/etc/wireguard/publickey"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API_KEY environment variable not set")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
