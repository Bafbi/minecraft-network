package web

import (
	"fmt"
	"net/http"

	// "strings"

	"github.com/bafbi/minecraft-network/services/permissions-webapp/internal/core"
	"github.com/bafbi/minecraft-network/services/permissions-webapp/internal/web/components"
	"github.com/bafbi/minecraft-network/services/permissions-webapp/internal/web/views"
	"github.com/go-chi/chi/v5"
	// "github.com/go-logr/logr" // If you pass a logger
)

// var logger logr.Logger // Set this up

func RegisterRoutes(r *chi.Mux /*, log logr.Logger*/) {
	// logger = log
	r.Get("/", handleDashboard) // Or redirect to policies
	r.Get("/policies", handleListPolicies)
	r.Post("/policies", handleAddPolicy)
	r.Post("/policies/remove", handleRemovePolicy) // Using POST for simplicity with hx-vals
	r.Get("/policies/add-form", handleAddPolicyForm)

	// Serve static files (htmx.min.js)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./internal/static"))))
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/policies", http.StatusFound)
}

func handleListPolicies(w http.ResponseWriter, r *http.Request) {
	if core.Enforcer == nil {
		http.Error(w, "Casbin enforcer not initialized", http.StatusInternalServerError)
		return
	}
	rawPolicies, err := core.Enforcer.GetPolicy() // Gets 'p' policies
	if err != nil {
		http.Error(w, "Failed to get policies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var viewPolicies []components.PolicyRule
	for _, p := range rawPolicies {
		// Assuming p rules have at least 4 fields: sub, obj, act, eft
		if len(p) >= 4 {
			viewPolicies = append(viewPolicies, components.PolicyRule{
				SubLogic: p[0],
				ObjLogic: p[1],
				Action:   p[2],
				Effect:   p[3],
			})
		}
	}
	// If it's an HTMX request for just the table body, render only that
	// For now, always render the full page for simplicity
	views.PoliciesPage(viewPolicies).Render(r.Context(), w)
}

func handleAddPolicyForm(w http.ResponseWriter, r *http.Request) {
	views.AddPolicyForm().Render(r.Context(), w)
}

func handleAddPolicy(w http.ResponseWriter, r *http.Request) {
	if core.Enforcer == nil {
		http.Error(w, "Casbin enforcer not initialized", http.StatusInternalServerError)
		return
	}
	subLogic := r.FormValue("sub_logic")
	objLogic := r.FormValue("obj_logic")
	action := r.FormValue("action")
	effect := r.FormValue("effect")

	if subLogic == "" || objLogic == "" || action == "" || effect == "" {
		// Send an error back, possibly as an HTMX out-of-band swap to a notification area
		// For now, simple error
		http.Error(w, "All policy fields are required", http.StatusBadRequest)
		// You could also use OOB swap with HTMX:
		// w.Header().Set("HX-Retarget", "#notifications")
		// w.Header().Set("HX-Reswap", "innerHTML")
		// fmt.Fprint(w, "<div class='error'>All fields required</div>")
		return
	}

	rule := []string{subLogic, objLogic, action, effect}
	added, err := core.Enforcer.AddPolicy(rule)
	if err != nil {
		http.Error(w, "Failed to add policy: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !added {
		// Could mean policy already exists
		// Send a specific message back
		// For now, just re-render the row (or do nothing if it's a duplicate)
		// Let's assume for now we just want to show the new state.
		// The best HTMX way is to return the new row to be appended.
		components.PolicyRow(components.PolicyRule{SubLogic: subLogic, ObjLogic: objLogic, Action: action, Effect: effect}).Render(r.Context(), w)
		return // Or an appropriate message
	}

	if err := core.Enforcer.SavePolicy(); err != nil {
		http.Error(w, "Failed to save policy: "+err.Error(), http.StatusInternalServerError)
		return
	}
	core.PublishNatsPolicyUpdate() // You'll need this function in core/nats.go

	// Return the new row as an HTML fragment for HTMX to append
	components.PolicyRow(components.PolicyRule{SubLogic: subLogic, ObjLogic: objLogic, Action: action, Effect: effect}).Render(r.Context(), w)
}

func handleRemovePolicy(w http.ResponseWriter, r *http.Request) {
	if core.Enforcer == nil {
		http.Error(w, "Casbin enforcer not initialized", http.StatusInternalServerError)
		return
	}
	subLogic := r.FormValue("sub_logic")
	objLogic := r.FormValue("obj_logic")
	action := r.FormValue("action")
	effect := r.FormValue("effect")
	rule := []string{subLogic, objLogic, action, effect}

	removed, err := core.Enforcer.RemovePolicy(rule)
	if err != nil {
		http.Error(w, "Failed to remove policy: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !removed {
		// Policy not found, could send a 204 No Content or an error message
		// For HTMX, if the row is already gone due to optimistic UI or this is a confirmation,
		// an empty successful response might be fine.
		w.WriteHeader(http.StatusOK) // Or http.StatusNoContent if nothing to return
		fmt.Fprint(w, "<!-- Policy not found or already removed -->")
		return
	}

	if err := core.Enforcer.SavePolicy(); err != nil {
		http.Error(w, "Failed to save policy after removal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	core.PublishNatsPolicyUpdate()

	// HTMX expects an empty response on successful deletion if the target is removed
	// Or, if you are replacing a section, return the new content for that section.
	// Since hx-target="closest tr" hx-swap="outerHTML", an empty 200 OK will remove the row.
	w.WriteHeader(http.StatusOK)
}
