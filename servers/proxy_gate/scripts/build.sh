#!/bin/bash
set -e

SCRIPT_DIR_PROXY_GATE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR_PROXY_GATE="$(cd "${SCRIPT_DIR_PROXY_GATE}/.." && pwd)"
SERVICE_NAME_PROXY_GATE="$(basename "${SERVICE_DIR_PROXY_GATE}")" # Should be "proxy_gate"

# Path to the generic build script
GENERIC_BUILD_SCRIPT_PATH="${SCRIPT_DIR_PROXY_GATE}/../../../scripts/build-service.sh"

# Call the generic build script
bash "${GENERIC_BUILD_SCRIPT_PATH}" "${SERVICE_NAME_PROXY_GATE}" "${SERVICE_DIR_PROXY_GATE}"
