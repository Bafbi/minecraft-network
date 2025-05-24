#!/bin/bash

# Script to activate development environment mode (local, remote-dev, remote-prod)

# Usage: source scripts/activate.sh [local|remote-dev|remote-prod]
# Default is local if no argument is provided.

# Function to ask for confirmation
confirm_action() {
  while true; do
    read -r -p "$1 [y/N]: " response
    case "$response" in
      [yY][eE][sS]|[yY])
        return 0 # Yes
        ;; # <-- MISSING SEMICOLONS HERE
      [nN][oO]|[nN]|"") # Default to No if Enter is pressed
        return 1 # No
        ;; # <-- AND HERE
      *)
        echo "Invalid input. Please answer yes (y) or no (n)."
        ;; # <-- AND HERE
    esac
  done
}

# Clear previous settings
unset DEV_MODE
unset IMAGE_REGISTRY
unset K8S_CONTEXT_TARGET # Renamed to avoid confusion with current context
unset K8S_NAMESPACE
unset HELM_VALUES_FILE

# --- Configuration Section ---
# You should customize these values for your environments

# Local (Kind) Configuration
LOCAL_IMAGE_REGISTRY_PREFIX="" # For Kind, images are loaded directly
LOCAL_K8S_CONTEXT="kind-minecraft-net"
LOCAL_K8S_NAMESPACE="default" # Or your preferred local namespace

# Remote Development Configuration
REMOTE_DEV_IMAGE_REGISTRY_PREFIX="ghcr.io/bafbi" # Your GHCR prefix
REMOTE_DEV_K8S_CONTEXT="your-remote-dev-cluster-context" # !!! CHANGE THIS !!!
REMOTE_DEV_K8S_NAMESPACE="minecraft-dev"                 # !!! CHANGE THIS !!!

# Remote Production Configuration
REMOTE_PROD_IMAGE_REGISTRY_PREFIX="ghcr.io/bafbi" # Your GHCR prefix
REMOTE_PROD_K8S_CONTEXT="your-remote-prod-cluster-context" # !!! CHANGE THIS !!!
REMOTE_PROD_K8S_NAMESPACE="minecraft-prod"                 # !!! CHANGE THIS !!!

# --- End Configuration Section ---

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Helm values files (relative to project root)
LOCAL_HELM_VALUES_FILE_REL="values/dev-values.yaml"
REMOTE_DEV_HELM_VALUES_FILE_REL="values/dev-values.yaml" # Or a specific remote-dev-values.yaml
REMOTE_PROD_HELM_VALUES_FILE_REL="values/prod-values.yaml"

# Default mode
DEFAULT_MODE="local"
MODE=${1:-$DEFAULT_MODE}

case "$MODE" in
  "local")
    export DEV_MODE="local"
    export IMAGE_REGISTRY="$LOCAL_IMAGE_REGISTRY_PREFIX"
    export K8S_CONTEXT_TARGET="$LOCAL_K8S_CONTEXT"
    export K8S_NAMESPACE="$LOCAL_K8S_NAMESPACE"
    export HELM_VALUES_FILE="${PROJECT_ROOT}/${LOCAL_HELM_VALUES_FILE_REL}"
    ACTIVATION_MESSAGE="Activated LOCAL development mode."
    ;;
  "remote-dev")
    export DEV_MODE="remote-dev"
    export IMAGE_REGISTRY="$REMOTE_DEV_IMAGE_REGISTRY_PREFIX"
    export K8S_CONTEXT_TARGET="$REMOTE_DEV_K8S_CONTEXT"
    export K8S_NAMESPACE="$REMOTE_DEV_K8S_NAMESPACE"
    export HELM_VALUES_FILE="${PROJECT_ROOT}/${REMOTE_DEV_HELM_VALUES_FILE_REL}"
    ACTIVATION_MESSAGE="Activated REMOTE-DEV mode."
    ADDITIONAL_INFO="Ensure you are logged into container registry: $IMAGE_REGISTRY"
    ;;
  "remote-prod")
    export DEV_MODE="remote-prod"
    export IMAGE_REGISTRY="$REMOTE_PROD_IMAGE_REGISTRY_PREFIX"
    export K8S_CONTEXT_TARGET="$REMOTE_PROD_K8S_CONTEXT"
    export K8S_NAMESPACE="$REMOTE_PROD_K8S_NAMESPACE"
    export HELM_VALUES_FILE="${PROJECT_ROOT}/${REMOTE_PROD_HELM_VALUES_FILE_REL}"
    ACTIVATION_MESSAGE="Activated REMOTE-PROD mode."
    ADDITIONAL_INFO="Ensure you are logged into container registry: $IMAGE_REGISTRY"
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo "!!! WARNING: Production mode activated. Use caution. !!!"
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    ;;
  *)
    echo "Invalid mode: $MODE. Use 'local', 'remote-dev', or 'remote-prod'."
    return 1 # Use return for sourced scripts
    ;;
esac

echo "$ACTIVATION_MESSAGE"
if [[ -n "$ADDITIONAL_INFO" ]]; then
  echo "$ADDITIONAL_INFO"
fi

echo "--------------------------------------------------"
echo "DEV_MODE:           $DEV_MODE"
echo "IMAGE_REGISTRY:     $IMAGE_REGISTRY"
echo "K8S_CONTEXT_TARGET: $K8S_CONTEXT_TARGET"
echo "K8S_NAMESPACE:      $K8S_NAMESPACE"
echo "HELM_VALUES_FILE:   $HELM_VALUES_FILE"
echo "--------------------------------------------------"

# Check if HELM_VALUES_FILE exists
if [[ ! -f "$HELM_VALUES_FILE" ]]; then
    echo "WARNING: Helm values file ($HELM_VALUES_FILE) not found!"
    echo "You may need to create it or adjust the path in activate.sh."
fi


# Attempt to set Kubernetes context
CURRENT_K8S_CONTEXT=$(kubectl config current-context 2>/dev/null)

if [[ "$CURRENT_K8S_CONTEXT" != "$K8S_CONTEXT_TARGET" ]]; then
  echo ""
  echo "The target Kubernetes context is '$K8S_CONTEXT_TARGET'."
  if [[ -n "$CURRENT_K8S_CONTEXT" ]]; then
    echo "Your current context is '$CURRENT_K8S_CONTEXT'."
  else
    echo "You do not seem to have a current Kubernetes context set."
  fi

  if confirm_action "Do you want to switch to context '$K8S_CONTEXT_TARGET' now?"; then
    if kubectl config use-context "$K8S_CONTEXT_TARGET"; then
      echo "Successfully switched to Kubernetes context '$K8S_CONTEXT_TARGET'."
    else
      echo "Failed to switch Kubernetes context. Please do it manually: kubectl config use-context $K8S_CONTEXT_TARGET"
    fi
  else
    echo "Kubernetes context not changed. Remember to set it manually if needed: kubectl config use-context $K8S_CONTEXT_TARGET"
  fi
else
  echo ""
  echo "Kubernetes context '$K8S_CONTEXT_TARGET' is already active."
fi

echo ""
echo "To deactivate this mode, start a new shell or unset the environment variables."
