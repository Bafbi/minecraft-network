package config

import (
	"fmt"
	"os"
)

type AppConfig struct {
	ValkeyHost      string
	ValkeyPort      string
	ValkeyPassword  string
	NatsURL         string
	CasbinModelPath string // Path to the casbin model file (mounted from ConfigMap)
}

func Load() (*AppConfig, error) {
	cfg := &AppConfig{
		ValkeyHost:      os.Getenv("VALKEY_HOST_WEBAPP"), // Use specific env vars for webapp
		ValkeyPort:      os.Getenv("VALKEY_PORT_WEBAPP"),
		ValkeyPassword:  os.Getenv("VALKEY_PASSWORD_WEBAPP"),
		NatsURL:         os.Getenv("NATS_URL_WEBAPP"),
		CasbinModelPath: os.Getenv("CASBIN_MODEL_PATH_WEBAPP"),
	}

	if cfg.ValkeyHost == "" {
		return nil, fmt.Errorf("VALKEY_HOST_WEBAPP must be set")
	}
	if cfg.ValkeyPort == "" {
		cfg.ValkeyPort = "6379"
	}
	if cfg.NatsURL == "" {
		// Default to the same NATS as the proxy if not specified for webapp
		cfg.NatsURL = os.Getenv("NATS_URL") // Fallback to general NATS_URL
		if cfg.NatsURL == "" {
			cfg.NatsURL = "nats://network-nats:4222" // Hardcoded default
		}
	}
	if cfg.CasbinModelPath == "" {
		return nil, fmt.Errorf("CASBIN_MODEL_PATH_WEBAPP must be set")
	}

	return cfg, nil
}
