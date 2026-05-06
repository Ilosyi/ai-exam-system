.PHONY: help test-backend test-frontend test-all test-coverage test-watch test-setup

help:
	@echo "Testing commands:"
	@echo "  make test-backend      - Run all backend tests"
	@echo "  make test-frontend     - Run all frontend tests"
	@echo "  make test-all          - Run all tests (backend + frontend)"
	@echo "  make test-coverage     - Run tests with coverage reports"
	@echo "  make test-watch        - Run frontend tests in watch mode"

test-backend:
	cd server && go test ./... -v

test-frontend:
	cd client && pnpm test -- --run

test-all: test-backend test-frontend
	@echo "✅ All tests completed!"

test-coverage:
	@echo "Backend coverage:"
	cd server && go test ./... -cover
	@echo ""
	@echo "Frontend coverage:"
	cd client && pnpm test:coverage

test-watch:
	cd client && pnpm test

test-setup:
	@echo "Setting up test environment..."
	cd server && go mod tidy
	cd client && pnpm install
	@echo "✅ Test environment ready!"
