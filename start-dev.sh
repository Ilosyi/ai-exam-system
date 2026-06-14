#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$ROOT_DIR/server"
CLIENT_DIR="$ROOT_DIR/client"
SERVER_PORT="$(sed -nE 's/.*"serverPort"[[:space:]]*:[[:space:]]*([0-9]+).*/\1/p' "$ROOT_DIR/config.json" | head -1)"
CLIENT_PORT="$(sed -nE 's/.*"clientPort"[[:space:]]*:[[:space:]]*([0-9]+).*/\1/p' "$ROOT_DIR/config.json" | head -1)"

SERVER_PID=""
CLIENT_PID=""

cleanup() {
  local exit_code=$?
  if [[ -n "$CLIENT_PID" ]] && kill -0 "$CLIENT_PID" 2>/dev/null; then
    kill "$CLIENT_PID" 2>/dev/null || true
  fi
  if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
  fi
  wait 2>/dev/null || true
  exit "$exit_code"
}

wait_for_processes() {
  while true; do
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
      wait "$SERVER_PID" 2>/dev/null || return $?
      return 1
    fi
    if ! kill -0 "$CLIENT_PID" 2>/dev/null; then
      wait "$CLIENT_PID" 2>/dev/null || return $?
      return 1
    fi
    sleep 1
  done
}

trap cleanup INT TERM EXIT

echo "[start-dev] cleaning old dev processes..."
"$ROOT_DIR/stop-dev.sh"

echo "[start-dev] starting backend..."
(
  cd "$SERVER_DIR"
  go run .
) &
SERVER_PID=$!

echo "[start-dev] starting frontend..."
(
  cd "$CLIENT_DIR"
  pnpm dev
) &
CLIENT_PID=$!

echo "[start-dev] backend pid: $SERVER_PID"
echo "[start-dev] frontend pid: $CLIENT_PID"
echo "[start-dev] backend url: http://localhost:${SERVER_PORT:-8080}"
echo "[start-dev] frontend url: http://localhost:${CLIENT_PORT:-3000}"
echo "[start-dev] press Ctrl+C to stop both"

wait_for_processes
