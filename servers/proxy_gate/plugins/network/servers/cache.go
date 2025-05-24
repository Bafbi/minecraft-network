package servers

import (
	"sync"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/metadata" // Shared metadata
	"github.com/go-logr/logr"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	// For managing the registeredServersCache slice
)

var (
	// For proxy.RegisteredServer instances. Gate's registration is not inherently thread-safe for concurrent Register/Unregister.
	// This mutex protects our list and the calls to p.Register/Unregister.
	registryMu sync.Mutex
	// Cache of servers registered with the Gate proxy.
	registeredServersCache = make(map[string]proxy.RegisteredServer) // name -> RegisteredServer

	// For metadata.Metadata cache
	metadataCacheMu sync.RWMutex
	serversMetadata = make(map[string]metadata.Metadata) // name -> server metadata

	cacheLog logr.Logger
)

// InitCache initializes the server cache module.
func InitCache(log logr.Logger) {
	cacheLog = log.WithName("ServerCache")
	cacheLog.Info("Server cache initialized")
}

// --- Internal metadata cache functions ---
func updateMetadataInCache(name string, meta metadata.Metadata) {
	metadataCacheMu.Lock()
	defer metadataCacheMu.Unlock()
	serversMetadata[name] = meta
	cacheLog.V(1).Info("Updated server metadata in cache", "serverName", name)
}

func deleteMetadataFromCache(name string) {
	metadataCacheMu.Lock()
	defer metadataCacheMu.Unlock()
	delete(serversMetadata, name)
	cacheLog.V(1).Info("Deleted server metadata from cache", "serverName", name)
}

func getMetadataFromCache(name string) (metadata.Metadata, bool) {
	metadataCacheMu.RLock()
	defer metadataCacheMu.RUnlock()
	meta, exists := serversMetadata[name]
	return meta, exists
}

// --- Internal registered server cache functions (used by registry.go) ---
func addRegisteredServerToCache(name string, server proxy.RegisteredServer) {
	// registryMu is assumed to be held by the caller (registry.go)
	registeredServersCache[name] = server
	cacheLog.V(1).Info("Added server to Gate registration cache", "serverName", name)
}

func getRegisteredServerFromCache(name string) (proxy.RegisteredServer, bool) {
	// registryMu is assumed to be held by the caller (registry.go)
	server, exists := registeredServersCache[name]
	return server, exists
}

func deleteRegisteredServerFromCache(name string) {
	// registryMu is assumed to be held by the caller (registry.go)
	delete(registeredServersCache, name)
	cacheLog.V(1).Info("Removed server from Gate registration cache", "serverName", name)
}

func getAllRegisteredServersFromCache() []proxy.RegisteredServer {
	// registryMu is assumed to be held by the caller or this needs its own lock if called independently
	registryMu.Lock() // Or RLock if just reading the map keys/values into a new slice
	defer registryMu.Unlock()

	list := make([]proxy.RegisteredServer, 0, len(registeredServersCache))
	for _, srv := range registeredServersCache {
		list = append(list, srv)
	}
	return list
}
