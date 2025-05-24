package permissions

import (
	"errors" // Standard Go errors package
	"fmt"
	"strings"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/players"
	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/servers"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/govaluate"
	"github.com/go-logr/logr"
	"go.minekube.com/gate/pkg/util/uuid"
)

// --- Getter Interfaces (remain the same) ---

type CasbinEnforcerGetter func() *casbin.Enforcer

var (
	getLog      func() logr.Logger
	getEnforcer CasbinEnforcerGetter
)

func InitCustomFunctionDeps(
	loggerFunc func() logr.Logger,
	ceg CasbinEnforcerGetter,
) {
	getLog = loggerFunc
	getEnforcer = ceg
}

// EvalSubjectAttributes evaluates if the subject (player) matches the given govaluate expression.
// args[0]: player_uuid (string) - r.sub
// args[1]: p_sub_eval_logic (string) - The govaluate expression from the policy
func EvalSubjectAttributes(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("EvalSubjectAttributes: expected 2 arguments, got %d", len(args))
	}
	playerUUIDStr, ok := args[0].(string)
	if !ok {
		return false, errors.New("EvalSubjectAttributes: first argument (playerUUID) must be a string")
	}
	evalLogic, ok := args[1].(string)
	if !ok {
		return false, errors.New("EvalSubjectAttributes: second argument (evalLogic) must be a string")
	}

	log := getLog().WithValues("func", "EvalSubjectAttributes", "playerUUID", playerUUIDStr, "logic", evalLogic)

	if evalLogic == "true" || evalLogic == "*" { // Handle simple "always true" or wildcard
		return true, nil
	}
	if evalLogic == "false" {
		return false, nil
	}

	playerID, err := uuid.Parse(playerUUIDStr)
	if err != nil {
		log.Error(err, "Invalid player UUID format")
		return false, fmt.Errorf("invalid player UUID: %s", playerUUIDStr)
	}

	meta, found := players.GetMetadataByUUID(playerID)
	var labels, annotations map[string]string
	if found {
		labels = meta.Labels
		annotations = meta.Annotations
	} else {
		log.V(1).Info("Player metadata not found for evaluation")
		labels = make(map[string]string)
		annotations = make(map[string]string)
	}

	enforcer := getEnforcer()
	if enforcer == nil {
		log.Error(nil, "Casbin enforcer is not available")
		return false, errors.New("enforcer not available")
	}

	parameters := map[string]interface{}{
		"uuid":        playerUUIDStr, // The player's UUID
		"labels":      labels,        // Player's labels map
		"annotations": annotations,   // Player's annotations map
		// Add a helper function for group checks within govaluate expressions
		"inGroup": func(groupName string) bool {
			// Note: govaluate function arguments are passed as []interface{}
			// This is a simplified example; real group check might need more robust arg handling.
			// For HasRoleForUser, we need playerUUIDStr and groupName.
			// The `groupName` comes from the expression, e.g., `inGroup('admin')`
			res, errGH := enforcer.HasRoleForUser(playerUUIDStr, groupName)
			if errGH != nil {
				log.Error(errGH, "Error checking Casbin group membership in govaluate", "group", groupName)
				return false
			}
			return res
		},
	}

	expression, err := govaluate.NewEvaluableExpression(evalLogic)
	if err != nil {
		log.Error(err, "Failed to parse subject evaluation logic as govaluate expression")
		return false, fmt.Errorf("failed to parse subject eval logic '%s': %w", evalLogic, err)
	}

	result, err := expression.Evaluate(parameters)
	if err != nil {
		// This can happen if the expression references a field not in parameters,
		// or if types are mismatched, or if a label/annotation is missing and accessed directly.
		log.V(1).Info("Failed to evaluate subject logic expression", "error", err.Error())
		return false, nil // Treat evaluation errors as a non-match (false)
	}

	if boolResult, ok := result.(bool); ok {
		return boolResult, nil
	}

	log.V(1).Info("Subject logic expression did not return a boolean", "result", result)
	return false, fmt.Errorf("subject eval logic '%s' did not return a boolean", evalLogic)
}

// EvalObjectAttributes evaluates if the object matches the given govaluate expression.
// args[0]: r_sub (string) - The requesting subject's UUID. Needed if obj logic refers to subject.
// args[1]: r_obj (string) - The object identifier (e.g., "server:lobby-1", "command:kick")
// args[2]: p_obj_eval_logic (string) - The govaluate expression from the policy
func EvalObjectAttributes(args ...interface{}) (interface{}, error) {
	if len(args) != 3 {
		return false, fmt.Errorf("EvalObjectAttributes: expected 3 arguments, got %d", len(args))
	}
	rSubUUIDStr, ok := args[0].(string)
	if !ok {
		return false, errors.New("EvalObjectAttributes: first argument (r_sub_uuid) must be a string")
	}
	objIdentifier, ok := args[1].(string)
	if !ok {
		return false, errors.New("EvalObjectAttributes: second argument (objIdentifier) must be a string")
	}
	evalLogic, ok := args[2].(string)
	if !ok {
		return false, errors.New("EvalObjectAttributes: third argument (evalLogic) must be a string")
	}

	log := getLog().WithValues("func", "EvalObjectAttributes", "r_sub", rSubUUIDStr, "object", objIdentifier, "logic", evalLogic)

	if evalLogic == "true" || evalLogic == "*" { // Handle simple "always true" or wildcard
		return true, nil
	}
	if evalLogic == "false" {
		return false, nil
	}

	objParts := strings.SplitN(objIdentifier, ":", 2)
	objType := objParts[0]
	objName := ""
	if len(objParts) > 1 {
		objName = objParts[1]
	} else {
		// If no colon, objType is the identifier itself, objName remains empty.
		// This could be for simple object names like "global_chat".
	}

	parameters := map[string]interface{}{
		"type":       objType,     // The object's type (e.g., "server", "command")
		"name":       objName,     // The object's name (e.g., "lobby-1", "kick")
		"r_sub_uuid": rSubUUIDStr, // Requesting subject's UUID, for policies like "labels.owner_uuid == r_sub_uuid"
		// Initialize metadata maps to empty to prevent nil access if not found
		"labels":      make(map[string]string),
		"annotations": make(map[string]string),
	}

	// Populate object-specific metadata
	switch objType {
	case "server":
		if objName != "" {
			meta, found := servers.GetMetadataByName(objName)
			if found {
				parameters["labels"] = meta.Labels
				parameters["annotations"] = meta.Annotations
			}
		}
	// Other object types can be handled here if needed
	default:
		log.V(1).Info("Unknown object type for metadata fetching", "objType", objType)
	}

	expression, err := govaluate.NewEvaluableExpression(evalLogic)
	if err != nil {
		log.Error(err, "Failed to parse object evaluation logic as govaluate expression")
		return false, fmt.Errorf("failed to parse object eval logic '%s': %w", evalLogic, err)
	}

	result, err := expression.Evaluate(parameters)
	if err != nil {
		log.V(1).Info("Failed to evaluate object logic expression", "error", err.Error())
		return false, nil // Treat evaluation errors as a non-match
	}

	if boolResult, ok := result.(bool); ok {
		return boolResult, nil
	}

	log.V(1).Info("Object logic expression did not return a boolean", "result", result)
	return false, fmt.Errorf("object eval logic '%s' did not return a boolean", evalLogic)
}
