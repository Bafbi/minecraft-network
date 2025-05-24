package servers

import (
	"net"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/metadata"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	// "github.com/go-logr/logr" // Handled by cacheLog or specific registryLog if needed
)

// RegisterOrUpdateServer handles registration or update of a server with the Gate proxy
// and updates local caches.
// It's called by the KV watcher or potentially an API endpoint.
func RegisterOrUpdateServer(p *proxy.Proxy, name string, meta metadata.Metadata) {
	registryMu.Lock() // Protects p.Register and registeredServersCache
	defer registryMu.Unlock()

	address, exists := meta.GetAnnotation("server/address")
	if !exists {
		cacheLog.Error(nil, "Server address not found in metadata, cannot register", "serverName", name)
		return
	}
	serverAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		cacheLog.Error(err, "Failed to resolve server address, cannot register", "serverName", name, "addr", address)
		return
	}
	serverInfo := proxy.NewServerInfo(name, serverAddr) // Create ServerInfo

	// Update metadata cache regardless of registration status
	updateMetadataInCache(name, meta) // from cache.go

	if _, alreadyRegistered := registeredServersCache[name]; !alreadyRegistered {
		registeredServer, err := p.Register(serverInfo)
		if err != nil {
			cacheLog.Error(err, "Failed to register server with Gate proxy", "serverName", name)
			// If registration fails, we might not want to cache its metadata either, or mark it as "unhealthy"
			// For now, metadata is cached, but it won't be in registeredServersCache.
			return
		}
		addRegisteredServerToCache(name, registeredServer) // from cache.go
		cacheLog.Info("Server registered with Gate proxy", "serverName", name, "address", address)
	} else {
		// Server already known to Gate (e.g. address changed, but name is the same).
		// Gate itself doesn't have a direct "update server address" API.
		// Typically, you'd unregister and re-register if the address changes.
		// For now, we assume if it's in registeredServersCache, Gate knows about it.
		// The metadata update is handled by updateMetadataInCache above.
		cacheLog.Info("Server already registered, metadata updated", "serverName", name)
	}
}

// UnregisterServer handles unregistration of a server from the Gate proxy
// and updates local caches.
func UnregisterServer(p *proxy.Proxy, name string) {
	registryMu.Lock() // Protects p.Unregister and registeredServersCache
	defer registryMu.Unlock()

	deleteMetadataFromCache(name) // from cache.go

	if server, exists := registeredServersCache[name]; exists {
		p.Unregister(server.ServerInfo())
		deleteRegisteredServerFromCache(name) // from cache.go
		cacheLog.Info("Server unregistered from Gate proxy", "serverName", name)
	} else {
		cacheLog.Info("Server not found in Gate registration cache, no unregistration needed", "serverName", name)
	}
}
