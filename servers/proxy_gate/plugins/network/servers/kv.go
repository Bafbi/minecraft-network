package servers

import (
	"encoding/json"
	"fmt"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/metadata"
	"go.minekube.com/gate/pkg/edition/java/proxy"

	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	// No proxy.RegisteredServer or proxy.ServerInfo needed here, only metadata.
)

var (
	serversKV nats.KeyValue
	kvLog     logr.Logger
)

// InitializeKVStore initializes the NATS KV store for servers.
func InitializeKVStore(js nats.JetStreamContext, log logr.Logger) error {
	kvLog = log.WithName("ServerKV")
	var err error
	serversKV, err = js.KeyValue("servers")
	if err != nil {
		kvLog.Info("Servers KV store not found, attempting to create.")
		serversKV, err = js.CreateKeyValue(&nats.KeyValueConfig{Bucket: "servers"})
		if err != nil {
			kvLog.Error(err, "Failed to create servers KV store")
			return fmt.Errorf("failed to create servers KV store: %w", err)
		}
	}
	kvLog.Info("Server KV store initialized")
	return nil
}

// WatchKVStore watches for server metadata changes in NATS KV.
// It calls functions from registry.go to handle actual registration/unregistration.
func WatchKVStore(p *proxy.Proxy) { // Needs proxy.Proxy to pass to registry functions
	if serversKV == nil {
		kvLog.Error(nil, "Servers KV store not initialized. Cannot start watcher.")
		return
	}
	watcher, err := serversKV.WatchAll() // Don't ignore deletes for servers
	if err != nil {
		kvLog.Error(err, "Unable to start server KV watch")
		return
	}
	kvLog.Info("Starting server KV watcher")

	for entry := range watcher.Updates() {
		if entry == nil {
			continue
		}
		serverName := entry.Key()

		if entry.Operation() == nats.KeyValueDelete {
			// Unregister from Gate proxy and update local caches
			UnregisterServer(p, serverName) // From registry.go
		} else {
			var meta metadata.Metadata
			if err := json.Unmarshal(entry.Value(), &meta); err != nil {
				kvLog.Error(err, "Failed to unmarshal server metadata from KV", "key", serverName, "value", string(entry.Value()))
				continue
			}
			// Register with Gate proxy and update local caches
			RegisterOrUpdateServer(p, serverName, meta) // From registry.go
		}
	}
	kvLog.Info("Server KV watcher stopped.")
}

// --- Internal KV interaction functions (used by api.go or registry.go) ---
func putMetadataToKV(name string, meta metadata.Metadata) error {
	if serversKV == nil {
		return fmt.Errorf("servers KV not initialized")
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal server metadata for KV: %w", err)
	}
	_, err = serversKV.Put(name, data)
	return err
}

// deleteMetadataFromKV is not strictly needed if unregistration handles it,
// but can be a utility.
// func deleteMetadataFromKV(name string) error { ... }
