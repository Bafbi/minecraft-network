package config

import (
	"log"
	"os"
)

type Config struct {
	GRPCPort             string
	ValkeyAddr           string
	ValkeyPassword       string
	ValkeyKey            string
	NATSAddr             string
	NATSUser             string
	NATSPassword         string
	PlayerMetadataPrefix string // e.g., "player.metadata."
	ServerMetadataPrefix string // e.g., "server.metadata."
}

func LoadConfig() *Config {
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "localhost:6379"
	}
	valkeyPassword := os.Getenv("VALKEY_PASSWORD")
	valkeyKey := os.Getenv("VALKEY_DB")
	if valkeyKey == "" {
		valkeyKey = "casbin"
	}

	natsAddr := os.Getenv("NATS_ADDR")
	if natsAddr == "" {
		natsAddr = "nats://localhost:4222"
	}
	natsUser := os.Getenv("NATS_USER")
	natsPassword := os.Getenv("NATS_PASSWORD")

	playerMetaPrefix := os.Getenv("PLAYER_METADATA_PREFIX")
	if playerMetaPrefix == "" {
		playerMetaPrefix = "player."
	}
	serverMetaPrefix := os.Getenv("SERVER_METADATA_PREFIX")
	if serverMetaPrefix == "" {
		serverMetaPrefix = "server."
	}

	log.Printf("Loading config: GRPC_PORT=%s, VALKEY_ADDR=%s, NATS_ADDR=%s, PlayerMetaPrefix=%s, ServerMetaPrefix=%s",
		grpcPort, valkeyAddr, natsAddr, playerMetaPrefix, serverMetaPrefix)

	return &Config{
		GRPCPort:             grpcPort,
		ValkeyAddr:           valkeyAddr,
		ValkeyPassword:       valkeyPassword,
		ValkeyKey:            valkeyKey,
		NATSAddr:             natsAddr,
		NATSUser:             natsUser,
		NATSPassword:         natsPassword,
		PlayerMetadataPrefix: playerMetaPrefix,
		ServerMetadataPrefix: serverMetaPrefix,
	}
}
