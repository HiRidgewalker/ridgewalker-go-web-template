#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

# The application loads configs/.env itself; this script only exists so the
# development command has a stable, discoverable entry point.
exec go run ./cmd/api
