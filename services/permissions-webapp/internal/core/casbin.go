package core

import (
	"fmt"
	// "os"

	"github.com/bafbi/minecraft-network/services/permissions-webapp/internal/config" // Your config package
	"github.com/casbin/casbin/v2"
	redisadapter "github.com/casbin/redis-adapter/v3"
	// "github.com/go-logr/logr" // If you use a logger here
)

var Enforcer *casbin.Enforcer

func InitCasbin(cfg *config.AppConfig /*, logger logr.Logger*/) error {
	// modelPath := os.Getenv("CASBIN_MODEL_PATH_WEBAPP") // Or get from cfg
	// For a webapp, the model is usually not needed to *manipulate* policies,
	// but it's good if it *knows* the model structure for validation or UI.
	// However, the Casbin enforcer itself needs the model to load/save policies correctly
	// if it's going to validate them against the model structure.
	// Let's assume the webapp also has access to the model file or its content.
	// For simplicity, if the webapp *only* adds/removes raw policy strings and relies
	// on the Gate proxy to validate them against the model upon load, it might not need the model.
	// But it's safer if it uses the same model.
	// Let's assume model path is configured.

	adapter, errAdapter := redisadapter.NewAdapterWithPassword("tcp", fmt.Sprintf("%s:%s", cfg.ValkeyHost, cfg.ValkeyPort), cfg.ValkeyPassword)
	if errAdapter != nil {
		return fmt.Errorf("webapp: failed to create Casbin Redis adapter: %w", errAdapter)
	}

	// The webapp needs the model file to correctly interpret and save policies,
	// especially if it's doing more than just raw string manipulation.
	// This model file should be the same one used by your Gate proxy.
	// You can mount it via ConfigMap to the webapp pod as well.
	var err error
	Enforcer, err = casbin.NewEnforcer(cfg.CasbinModelPath, adapter)
	if err != nil {
		return fmt.Errorf("webapp: failed to create Casbin enforcer: %w", err)
	}

	err = Enforcer.LoadPolicy()
	if err != nil {
		return fmt.Errorf("webapp: failed to load Casbin policies: %w", err)
	}
	// logger.Info("Webapp Casbin enforcer initialized and policies loaded")
	return nil
}
