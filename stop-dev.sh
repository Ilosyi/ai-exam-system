#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}[stop-dev] stopping all dev processes...${NC}"

# Kill frontend process (pnpm dev / vite)
if pgrep -f "pnpm dev" > /dev/null; then
  pkill -f "pnpm dev" || true
  echo -e "${GREEN}[stop-dev] ✓ frontend stopped${NC}"
else
  echo -e "${YELLOW}[stop-dev] - frontend not running${NC}"
fi

# Kill backend process (go run)
if pgrep -f "go run" > /dev/null; then
  pkill -f "go run" || true
  echo -e "${GREEN}[stop-dev] ✓ backend stopped${NC}"
else
  echo -e "${YELLOW}[stop-dev] - backend not running${NC}"
fi

echo -e "${GREEN}[stop-dev] done${NC}"
