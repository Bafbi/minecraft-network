#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
SERVER_NAME="$(basename "${PROJECT_DIR}")"
IMAGE_NAME="${SERVER_NAME}:dev"
KIND_CLUSTER_NAME="minecraft-net"

docker buildx inspect builder >/dev/null 2>&1 || docker buildx create --name builder --use

echo "🔨 Building Go-based Gate proxy image '${IMAGE_NAME}'..."
docker buildx build \
  --platform linux/amd64 \
  --load \
  -t "${IMAGE_NAME}" \
  "${PROJECT_DIR}"

echo "📦 Loading image into kind cluster '${KIND_CLUSTER_NAME}'..."
kind load docker-image "${IMAGE_NAME}" --name "${KIND_CLUSTER_NAME}"

echo "✅ Done: '${IMAGE_NAME}' built and loaded into '${KIND_CLUSTER_NAME}'."
