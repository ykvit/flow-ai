# Makefile for the Flow-AI Project
# Uses a unified Docker build stage for all development and CI tasks, ensuring consistency.

SHELL := /bin/bash
HOST_UID := $(shell id -u)
HOST_GID := $(shell id -g)

# Base Docker Compose command setup
COMPOSE_BASE_FILE := -f docker/compose.base.yaml
COMPOSE_DEV_FILE  := -f docker/compose.dev.yaml
COMPOSE_PROD_FILE := -f docker/compose.prod.yaml
COMPOSE_TEST_FILE := -f docker/compose.test.yaml

# GPU support is optional, controlled by `make <command> GPU=1`
COMPOSE_GPU :=
ifeq ($(GPU),1)
	COMPOSE_GPU := -f docker/compose.gpu.yaml
endif

# DRY command variables
# We use the DEV environment for all one-off tool commands
COMPOSE_DEV_CMD  := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_DEV_FILE) $(COMPOSE_GPU)
COMPOSE_PROD_CMD := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_PROD_FILE) $(COMPOSE_GPU)
COMPOSE_TEST_CMD := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_TEST_FILE)

# Service names for clarity
BACKEND_SERVICE_NAME := flow-ai
FRONTEND_SERVICE_NAME := frontend

.PHONY: help dev prod down logs swag lint format format-check dev-gpu prod-gpu down-dev down-prod down-test test test-backend test-frontend test-ci
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
	# @$(MAKE) test-frontend

test-backend: ##> 🧪 (Backend) Run Go integration tests.
	@echo "--- Running Go integration tests ---"
	HOST_UID=$(HOST_UID) HOST_GID=$(HOST_GID) $(COMPOSE_TEST_CMD) up \
		--build \
		--abort-on-container-exit \
		--exit-code-from $(BACKEND_SERVICE_NAME) \
		$(BACKEND_SERVICE_NAME)

test-frontend: ##> 🧪 (Frontend) Run frontend tests.
	@echo "--- Running frontend tests (placeholder) ---"
	# $(COMPOSE_TEST_CMD) up --build --abort-on-container-exit --exit-code-from $(FRONTEND_SERVICE_NAME) $(FRONTEND_SERVICE_NAME)

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
	@echo "--- Building test images with no cache ---"
	HOST_UID=$(HOST_UID) HOST_GID=$(HOST_GID) $(COMPOSE_TEST_CMD) build --no-cache $(BACKEND_SERVICE_NAME)
	@echo "--- Running CI integration tests ---"
	HOST_UID=$(HOST_UID) HOST_GID=$(HOST_GID) $(COMPOSE_TEST_CMD) up \
		--abort-on-container-exit \
		--exit-code-from $(BACKEND_SERVICE_NAME) \
		$(BACKEND_SERVICE_NAME)

format-check: ##> 🧐 Check if Go code is formatted (for CI).
	@echo "--- Checking Go code formatting ---"
	@if [ -n "$$($(COMPOSE_DEV_CMD) run --rm $(BACKEND_SERVICE_NAME) goimports -l .)" ]; then \
		echo "The following files are not formatted correctly:"; \
		$(COMPOSE_DEV_CMD) run --rm $(BACKEND_SERVICE_NAME) goimports -l .; \
		exit 1; \
	fi
	@echo "All files are correctly formatted."

##@ Code Quality & Docs
swag: ##> 📄 Regenerate Swagger/OpenAPI documentation.
	@echo "--- Regenerating Swagger documentation ---"
	# We run as the host user to ensure the generated files have the correct permissions.
	@$(COMPOSE_DEV_CMD) run --rm --user $(HOST_UID):$(HOST_GID) $(BACKEND_SERVICE_NAME) \
		swag init -g ./cmd/server/main.go --output ./docs

lint: ##> 🔍 Run the Go linter (golangci-lint).
	@echo "--- Running Go linter ---"
	@$(COMPOSE_DEV_CMD) run --rm $(BACKEND_SERVICE_NAME) golangci-lint run -v ./...

format: ##> ✨ Automatically format all Go source code.
	@echo "--- Formatting Go code ---"
	@$(COMPOSE_DEV_CMD) run --rm $(BACKEND_SERVICE_NAME) goimports -w .