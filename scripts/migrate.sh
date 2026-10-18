#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ -f configs/.env ]]; then
  set -a
  # shellcheck disable=SC1091
  source configs/.env
  set +a
fi

if ! command -v goose >/dev/null 2>&1; then
  echo "goose CLI not found. Install it with: brew install goose" >&2
  exit 1
fi

: "${DATABASE_URL:?DATABASE_URL is required}"

goose -dir db/migrations -table learning_go_web_notebook_goose_db_version postgres "$DATABASE_URL" "$@"
