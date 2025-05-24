package core

import (
	"fmt"
	"log" // Or use your logger

	"github.com/nats-io/nats.go"
)

var nc *nats.Conn

const CasbinPolicyUpdateSubject = "network.casbin.policy.updated" // Same as proxy

func InitNats(natsURL string) error {
	var err error
	nc, err = nats.Connect(natsURL)
	if err != nil {
		return fmt.Errorf("webapp: failed to connect to NATS at %s: %w", natsURL, err)
	}
	log.Printf("Webapp connected to NATS at %s", nc.ConnectedUrl())
	return nil
}

func PublishNatsPolicyUpdate() {
	if nc == nil {
		log.Println("Webapp: NATS connection is nil, cannot publish policy update")
		return
	}
	if err := nc.Publish(CasbinPolicyUpdateSubject, []byte("reload_policies_from_webapp")); err != nil {
		log.Printf("Webapp: Failed to publish Casbin policy update to NATS: %v", err)
	} else {
		log.Println("Webapp: Published Casbin policy update to NATS")
	}
}

func CloseNats() {
	if nc != nil {
		nc.Close()
		log.Println("Webapp NATS connection closed")
	}
}
