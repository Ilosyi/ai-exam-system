#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONFIG_FILE="$ROOT_DIR/config.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

read_port() {
  local key="$1"
  local fallback="$2"
  local value=""
  if [[ -f "$CONFIG_FILE" ]]; then
    value="$(sed -nE "s/.*\"$key\"[[:space:]]*:[[:space:]]*([0-9]+).*/\\1/p" "$CONFIG_FILE" | head -1)"
  fi
  echo "${value:-$fallback}"
}

stop_by_port() {
  local label="$1"
  local port="$2"
  local pids
  pids="$(lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)"
  if [[ -z "$pids" ]]; then
    echo -e "${YELLOW}[stop-dev] - no listener on ${label} port ${port}${NC}"
    return
  fi

  echo "$pids" | xargs kill 2>/dev/null || true
  sleep 1

  pids="$(lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)"
  if [[ -n "$pids" ]]; then
    echo "$pids" | xargs kill -9 2>/dev/null || true
  fi
  echo -e "${GREEN}[stop-dev] ✓ ${label} port ${port} stopped${NC}"
}

SERVER_PORT="$(read_port serverPort 8080)"
CLIENT_PORT="$(read_port clientPort 3000)"

echo -e "${YELLOW}[stop-dev] stopping all dev processes...${NC}"

stop_by_port "frontend" "$CLIENT_PORT"
stop_by_port "backend" "$SERVER_PORT"

echo -e "${GREEN}[stop-dev] done${NC}"
