package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time" // Added for context timeout in Ping

	"github.com/casbin/casbin/v2"
	casbinredisadapter "github.com/casbin/redis-adapter/v2"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/bafbi/minecraft-network/services/permissions-checker/auth"
	"github.com/bafbi/minecraft-network/services/permissions-checker/cache"
	"github.com/bafbi/minecraft-network/services/permissions-checker/config"
)

type authService struct {
	auth.UnimplementedAuthServiceServer
	enforcer      *casbin.Enforcer
	metadataCache *cache.MetadataCache
}

func NewAuthService(e *casbin.Enforcer, mc *cache.MetadataCache) *authService {
	return &authService{enforcer: e, metadataCache: mc}
}

// CheckPermission implements the gRPC method
func (s *authService) CheckPermission(ctx context.Context, req *auth.AuthRequest) (*auth.AuthResponse, error) {
	log.Printf("CheckPermission Request for Player %s (%s): Action='%s', Resource='%s', Server='%s'",
		req.GetPlayerName(), req.GetPlayerUuid(), req.GetAction(), req.GetResource(), req.GetServerName())

	// 1. Get player metadata from local cache
	playerAttrs := s.metadataCache.GetPlayerMetadata(req.GetPlayerUuid())
	if playerAttrs == nil {
		log.Printf("Warning: Player metadata not found in cache for %s. Defaulting to empty attributes.", req.GetPlayerUuid())
		playerAttrs = &structpb.Struct{} // Provide empty struct to Casbin
	}

	// Important: Casbin policies might expect player_uuid to be in the player attributes map
	// Add player_uuid to the attributes if it's not already there
	playerAttrsMap := playerAttrs.AsMap()
	if _, ok := playerAttrsMap["player_uuid"]; !ok {
		playerAttrsMap["player_uuid"] = req.GetPlayerUuid()
	}

	// 2. Get server metadata from local cache (if server_name is provided)
	serverAttrs := &structpb.Struct{} // Default to empty
	if req.GetServerName() != "" {
		serverAttrs = s.metadataCache.GetServerMetadata(req.GetServerName())
		if serverAttrs == nil {
			log.Printf("Warning: Server metadata not found in cache for %s. Defaulting to empty attributes.", req.GetServerName())
			serverAttrs = &structpb.Struct{} // Provide empty struct to Casbin
		}
	}

	// Casbin expects arguments as interfaces.
	// The order must match model.conf's request_definition (r = player, server, action, resource)
	// Pass the *modified* playerAttrsMap
	allowed, err := s.enforcer.Enforce(playerAttrsMap, serverAttrs.AsMap(), req.GetAction(), req.GetResource())
	if err != nil {
		log.Printf("Casbin enforcement error: %v", err)
		return &auth.AuthResponse{Allowed: false, Message: fmt.Sprintf("Internal error: %v", err)}, nil
	}

	decision := "DENIED"
	if allowed {
		decision = "ALLOWED"
	}
	log.Printf("Decision for Player %s (%s) on %s %s (Server: %s): %s",
		req.GetPlayerName(), req.GetPlayerUuid(), req.GetAction(), req.GetResource(), req.GetServerName(), decision)

	return &auth.AuthResponse{Allowed: allowed, Message: fmt.Sprintf("Permission %s", decision)}, nil
}

// AddPolicy implements the gRPC method to add policies
func (s *authService) AddPolicy(ctx context.Context, req *auth.PolicyManagementRequest) (*auth.PolicyManagementResponse, error) {
	var addedCount int
	for _, rule := range req.GetRules() {
		// Casbin's AddPolicy expects string arguments matching the policy_definition (p = id, target_action, target_resource, player_condition_expr, server_condition_expr, effect, priority)
		ok, err := s.enforcer.AddPolicy(
			rule.GetId(),
			rule.GetTargetAction(),
			rule.GetTargetResource(),
			rule.GetPlayerConditionExpression(),
			rule.GetServerConditionExpression(),
			rule.GetEffect(),
			fmt.Sprintf("%d", rule.GetPriority()), // Convert int32 to string for Casbin
		)
		if err != nil {
			log.Printf("Error adding policy %s: %v", rule.GetId(), err)
			return &auth.PolicyManagementResponse{Success: false, Message: fmt.Sprintf("Error adding policy %s: %v", rule.GetId(), err)}, nil
		}
		if ok {
			addedCount++
			log.Printf("Added policy: %v", rule)
		} else {
			log.Printf("Policy already exists: %v", rule)
		}
	}
	s.enforcer.LoadPolicy() // Reload policy to ensure consistency if not automatically done by adapter
	return &auth.PolicyManagementResponse{Success: true, Message: fmt.Sprintf("Successfully added %d policies", addedCount)}, nil
}

// RemovePolicy implements the gRPC method to remove policies
func (s *authService) RemovePolicy(ctx context.Context, req *auth.PolicyManagementRequest) (*auth.PolicyManagementResponse, error) {
	var removedCount int
	for _, rule := range req.GetRules() {
		ok, err := s.enforcer.RemovePolicy(
			rule.GetId(),
			rule.GetTargetAction(),
			rule.GetTargetResource(),
			rule.GetPlayerConditionExpression(),
			rule.GetServerConditionExpression(),
			rule.GetEffect(),
			fmt.Sprintf("%d", rule.GetPriority()),
		)
		if err != nil {
			log.Printf("Error removing policy %s: %v", rule.GetId(), err)
			return &auth.PolicyManagementResponse{Success: false, Message: fmt.Sprintf("Error removing policy %s: %v", rule.GetId(), err)}, nil
		}
		if ok {
			removedCount++
			log.Printf("Removed policy: %v", rule)
		} else {
			log.Printf("Policy not found: %v", rule)
		}
	}
	s.enforcer.LoadPolicy()
	return &auth.PolicyManagementResponse{Success: true, Message: fmt.Sprintf("Successfully removed %d policies", removedCount)}, nil
}

// ListPolicies implements the gRPC method to list all policies
func (s *authService) ListPolicies(ctx context.Context, req *emptypb.Empty) (*auth.PolicyManagementRequest, error) {
	policies, err := s.enforcer.GetPolicy()
	if err != nil {
		log.Printf("Error retrieving policies: %v", err)
		return nil, fmt.Errorf("error retrieving policies: %v", err)
	}
	var pbRules []*auth.PolicyRule

	for _, p := range policies {
		if len(p) == 7 { // Ensure the policy string has the correct number of fields
			priority, err := strconv.Atoi(p[6])
			if err != nil {
				log.Printf("Warning: Could not parse priority for policy %s: %v", p[0], err)
				priority = 0 // Default to 0 or handle error as appropriate
			}
			pbRules = append(pbRules, &auth.PolicyRule{
				Id:                        p[0],
				TargetAction:              p[1],
				TargetResource:            p[2],
				PlayerConditionExpression: p[3],
				ServerConditionExpression: p[4],
				Effect:                    p[5],
				Priority:                  int32(priority),
			})
		} else {
			log.Printf("Warning: Policy has unexpected number of fields: %v", p)
		}
	}
	return &auth.PolicyManagementRequest{Rules: pbRules}, nil
}

func main() {
	cfg := config.LoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- 2. Initialize Casbin Enforcer ---
	adapter := casbinredisadapter.NewAdpaterWithOption(
		casbinredisadapter.WithNetwork("tcp"),
		casbinredisadapter.WithAddress(cfg.ValkeyAddr),
		casbinredisadapter.WithPassword(cfg.ValkeyPassword),
		casbinredisadapter.WithKey(cfg.ValkeyKey),
	)
	enforcer, err := casbin.NewEnforcer("model.conf", adapter)
	if err != nil {
		log.Fatalf("Failed to create Casbin enforcer: %v", err)
	}
	err = enforcer.LoadPolicy()
	if err != nil {
		log.Fatalf("Failed to load policies from adapter: %v", err)
	}
	log.Println("Casbin policies loaded.")

	// --- 3. Initialize NATS Connection and Key-Value Store for Metadata ---
	natsOpts := []nats.Option{nats.Name("MinecraftAuthService")}
	if cfg.NATSUser != "" && cfg.NATSPassword != "" {
		natsOpts = append(natsOpts, nats.UserInfo(cfg.NATSUser, cfg.NATSPassword))
	}
	nc, err := nats.Connect(cfg.NATSAddr, natsOpts...)
	if err != nil {
		log.Fatalf("Could not connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Successfully connected to NATS!")

	// Bind to the KV store (assuming it already exists as a JetStream KV bucket named "metadata")
	// If you're running JetStream for the first time or need to create the bucket:
	js, jsErr := nc.JetStream()
	if jsErr != nil {
		log.Fatalf("Could not get JetStream context: %v", jsErr)
	}
	kv, err := js.KeyValue("metadata")
	if err != nil && err == nats.ErrBucketNotFound {
		log.Printf("NATS KV bucket 'metadata' not found, attempting to create.")
		kv, err = js.CreateKeyValue(&nats.KeyValueConfig{Bucket: "metadata"})
		if err != nil {
			log.Fatalf("Failed to create NATS KV bucket 'metadata': %v", err)
		}
	} else if err != nil {
		log.Fatalf("Failed to get NATS KV bucket 'metadata': %v", err)
	}
	log.Println("Successfully connected to NATS KV 'metadata' bucket!")

	// --- 4. Initialize and Start Metadata Cache ---
	metadataCache := cache.NewMetadataCache(kv, cfg.PlayerMetadataPrefix, cfg.ServerMetadataPrefix)
	ctx, cancelMain := context.WithCancel(context.Background()) // Use a different context for main app lifetime
	metadataCache.StartWatching(ctx)

	// --- 5. Start gRPC server ---
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	auth.RegisterAuthServiceServer(s, NewAuthService(enforcer, metadataCache))

	log.Printf("gRPC server listening on port %s", cfg.GRPCPort)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// --- Graceful Shutdown ---
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // Block until a signal is received

	log.Println("Shutting down server...")
	s.GracefulStop()
	cancelMain() // Stop NATS KV watchers
	log.Println("Server gracefully stopped.")
}
