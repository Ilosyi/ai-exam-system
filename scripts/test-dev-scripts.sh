#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SERVER_PORT="$(sed -nE 's/.*"serverPort"[[:space:]]*:[[:space:]]*([0-9]+).*/\1/p' "$ROOT_DIR/config.json" | head -1)"
CLIENT_PORT="$(sed -nE 's/.*"clientPort"[[:space:]]*:[[:space:]]*([0-9]+).*/\1/p' "$ROOT_DIR/config.json" | head -1)"

if [[ -z "$SERVER_PORT" || -z "$CLIENT_PORT" ]]; then
  echo "failed to read ports from config.json" >&2
  exit 1
fi

"$ROOT_DIR/stop-dev.sh" >/tmp/week05-stop-dev-test.log 2>&1 || {
  cat /tmp/week05-stop-dev-test.log >&2
  exit 1
}

if lsof -nP -iTCP:"$SERVER_PORT" -sTCP:LISTEN >/dev/null 2>&1; then
  echo "server port $SERVER_PORT is still listening after stop-dev.sh" >&2
  lsof -nP -iTCP:"$SERVER_PORT" -sTCP:LISTEN >&2 || true
  exit 1
fi

if lsof -nP -iTCP:"$CLIENT_PORT" -sTCP:LISTEN >/dev/null 2>&1; then
  echo "client port $CLIENT_PORT is still listening after stop-dev.sh" >&2
  lsof -nP -iTCP:"$CLIENT_PORT" -sTCP:LISTEN >&2 || true
  exit 1
fi

echo "dev scripts stop configured ports cleanly"
