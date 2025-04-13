#!/bin/bash

set -e

# Detect paths relative to script location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVER_NAME="$(basename "${SCRIPT_DIR}")"
IMAGE_NAME="${SERVER_NAME}:dev"
KIND_CLUSTER_NAME="minecraft-net"

# Ensure buildx is ready
docker buildx inspect builder >/dev/null 2>&1 || docker buildx create --name builder --use

echo "🔨 Building image '${IMAGE_NAME}' with buildx..."
docker buildx build \
  --platform linux/amd64 \
  --load \
  -t "${IMAGE_NAME}" \
  "${SCRIPT_DIR}"

echo "📦 Loading image into kind cluster '${KIND_CLUSTER_NAME}'..."
kind load docker-image "${IMAGE_NAME}" --name "${KIND_CLUSTER_NAME}"

echo "✅ Done: '${IMAGE_NAME}' built and loaded into '${KIND_CLUSTER_NAME}'."
