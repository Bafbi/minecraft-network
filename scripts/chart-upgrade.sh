#!/bin/bash
set -e

# This script is intended to be run in a Bash environment.
# It relies on environment variables set by sourcing 'activate.fish' (for Fish users)
# or 'activate.sh' (for Bash users) in the parent shell.

# Check for necessary environment variables
if [[ -z "$DEV_MODE" || -z "$K8S_CONTEXT_TARGET" || -z "$HELM_VALUES_FILE" || -z "$K8S_NAMESPACE" ]]; then
  echo "Error: Environment not activated."
  echo "Please run 'source scripts/activate.fish [local|remote-dev|remote-prod]' (for Fish shell)"
  echo "or 'source scripts/activate.sh [local|remote-dev|remote-prod]' (for Bash shell) first."
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "--------------------------------------------------"
echo "Helm Chart Upgrade/Install"
echo "--------------------------------------------------"
echo "Mode:               $DEV_MODE"
echo "Target K8s Context: $K8S_CONTEXT_TARGET"
echo "Target K8s Ns:    $K8S_NAMESPACE"
echo "Helm Values File:   $HELM_VALUES_FILE"
echo "Chart Path:         ${PROJECT_ROOT}/charts/network"
echo "--------------------------------------------------"

# No need to switch context here if activate.fish/activate.sh already prompted
# and potentially switched it. However, explicitly setting it ensures the script
# uses the intended context, even if the user didn't confirm the switch in activate.
# This can be seen as a safety measure or a bit redundant depending on preference.
# For now, let's keep it to ensure the script is explicit.
echo ""
echo "Ensuring Kubernetes context is set to: $K8S_CONTEXT_TARGET"
if ! kubectl config use-context "$K8S_CONTEXT_TARGET"; then
    echo "Error: Failed to switch Kubernetes context to '$K8S_CONTEXT_TARGET'."
    echo "Please ensure the context exists and you have access."
    exit 1
fi
echo "Successfully set Kubernetes context to '$K8S_CONTEXT_TARGET'."
echo ""


# Ensure the target namespace exists, or Helm will try with --create-namespace
# It's good practice to check/create it if your RBAC doesn't allow Helm to create namespaces.
# For simplicity, relying on Helm's --create-namespace for now.

# echo "Updating Helm dependencies for chart at '${PROJECT_ROOT}/charts/network'..."
# if ! helm dependency update "${PROJECT_ROOT}/charts/network"; then
#     echo "Error: Failed to update Helm dependencies."
#     exit 1
# fi
# echo "Helm dependencies updated."
# echo ""

echo "Upgrading/Installing Helm chart 'network'..."
echo "Command: helm upgrade --install network \"${PROJECT_ROOT}/charts/network\" -f \"$HELM_VALUES_FILE\" --namespace \"$K8S_NAMESPACE\" --create-namespace"
if helm upgrade --install network \
  "${PROJECT_ROOT}/charts/network" \
  -f "$HELM_VALUES_FILE" \
  --namespace "$K8S_NAMESPACE" \
  --create-namespace; then
  echo "✅ Helm chart 'network' upgrade/install process initiated successfully."
else
  echo "❌ Error during Helm chart 'network' upgrade/install."
  exit 1
fi

# You might want to add a `helm status network -n $K8S_NAMESPACE` here
# or `kubectl get pods -n $K8S_NAMESPACE -l app.kubernetes.io/instance=network`
# to show the status after the upgrade.
echo ""
echo "Run 'kubectl get pods -n $K8S_NAMESPACE -w' to watch pod statuses."
