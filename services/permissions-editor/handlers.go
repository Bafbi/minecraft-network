// services/permissions-editor/handlers.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/go-chi/chi/v5"

	// IMPORTANT: Corrected module name
	authpb "github.com/bafbi/minecraft-network/services/permissions-checker/auth"
	// IMPORTANT: Corrected module name
	"github.com/bafbi/minecraft-network/services/permissions-editor/templates"
)

func readJSONBody(r *http.Request, target interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	if len(body) == 0 {
		return fmt.Errorf("request body is empty")
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

func (s *AppState) listPlayersHandler(w http.ResponseWriter, r *http.Request) {
	keys, err := s.NATSKV.Keys(s.PlayerMetadataPrefix + ">")
	if err != nil {
		http.Error(w, "Failed to list player keys: "+err.Error(), http.StatusInternalServerError)
		return
	}

	playerUUIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		playerUUIDs = append(playerUUIDs, strings.TrimPrefix(key, s.PlayerMetadataPrefix))
	}

	render(w, r, templates.PlayersList(playerUUIDs))
}

func (s *AppState) getPlayerDetailHandler(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		http.Error(w, "Player UUID not provided in URL", http.StatusBadRequest)
		return
	}

	entry, err := s.NATSKV.Get(s.PlayerMetadataPrefix + uuid)
	if err != nil {
		if err == nats.ErrKeyNotFound {
			http.Error(w, "Player not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get player metadata: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(entry.Value(), &metadata); err != nil {
		http.Error(w, "Failed to unmarshal player metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	render(w, r, templates.PlayerDetail(uuid, metadata))
}

func (s *AppState) updatePlayerMetadataHandler(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		http.Error(w, "Player UUID not provided in URL", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	updatedMeta := make(map[string]interface{})
	metadataInput := r.FormValue("metadata_input")

	lines := strings.Split(metadataInput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			valStr := strings.TrimSpace(parts[1])

			if b, err := parseBool(valStr); err == nil {
				updatedMeta[key] = b
			} else if i, err := parseInt(valStr); err == nil {
				updatedMeta[key] = i
			} else if f, err := parseFloat(valStr); err == nil {
				updatedMeta[key] = f
			} else if strings.Contains(valStr, ",") {
				updatedMeta[key] = strings.Split(valStr, ",")
			} else {
				updatedMeta[key] = valStr
			}
		} else {
			log.Printf("Skipping malformed metadata line: %s", line)
		}
	}

	updatedBytes, err := json.Marshal(updatedMeta)
	if err != nil {
		http.Error(w, "Failed to marshal updated metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := s.NATSKV.Put(context.Background(), s.PlayerMetadataPrefix+uuid, updatedBytes); err != nil {
		http.Error(w, "Failed to update player metadata in NATS KV: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/player/%s", uuid), http.StatusSeeOther)
}

func (s *AppState) listServersHandler(w http.ResponseWriter, r *http.Request) {
	keys, err := s.NATSKV.Keys(s.ServerMetadataPrefix + ">")
	if err != nil {
		http.Error(w, "Failed to list server keys: "+err.Error(), http.StatusInternalServerError)
		return
	}

	serverNames := make([]string, 0, len(keys))
	for _, key := range keys {
		serverNames = append(serverNames, strings.TrimPrefix(key, s.ServerMetadataPrefix))
	}

	render(w, r, templates.ServersList(serverNames))
}

func (s *AppState) getServerDetailHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "Server name not provided in URL", http.StatusBadRequest)
		return
	}

	entry, err := s.NATSKV.Get(s.ServerMetadataPrefix + name)
	if err != nil {
		if err == nats.ErrKeyNotFound {
			http.Error(w, "Server not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get server metadata: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(entry.Value(), &metadata); err != nil {
		http.Error(w, "Failed to unmarshal server metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	render(w, r, templates.ServerDetail(name, metadata))
}

func (s *AppState) updateServerMetadataHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "Server name not provided in URL", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	updatedMeta := make(map[string]interface{})
	metadataInput := r.FormValue("metadata_input")

	lines := strings.Split(metadataInput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			valStr := strings.TrimSpace(parts[1])

			if b, err := parseBool(valStr); err == nil {
				updatedMeta[key] = b
			} else if i, err := parseInt(valStr); err == nil {
				updatedMeta[key] = i
			} else if f, err := parseFloat(valStr); err == nil {
				updatedMeta[key] = f
			} else if strings.Contains(valStr, ",") {
				updatedMeta[key] = strings.Split(valStr, ",")
			} else {
				updatedMeta[key] = valStr
			}
		} else {
			log.Printf("Skipping malformed metadata line: %s", line)
		}
	}

	updatedBytes, err := json.Marshal(updatedMeta)
	if err != nil {
		http.Error(w, "Failed to marshal updated metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := s.NATSKV.Put(s.ServerMetadataPrefix+name, updatedBytes); err != nil {
		http.Error(w, "Failed to update server metadata in NATS KV: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/server/%s", name), http.StatusSeeOther)
}

func (s *AppState) listPoliciesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.AuthClient.ListPolicies(ctx, &emptypb.Empty{})
	if err != nil {
		http.Error(w, "Failed to list policies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	render(w, r, templates.Policies(resp.GetRules()))
}

func (s *AppState) addPolicyHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	priority, err := parseInt(r.FormValue("priority"))
	if err != nil {
		http.Error(w, "Invalid priority: "+err.Error(), http.StatusBadRequest)
		return
	}

	newPolicy := &authpb.PolicyRule{
		Id:                        r.FormValue("id"),
		TargetAction:              r.FormValue("targetAction"),
		TargetResource:            r.FormValue("targetResource"),
		PlayerConditionExpression: r.FormValue("playerConditionExpression"),
		ServerConditionExpression: r.FormValue("serverConditionExpression"),
		Effect:                    r.FormValue("effect"),
		Priority:                  int32(priority),
	}

	if newPolicy.Id == "" || newPolicy.TargetAction == "" || newPolicy.TargetResource == "" || newPolicy.Effect == "" {
		http.Error(w, "All fields (ID, Target Action, Target Resource, Effect) are required.", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.AuthClient.AddPolicy(ctx, &authpb.PolicyManagementRequest{Rules: []*authpb.PolicyRule{newPolicy}})
	if err != nil {
		http.Error(w, "Failed to add policy via gRPC: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !resp.GetSuccess() {
		http.Error(w, "Failed to add policy: "+resp.GetMessage(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/policies", http.StatusSeeOther)
}

func (s *AppState) deletePolicyHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	priority, err := parseInt(r.FormValue("priority"))
	if err != nil {
		http.Error(w, "Invalid priority for delete: "+err.Error(), http.StatusBadRequest)
		return
	}

	policyToDelete := &authpb.PolicyRule{
		Id:                        r.FormValue("id"),
		TargetAction:              r.FormValue("targetAction"),
		TargetResource:            r.FormValue("targetResource"),
		PlayerConditionExpression: r.FormValue("playerConditionExpression"),
		ServerConditionExpression: r.FormValue("serverConditionExpression"),
		Effect:                    r.FormValue("effect"),
		Priority:                  int32(priority),
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.AuthClient.RemovePolicy(ctx, &authpb.PolicyManagementRequest{Rules: []*authpb.PolicyRule{policyToDelete}})
	if err != nil {
		http.Error(w, "Failed to delete policy via gRPC: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !resp.GetSuccess() {
		http.Error(w, "Failed to delete policy: "+resp.GetMessage(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func parseBool(s string) (bool, error) {
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}
	return false, fmt.Errorf("invalid boolean: %s", s)
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
