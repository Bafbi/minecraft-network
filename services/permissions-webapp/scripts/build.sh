#!/bin/bash
set -e

SCRIPT_DIR_WEBAPP="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR_WEBAPP="$(cd "${SCRIPT_DIR_WEBAPP}/.." && pwd)"
SERVICE_NAME_WEBAPP="$(basename "${SERVICE_DIR_WEBAPP}")" # Should be "permissions-webapp"

GENERIC_BUILD_SCRIPT_PATH="${SCRIPT_DIR_WEBAPP}/../../../scripts/build-service.sh"

bash "${GENERIC_BUILD_SCRIPT_PATH}" "${SERVICE_NAME_WEBAPP}" "${SERVICE_DIR_WEBAPP}"
