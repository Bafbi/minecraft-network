package permissions

import (
	"context" // Keep context if needed for future adapter/enforcer options
	"errors"
	"fmt"
	"os"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/util" // For KeyMatchFunc
	redisadapter "github.com/casbin/redis-adapter/v3"
	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	// "go.minekube.com/gate/pkg/util/uuid" // Not directly needed here, but custom_functions.go uses it
)

const (
	// NATS subject for broadcasting Casbin policy updates
	casbinPolicyUpdateSubject = "network.casbin.policy.updated"
)

var (
	// enforcerInstance is the singleton Casbin enforcer.
	// It's private to this package; access via GetEnforcer().
	enforcerInstance *casbin.Enforcer
	// Store the NATS connection if PublishPolicyUpdate needs it and it's not passed every time.
	// For now, let's pass it to PublishPolicyUpdate.
)

// InitCasbin initializes the Casbin enforcer, adapter, custom functions,
// loads policies, and subscribes to NATS for policy updates.
// It takes PlayerMetadataGetter and ServerMetadataGetter from custom_functions.go (same package).
func InitCasbin(
	_ context.Context, // Keep context for future use, e.g., if adapter needs it
	log logr.Logger,
	nc *nats.Conn,
) error {
	log.Info("Initializing Casbin service...")

	modelPath := os.Getenv("CASBIN_MODEL_PATH")
	if modelPath == "" {
		return fmt.Errorf("CASBIN_MODEL_PATH environment variable not set")
	}
	log.Info("Casbin model path from env", "path", modelPath)

	valkeyHost := os.Getenv("VALKEY_HOST")
	if valkeyHost == "" {
		return fmt.Errorf("VALKEY_HOST environment variable not set")
	}
	valkeyPort := os.Getenv("VALKEY_PORT")
	if valkeyPort == "" {
		valkeyPort = "6379"
		log.Info("VALKEY_PORT not set, using default", "port", valkeyPort)
	}
	valkeyPassword := os.Getenv("VALKEY_PASSWORD")

	redisAddr := fmt.Sprintf("%s:%s", valkeyHost, valkeyPort)
	log.Info("Connecting to Valkey/Redis for Casbin", "address", redisAddr, "password_set", valkeyPassword != "")

	adapter, errAdapter := redisadapter.NewAdapterWithOption(
		redisadapter.WithNetwork("tcp"),
		redisadapter.WithAddress(redisAddr),
		redisadapter.WithPassword(valkeyPassword),
	)
	if errAdapter != nil {
		return fmt.Errorf("failed to create Casbin Redis adapter: %w", errAdapter)
	}
	log.Info("Casbin Redis adapter created")

	var err error
	enforcerInstance, err = casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}
	log.Info("Casbin enforcer created")

	// Initialize dependencies for custom Casbin functions (from custom_functions.go)
	InitCustomFunctionDeps(
		func() logr.Logger { return log.WithName("CasbinCustomFunc") },
		func() *casbin.Enforcer { return enforcerInstance },
	)
	log.Info("Custom Casbin function dependencies initialized")

	// Register custom functions with Casbin
	enforcerInstance.AddFunction("eval_subject_attributes", EvalSubjectAttributes)
	enforcerInstance.AddFunction("eval_object_attributes", EvalObjectAttributes)
	enforcerInstance.AddFunction("keyMatch", util.KeyMatchFunc)
	log.Info("Custom Casbin functions registered")

	if err = enforcerInstance.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to load initial Casbin policies: %w", err)
	}
	log.Info("Casbin policies loaded successfully from Redis")

	if nc != nil {
		_, errSub := nc.Subscribe(casbinPolicyUpdateSubject, func(msg *nats.Msg) {
			log.Info("Received Casbin policy update notification via NATS, reloading policies",
				"subject", msg.Subject)
			if enforcerInstance == nil {
				log.Error(nil, "Enforcer instance is nil, cannot reload policies.")
				return
			}
			if errReload := enforcerInstance.LoadPolicy(); errReload != nil {
				log.Error(errReload, "Error reloading Casbin policies after NATS notification")
			} else {
				log.Info("Successfully reloaded Casbin policies after NATS notification")
			}
		})
		if errSub != nil {
			// Log as warning, don't make it fatal for plugin init
			log.Error(errSub, "Failed to subscribe to Casbin policy updates via NATS. Live policy updates may not be received.")
		} else {
			log.Info("Subscribed to Casbin policy updates via NATS", "subject", casbinPolicyUpdateSubject)
		}
	} else {
		log.Error(nil, "NATS connection is nil, cannot subscribe to Casbin policy updates")
	}

	log.Info("Casbin service initialized successfully.")
	return nil
}

// GetEnforcer returns the global Casbin enforcer instance.
// Ensure InitCasbin has been called and was successful.
func GetEnforcer() *casbin.Enforcer {
	return enforcerInstance
}

// PublishPolicyUpdate sends a notification via NATS that Casbin policies might have changed.
func PublishPolicyUpdate(log logr.Logger, nc *nats.Conn) {
	if nc == nil {
		log.Error(nil, "NATS connection is nil, cannot publish Casbin policy update")
		return
	}
	if err := nc.Publish(casbinPolicyUpdateSubject, []byte("reload_policies")); err != nil {
		log.Error(err, "Failed to publish Casbin policy update notification to NATS")
	} else {
		log.V(1).Info("Published Casbin policy update notification to NATS")
	}
}

// HasPermission checks if a subject (player ID string) has a specific permission.
func HasPermission(subjectID string, objectResource string, action string, log logr.Logger) (bool, error) {
	e := GetEnforcer()
	if e == nil {
		// Log with the passed-in logger for better context
		log.Error(nil, "Casbin enforcer not initialized, denying permission check.",
			"subject", subjectID, "object", objectResource, "action", action)
		return false, errors.New("casbin enforcer not initialized")
	}

	allowed, err := e.Enforce(subjectID, objectResource, action)
	if err != nil {
		log.Error(err, "Error during Casbin Enforce check",
			"subject", subjectID, "object", objectResource, "action", action)
		return false, err
	}

	// Logging verbosity can be controlled by the logger's configuration
	if !allowed {
		log.V(1).Info("Permission denied by Casbin",
			"subject", subjectID, "object", objectResource, "action", action)
	} else {
		log.V(1).Info("Permission allowed by Casbin",
			"subject", subjectID, "object", objectResource, "action", action)
	}
	return allowed, nil
}
