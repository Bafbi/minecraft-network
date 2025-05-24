module github.com/bafbi/minecraft-network/services/permissions-editor

go 1.24.3

// This line tells Go where to find your permissions-checker service's module for local dev
require github.com/bafbi/minecraft-network/services/permissions-checker v0.0.0-00010101000000-000000000000 // Placeholder for your permissions-checker module

require (
	github.com/a-h/templ v0.3.865
	github.com/go-chi/chi/v5 v5.2.1
	github.com/nats-io/nats.go v1.42.0
	google.golang.org/grpc v1.72.1
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
)

// Add this section to map the module name to its local path
replace github.com/bafbi/minecraft-network/services/permissions-checker => ../permissions-checker
