#!/bin/bash
set -e

SCRIPT_DIR_LOBBY="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR_LOBBY="$(cd "${SCRIPT_DIR_LOBBY}/.." && pwd)"
SERVICE_NAME_LOBBY="$(basename "${SERVICE_DIR_LOBBY}")" # Should be "lobby_minestom"

GENERIC_BUILD_SCRIPT_PATH="${SCRIPT_DIR_LOBBY}/../../../scripts/build-service.sh"

bash "${GENERIC_BUILD_SCRIPT_PATH}" "${SERVICE_NAME_LOBBY}" "${SERVICE_DIR_LOBBY}"
