#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

# 应用自身会加载 configs/.env；本脚本只提供稳定、易发现的开发入口。
exec go run ./cmd/api
