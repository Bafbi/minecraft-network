#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

sh "${SCRIPT_DIR}/build-and-load.sh"

kubectl rollout restart Deployment.apps network-proxy
