#!/usr/bin/env bash
set -euo pipefail

# example.server.sh
#
# Runs defacto2-server with environment variables loaded from a file (default: .env.local).
# Both the binary and env file are expected in the same directory as this script.

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env.local"
SERVER_BIN="${SCRIPT_DIR}/defacto2-server"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Error: environment file not found: $ENV_FILE" >&2
  exit 1
fi

if [[ ! -x "$SERVER_BIN" ]]; then
  echo "Error: server binary not found or not executable: $SERVER_BIN" >&2
  exit 1
fi

echo "Loading environment variables from $ENV_FILE"
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

exec "$SERVER_BIN" "$@"
