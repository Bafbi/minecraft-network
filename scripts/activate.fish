# Script to activate development environment mode (local, remote-dev, remote-prod) for Fish shell

# Usage: source scripts/activate.fish [local|remote-dev|remote-prod]
# Default is local if no argument is provided.

# Function to ask for confirmation
function confirm_action_fish
    while true
        # Fish's read command is a bit different
        read -P "$argv[1] [y/N]: " -l response
        switch "$response"
            case Y y Yes yes
                return 0 # Yes
            case N n No no "" # Default to No if Enter is pressed
                return 1 # No
            case '*'
                echo "Invalid input. Please answer yes (y) or no (n)."
        end
    end
end

# Clear previous settings (use -e to erase, -gx for global exported)
set -e DEV_MODE
set -e IMAGE_REGISTRY
set -e K8S_CONTEXT_TARGET
set -e K8S_NAMESPACE
set -e HELM_VALUES_FILE

# --- Configuration Section ---
# You should customize these values for your environments

# Local (Kind) Configuration
set -l LOCAL_IMAGE_REGISTRY_PREFIX ""
set -l LOCAL_K8S_CONTEXT "kind-minecraft-net"
set -l LOCAL_K8S_NAMESPACE "default"

# Remote Development Configuration
set -l REMOTE_DEV_IMAGE_REGISTRY_PREFIX "ghcr.io/bafbi"
set -l REMOTE_DEV_K8S_CONTEXT "your-remote-dev-cluster-context" # !!! CHANGE THIS !!!
set -l REMOTE_DEV_K8S_NAMESPACE "minecraft-dev"                 # !!! CHANGE THIS !!!

# Remote Production Configuration
set -l REMOTE_PROD_IMAGE_REGISTRY_PREFIX "ghcr.io/bafbi"
set -l REMOTE_PROD_K8S_CONTEXT "your-remote-prod-cluster-context" # !!! CHANGE THIS !!!
set -l REMOTE_PROD_K8S_NAMESPACE "minecraft-prod"                 # !!! CHANGE THIS !!!

# --- End Configuration Section ---

# Get script directory and project root
set -l SCRIPT_DIR (cd (dirname (status -f)); and pwd)
set -l PROJECT_ROOT (cd "$SCRIPT_DIR/.."; and pwd)

# Helm values files (relative to project root)
set -l LOCAL_HELM_VALUES_FILE_REL "values/dev-values.yaml"
set -l REMOTE_DEV_HELM_VALUES_FILE_REL "values/dev-values.yaml"
set -l REMOTE_PROD_HELM_VALUES_FILE_REL "values/prod-values.yaml"

# Default mode
set -l DEFAULT_MODE "local"
set -l MODE "$argv[1]"
if test -z "$MODE"
    set MODE "$DEFAULT_MODE"
end

# Using -gx to set global (exported) environment variables
switch "$MODE"
    case "local"
        set -gx DEV_MODE "local"
        set -gx IMAGE_REGISTRY "$LOCAL_IMAGE_REGISTRY_PREFIX"
        set -gx K8S_CONTEXT_TARGET "$LOCAL_K8S_CONTEXT"
        set -gx K8S_NAMESPACE "$LOCAL_K8S_NAMESPACE"
        set -gx HELM_VALUES_FILE "$PROJECT_ROOT/$LOCAL_HELM_VALUES_FILE_REL"
        set -l ACTIVATION_MESSAGE "Activated LOCAL development mode."
    case "remote-dev"
        set -gx DEV_MODE "remote-dev"
        set -gx IMAGE_REGISTRY "$REMOTE_DEV_IMAGE_REGISTRY_PREFIX"
        set -gx K8S_CONTEXT_TARGET "$REMOTE_DEV_K8S_CONTEXT"
        set -gx K8S_NAMESPACE "$REMOTE_DEV_K8S_NAMESPACE"
        set -gx HELM_VALUES_FILE "$PROJECT_ROOT/$REMOTE_DEV_HELM_VALUES_FILE_REL"
        set -l ACTIVATION_MESSAGE "Activated REMOTE-DEV mode."
        set -l ADDITIONAL_INFO "Ensure you are logged into container registry: $IMAGE_REGISTRY"
    case "remote-prod"
        set -gx DEV_MODE "remote-prod"
        set -gx IMAGE_REGISTRY "$REMOTE_PROD_IMAGE_REGISTRY_PREFIX"
        set -gx K8S_CONTEXT_TARGET "$REMOTE_PROD_K8S_CONTEXT"
        set -gx K8S_NAMESPACE "$REMOTE_PROD_K8S_NAMESPACE"
        set -gx HELM_VALUES_FILE "$PROJECT_ROOT/$REMOTE_PROD_HELM_VALUES_FILE_REL"
        set -l ACTIVATION_MESSAGE "Activated REMOTE-PROD mode."
        set -l ADDITIONAL_INFO "Ensure you are logged into container registry: $IMAGE_REGISTRY"
        echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
        echo "!!! WARNING: Production mode activated. Use caution. !!!"
        echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    case '*'
        echo "Invalid mode: $MODE. Use 'local', 'remote-dev', or 'remote-prod'."
        return 1 # Use return for sourced scripts
end

echo "$ACTIVATION_MESSAGE"
if set -q ADDITIONAL_INFO # Check if variable is defined
    echo "$ADDITIONAL_INFO"
end

echo "--------------------------------------------------"
echo "DEV_MODE:           $DEV_MODE"
echo "IMAGE_REGISTRY:     $IMAGE_REGISTRY"
echo "K8S_CONTEXT_TARGET: $K8S_CONTEXT_TARGET"
echo "K8S_NAMESPACE:      $K8S_NAMESPACE"
echo "HELM_VALUES_FILE:   $HELM_VALUES_FILE"
echo "--------------------------------------------------"

if not test -f "$HELM_VALUES_FILE"
    echo "WARNING: Helm values file ($HELM_VALUES_FILE) not found!"
    echo "You may need to create it or adjust the path in activate.fish."
end

# Attempt to set Kubernetes context
set -l CURRENT_K8S_CONTEXT (kubectl config current-context 2>/dev/null)

if [ "$CURRENT_K8S_CONTEXT" != "$K8S_CONTEXT_TARGET" ]
    echo ""
    echo "The target Kubernetes context is '$K8S_CONTEXT_TARGET'."
    if test -n "$CURRENT_K8S_CONTEXT"
        echo "Your current context is '$CURRENT_K8S_CONTEXT'."
    else
        echo "You do not seem to have a current Kubernetes context set."
    end

    if confirm_action_fish "Do you want to switch to context '$K8S_CONTEXT_TARGET' now?"
        if kubectl config use-context "$K8S_CONTEXT_TARGET"
            echo "Successfully switched to Kubernetes context '$K8S_CONTEXT_TARGET'."
        else
            echo "Failed to switch Kubernetes context. Please do it manually: kubectl config use-context $K8S_CONTEXT_TARGET"
        end
    else
        echo "Kubernetes context not changed. Remember to set it manually if needed: kubectl config use-context $K8S_CONTEXT_TARGET"
    end
else
    echo ""
    echo "Kubernetes context '$K8S_CONTEXT_TARGET' is already active."
end

echo ""
echo "To deactivate this mode, start a new shell or unset the environment variables (e.g., set -e DEV_MODE)."
