package servers

import (
	"fmt"
	"math/rand"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/metadata"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	// "github.com/go-logr/logr" // Logging done by cache/kv
)

var DefaultServerSelector = map[string]string{"type": "lobby"}

// GetRegisteredServerByName returns a registered server by its name, if present.
func GetRegisteredServerByName(name string) (proxy.RegisteredServer, bool) {
	return getRegisteredServerFromCache(name)
}

// GetMetadataByName retrieves server metadata by name from the cache.
func GetMetadataByName(name string) (metadata.Metadata, bool) {
	return getMetadataFromCache(name)
}

// UpdateMetadataByName updates server metadata in cache and persists to NATS KV.
// This will also trigger the KV watcher, which will call RegisterOrUpdateServer.
func UpdateMetadataByName(name string, modFunc func(meta *metadata.Metadata)) error {
	metadataCacheMu.Lock() // Ensure atomic read-modify-write for cache before KV push
	defer metadataCacheMu.Unlock()

	currentMeta, exists := serversMetadata[name]
	if !exists {
		// If server doesn't exist in cache, creating it via API might be complex
		// as it also needs to be registered with Gate.
		// This API is primarily for updating existing servers' metadata.
		// New servers should appear via KV watcher from an external source.
		return fmt.Errorf("server %s not found in cache for metadata update", name)
	}
	metaToModify := currentMeta
	modFunc(&metaToModify)

	// Persist to KV, which will then trigger the watcher to update Gate registration if needed.
	if err := putMetadataToKV(name, metaToModify); err != nil {
		return fmt.Errorf("failed to persist server metadata to KV for %s: %w", name, err)
	}
	// Do NOT update the local metadata cache here. The KV watcher will handle it.
	cacheLog.V(1).Info("API updated server metadata, KV updated", "serverName", name)
	return nil
}

// FindServersByLabels returns a list of *proxy.RegisteredServer* instances matching the label selectors.
func FindServersByLabels(selectors map[string]string) []proxy.RegisteredServer {
	registryMu.Lock() // Protect registeredServersCache read
	defer registryMu.Unlock()
	metadataCacheMu.RLock() // Protect serversMetadata read
	defer metadataCacheMu.RUnlock()

	var result []proxy.RegisteredServer
	for name, registeredServer := range registeredServersCache {
		meta, exists := serversMetadata[name]
		if !exists {
			continue // Should not happen if caches are consistent
		}
		if meta.MatchesLabels(selectors) { // Assumes MatchesLabels is a method on metadata.Metadata
			result = append(result, registeredServer)
		}
	}
	return result
}

// GetRandomDefaultServer attempts to find a random server matching the default selector.
func GetRandomDefaultServer() (proxy.RegisteredServer, bool) {
	defaultServers := FindServersByLabels(DefaultServerSelector)
	if len(defaultServers) == 0 {
		return nil, false
	}
	return defaultServers[rand.Intn(len(defaultServers))], true
}

// GetAllRegisteredServers returns all servers currently registered with Gate.
func GetAllRegisteredServers() []proxy.RegisteredServer {
	return getAllRegisteredServersFromCache() // from cache.go
}
