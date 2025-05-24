// services/permissions-editor/main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	authpb "github.com/bafbi/minecraft-network/services/permissions-checker/auth"
	"github.com/bafbi/minecraft-network/services/permissions-editor/templates"
)

// AppState holds common dependencies for handlers
type AppState struct {
	AuthClient           authpb.AuthServiceClient
	NATSKV               nats.KeyValue
	PlayerMetadataPrefix string
	ServerMetadataPrefix string
}

func main() {
	natsAddr := os.Getenv("NATS_ADDR")
	if natsAddr == "" {
		natsAddr = "nats://localhost:4222"
	}
	natsUser := os.Getenv("NATS_USER")
	natsPassword := os.Getenv("NATS_PASSWORD")

	natsOpts := []nats.Option{nats.Name("PermissionsEditor")}
	if natsUser != "" && natsPassword != "" {
		natsOpts = append(natsOpts, nats.UserInfo(natsUser, natsPassword))
	}

	nc, err := nats.Connect(natsAddr, natsOpts...)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Connected to NATS!")

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Error getting JetStream context: %v", err)
	}

	kv, err := js.KeyValue("metadata")
	if err != nil {
		log.Fatalf("Error getting KV bucket 'metadata': %v", err)
	}
	log.Println("Connected to NATS KV 'metadata' bucket!")

	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}

	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to gRPC service: %v", err)
	}
	defer conn.Close()
	authClient := authpb.NewAuthServiceClient(conn)
	log.Println("Connected to gRPC Permissions-Checker Service!")

	appState := &AppState{
		AuthClient:           authClient,
		NATSKV:               kv,
		PlayerMetadataPrefix: "player.metadata.",
		ServerMetadataPrefix: "server.metadata.",
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	r.Get("/", templ.Handler(templates.Base()))
	r.Get("/players", appState.listPlayersHandler)
	r.Get("/player/{uuid}", appState.getPlayerDetailHandler)
	r.Post("/player/{uuid}", appState.updatePlayerMetadataHandler)
	r.Get("/servers", appState.listServersHandler)
	r.Get("/server/{name}", appState.getServerDetailHandler)
	r.Post("/server/{name}", appState.updateServerMetadataHandler)
	r.Get("/policies", appState.listPoliciesHandler)
	r.Post("/policies", appState.addPolicyHandler)
	r.Delete("/policies", appState.deletePolicyHandler)

	log.Println("Web server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func render(w http.ResponseWriter, r *http.Request, component templ.Component) {
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to render component: %s", err.Error()), http.StatusInternalServerError)
	}
}
