#!/bin/bash

# Absolute path to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

kubectl config use-context kind-minecraft-net

helm upgrade --install network $PROJECT_ROOT/charts/network -f values/dev-values.yaml
