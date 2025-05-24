#!/bin/bash

# Generic script to build a service Docker image and either load to Kind or push to a registry.
# It expects DEV_MODE and IMAGE_REGISTRY to be set by activate.sh
#
# Usage: scripts/build-service.sh <service_name> <service_dir_path> [<dockerfile_name>]
#   <service_name>: Name of the service (e.g., proxy_gate, lobby_minestom, permissions-webapp)
#   <service_dir_path>: Absolute path to the service's directory
#   <dockerfile_name>: Optional, name of the Dockerfile (default: Dockerfile)

set -e # Exit immediately if a command exits with a non-zero status.

if [[ -z "$DEV_MODE" ]]; then
  echo "Error: DEV_MODE is not set. Please run 'source scripts/activate.sh [local|remote]' first."
  exit 1
fi

SERVICE_NAME=$1
SERVICE_DIR=$2
DOCKERFILE_NAME=${3:-Dockerfile} # Default to Dockerfile if not provided

if [[ -z "$SERVICE_NAME" || -z "$SERVICE_DIR" ]]; then
  echo "Usage: $0 <service_name> <service_dir_path> [<dockerfile_name>]"
  exit 1
fi

IMAGE_TAG_BASE="${SERVICE_NAME}" # Base tag, e.g., proxy_gate
IMAGE_TAG_DEV="${IMAGE_TAG_BASE}:dev" # Local/dev tag

# For remote, we prepend the registry and use a 'latest' tag or a git SHA.
# For simplicity, let's use 'latest' for remote pushes from this script.
# A more advanced setup would use git commit SHAs for remote tags.
IMAGE_TAG_REMOTE_LATEST="${IMAGE_REGISTRY}/${IMAGE_TAG_BASE}:latest"
IMAGE_TAG_REMOTE_DEV="${IMAGE_REGISTRY}/${IMAGE_TAG_BASE}:dev" # Also push a :dev tag to remote

KIND_CLUSTER_NAME="minecraft-net" # As defined in your kind-setup.sh

echo "🔨 Building Docker image for service '${SERVICE_NAME}' from '${SERVICE_DIR}/${DOCKERFILE_NAME}'..."
echo "DEV_MODE: ${DEV_MODE}"

# Ensure Docker Buildx builder exists
docker buildx inspect builder >/dev/null 2>&1 || docker buildx create --name builder --use

if [[ "$DEV_MODE" == "local" ]]; then
  # Build for local (Kind) - typically linux/amd64 unless your Kind node is different
  # Load directly into Kind
  echo "Building for LOCAL (Kind) and loading image '${IMAGE_TAG_DEV}'..."
  docker buildx build \
    --platform linux/amd64 \
    --load \
    -t "${IMAGE_TAG_DEV}" \
    -f "${SERVICE_DIR}/${DOCKERFILE_NAME}" \
    "${SERVICE_DIR}"

  echo "📦 Loading image '${IMAGE_TAG_DEV}' into Kind cluster '${KIND_CLUSTER_NAME}'..."
  kind load docker-image "${IMAGE_TAG_DEV}" --name "${KIND_CLUSTER_NAME}"
  echo "✅ Done: Image '${IMAGE_TAG_DEV}' built and loaded into Kind."

elif [[ "$DEV_MODE" == "remote" ]]; then
  # Build and push for remote registry
  # Build for multiple platforms if your remote cluster might use them
  echo "Building for REMOTE and pushing images '${IMAGE_TAG_REMOTE_LATEST}' and '${IMAGE_TAG_REMOTE_DEV}'..."
  docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --push \
    -t "${IMAGE_TAG_REMOTE_LATEST}" \
    -t "${IMAGE_TAG_REMOTE_DEV}" \
    -f "${SERVICE_DIR}/${DOCKERFILE_NAME}" \
    "${SERVICE_DIR}"
  echo "✅ Done: Images pushed to ${IMAGE_REGISTRY}."
  echo "   ${IMAGE_TAG_REMOTE_LATEST}"
  echo "   ${IMAGE_TAG_REMOTE_DEV}"
else
  echo "Error: Invalid DEV_MODE '${DEV_MODE}'. Must be 'local' or 'remote'."
  exit 1
fi
