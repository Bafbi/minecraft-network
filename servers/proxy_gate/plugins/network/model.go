package network

import (
	"sync"

	"github.com/nats-io/nats.go"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var (
	serversMu       sync.RWMutex
	servers         = make([]proxy.RegisteredServer, 0) // name → server
	serversKV       nats.KeyValue                       // Global KV store reference
	playersKV       nats.KeyValue                       // Global players KV store reference
	serversMetadata = make(map[string]Metadata)         // name → server metadata
	playersMetadata = make(map[string]Metadata)         // name → player metadata
)

type Metadata struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}
