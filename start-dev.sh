#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$ROOT_DIR/server"
CLIENT_DIR="$ROOT_DIR/client"

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

trap cleanup INT TERM EXIT

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
echo "[start-dev] press Ctrl+C to stop both"

wait -n "$SERVER_PID" "$CLIENT_PID"
