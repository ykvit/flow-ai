# Makefile for the Flow-AI Project
# Provides a simple and consistent interface for development, testing, and deployment tasks.

SHELL := /bin/bash
HOST_UID := $(shell id -u)
HOST_GID := $(shell id -g)

COMPOSE_BASE_FILE := -f docker/compose.base.yaml
COMPOSE_DEV_FILE  := -f docker/compose.dev.yaml
COMPOSE_PROD_FILE := -f docker/compose.prod.yaml
COMPOSE_TEST_FILE := -f docker/compose.test.yaml

COMPOSE_GPU :=
ifeq ($(GPU),1)
	COMPOSE_GPU := -f docker/compose.gpu.yaml
endif

COMPOSE_DEV_CMD  := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_DEV_FILE) $(COMPOSE_GPU)
COMPOSE_PROD_CMD := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_PROD_FILE) $(COMPOSE_GPU)
COMPOSE_TEST_CMD := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_TEST_FILE) $(COMPOSE_GPU)

BACKEND_SERVICE_NAME := flow-ai
FRONTEND_SERVICE_NAME := frontend

.PHONY: help dev prod down logs swag lint format dev-gpu prod-gpu down-dev down-prod down-test test test-backend test-frontend test-ci
.DEFAULT_GOAL := help

##@ Main Commands
help: ##> 📖 Show this help message with sections.
	@awk 'BEGIN {FS = ":.*?##> "} /^##@/ { printf("\n\033[1;33m%s\033[0m\n", substr($$0, 5)); next; } /^[a-zA-Z0-9_.-]+:.*?##> / { printf("  \033[36m%-20s\033[0m %s\n", $$1, $$2); }' $(MAKEFILE_LIST)

dev: ##> 🚀 Start the development environment (CPU).
	@echo "--- Starting development environment (CPU) ---"
	$(COMPOSE_DEV_CMD) up --build

prod: ##> 🚢 Start the production environment (CPU).
	@echo "--- Starting production environment (CPU) ---"
	$(COMPOSE_PROD_CMD) up --build -d

logs: ##> 📜 View logs from the development environment.
	@echo "--- Tailing logs for development environment ---"
	$(COMPOSE_DEV_CMD) logs -f

##@ Testing
test: ##> 🧪 Run all available tests for the project (backend & frontend).
	@$(MAKE) test-backend
	@echo "\nNote: Frontend tests are not yet implemented. Structure is ready."
	# @$(MAKE) test-frontend # Uncomment when frontend tests are available

test-backend: ##> 🧪 (Backend) Run Go integration tests.
	@echo "--- Running Go integration tests (CPU) ---"
	HOST_UID=$(HOST_UID) HOST_GID=$(HOST_GID) $(COMPOSE_TEST_CMD) up \
		--build \
		--abort-on-container-exit \
		--exit-code-from $(BACKEND_SERVICE_NAME) \
		$(BACKEND_SERVICE_NAME) # CORRECTED: Explicitly specify which service to run

test-frontend: ##> 🧪 (Frontend) Run frontend tests.
	@echo "--- Running frontend tests (CPU) ---"
	$(COMPOSE_TEST_CMD) up \
		--build \
		--abort-on-container-exit \
		--exit-code-from $(FRONTEND_SERVICE_NAME) \
		$(FRONTEND_SERVICE_NAME) # CORRECTED: Explicitly specify which service to run

##@ GPU Aliases
dev-gpu: ##> 🚀 (GPU) Start the development environment.
	@$(MAKE) dev GPU=1

prod-gpu: ##> 🚢 (GPU) Start the production environment.
	@$(MAKE) prod GPU=1

##@ Teardown
down-dev: ##> ⏹️ Stop and clean up the DEV environment.
	@echo "--- Stopping DEV environment ---"
	$(COMPOSE_DEV_CMD) down -v --remove-orphans

down-prod: ##> ⏹️ Stop and clean up the PROD environment.
	@echo "--- Stopping PROD environment ---"
	$(COMPOSE_PROD_CMD) down -v --remove-orphans

down-test: ##> ⏹️ Stop and clean up the TEST environment.
	@echo "--- Stopping TEST environment ---"
	$(COMPOSE_TEST_CMD) down -v --remove-orphans

down: ##> ☢️ Stop and clean up ALL environments.
	@echo "--- Stopping ALL environments ---"
	@$(MAKE) down-dev
	@$(MAKE) down-prod
	@$(MAKE) down-test

##@ CI-Specific Commands
test-ci: ##> 🤖 (Backend) Run tests for CI (no cache).
	@echo "--- Running CI integration tests (no cache) ---"
	HOST_UID=$(HOST_UID) HOST_GID=$(HOST_GID) $(COMPOSE_TEST_CMD) up \
		--build --no-cache \
		--abort-on-container-exit \
		--exit-code-from $(BACKEND_SERVICE_NAME) \
		$(BACKEND_SERVICE_NAME) # CORRECTED: Explicitly specify which service to run

##@ Code Quality & Docs
swag: ##> 📄 Regenerate Swagger/OpenAPI documentation.
	@echo "--- Regenerating Swagger documentation ---"
	@$(COMPOSE_DEV_CMD) run --build --rm $(BACKEND_SERVICE_NAME) sh -c "swag init -g ./cmd/server/main.go && chown -R $(HOST_UID):$(HOST_GID) docs"

lint: ##> 🔍 Run the Go linter (golangci-lint).
	@echo "--- Running Go linter ---"
	@$(COMPOSE_DEV_CMD) run --build --rm $(BACKEND_SERVICE_NAME) golangci-lint run ./...

format: ##> ✨ Automatically format all Go source code.
	@echo "--- Formatting Go code ---"
	@$(COMPOSE_DEV_CMD) run --build --rm $(BACKEND_SERVICE_NAME) goimports -w .