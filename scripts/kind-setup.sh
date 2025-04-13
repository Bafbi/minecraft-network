#!/bin/bash

# Absolute path to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
HOST_DATA_PATH="${PROJECT_ROOT}/data"

# Create a temporary config file with the actual path injected
CONFIG_FILE="$(mktemp)"
envsubst < "${SCRIPT_DIR}/kind-config.yaml.in" > "${CONFIG_FILE}"

# Create the cluster
kind create cluster --name minecraft-net --config "${CONFIG_FILE}"

# Cleanup the temp file
rm "${CONFIG_FILE}"

# Load your local images if needed
# kind load docker-image my-custom-proxy:latest --name minecraft-net

# Deploy your Helm stack
helm install network ${PROJECT_ROOT}/charts/network -f ${PROJECT_ROOT}/values/dev-values.yaml
