#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
HOST_DATA_PATH="${PROJECT_ROOT}/data" # Ensure this directory exists or is created if needed

# Create data directory if it doesn't exist
mkdir -p "$HOST_DATA_PATH"

# Create a temporary config file with the actual path injected
CONFIG_FILE="$(mktemp)"
# Export HOST_DATA_PATH so envsubst can use it
export HOST_DATA_PATH
envsubst < "${SCRIPT_DIR}/kind-config.yaml.in" > "${CONFIG_FILE}"
unset HOST_DATA_PATH # Unset after use

echo "Creating Kind cluster 'minecraft-net' with config: ${CONFIG_FILE}"
kind create cluster --name minecraft-net --config "${CONFIG_FILE}"

# Cleanup the temp file
rm "${CONFIG_FILE}"

kubectl config use-context kind-minecraft-net

# Activate local mode for building images for Kind
# This sets DEV_MODE=local and other relevant vars for the build scripts
source "${SCRIPT_DIR}/activate.sh" local
if [[ "$DEV_MODE" != "local" ]]; then
    echo "Error: Failed to activate local development mode for Kind setup."
    exit 1
fi

echo "Building and loading images into Kind..."
bash "${PROJECT_ROOT}/servers/lobby_minestom/scripts/build.sh"
bash "${PROJECT_ROOT}/servers/proxy_gate/scripts/build.sh"
bash "${PROJECT_ROOT}/services/permissions-webapp/scripts/build.sh" # Add webapp build

# Ensure Helm values file for dev exists
if [[ ! -f "$HELM_VALUES_FILE" ]]; then
    echo "Error: Helm values file for development ($HELM_VALUES_FILE) not found!"
    exit 1
fi

echo "Deploying Helm chart 'network' using values from $HELM_VALUES_FILE..."
helm dependency update "${PROJECT_ROOT}/charts/network"
helm upgrade --install network "${PROJECT_ROOT}/charts/network" -f "$HELM_VALUES_FILE" --namespace "$K8S_NAMESPACE" --create-namespace

echo "✅ Kind cluster 'minecraft-net' setup complete."
echo "   Proxy NodePort (if applicable): check 'kubectl get svc network-proxy -n $K8S_NAMESPACE'"
echo "   To use this environment: source scripts/activate.sh local"
