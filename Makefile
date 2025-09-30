# ====================================================================================
# Makefile for the Flow-AI Project (v3.1)
#
# Provides a simple and consistent interface for development, testing, and deployment tasks.
# ====================================================================================

# Use bash for all shell commands
SHELL := /bin/bash

# Get the host user's UID and GID to fix file permissions from Docker.
HOST_UID := $(shell id -u)
HOST_GID := $(shell id -g)

# Define the Docker Compose file configurations for each environment.
COMPOSE_BASE_FILE := -f docker/compose.base.yaml
COMPOSE_DEV_FILE  := -f docker/compose.dev.yaml
COMPOSE_PROD_FILE := -f docker/compose.prod.yaml
COMPOSE_TEST_FILE := -f docker/compose.test.yaml

# --- GPU Support ---
# Conditionally add the GPU compose file if GPU=1 is passed.
COMPOSE_GPU :=
ifeq ($(GPU),1)
	COMPOSE_GPU := -f docker/compose.gpu.yaml
endif
# --------------------

# Define the full docker compose commands, now including optional GPU support.
COMPOSE_DEV_CMD  := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_DEV_FILE) $(COMPOSE_GPU)
COMPOSE_PROD_CMD := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_PROD_FILE) $(COMPOSE_GPU)
COMPOSE_TEST_CMD := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_TEST_FILE) $(COMPOSE_GPU)

# The 'flow-ai' service is used as a runner for Go commands.
GO_RUNNER_SERVICE := flow-ai

.PHONY: help dev prod test down logs swag lint format dev-gpu prod-gpu test-gpu down-dev down-prod down-test test-ci

# Set the default goal to 'help'.
.DEFAULT_GOAL := help

##@ Main Commands
help: ##> 📖 Show this help message with sections.
	@awk 'BEGIN {FS = ":.*?##> "} \
	/^##@/ { \
		printf("\n\033[1;33m%s\033[0m\n", substr($$0, 5)); \
		next; \
	} \
	/^[a-zA-Z0-9_.-]+:.*?##> / { \
		printf("  \033[36m%-18s\033[0m %s\n", $$1, $$2); \
	}' $(MAKEFILE_LIST)

dev: ##> 🚀 Start the development environment (CPU).
	@echo "--- Starting development environment (CPU) ---"
	$(COMPOSE_DEV_CMD) up --build

prod: ##> 🚢 Start the production environment (CPU).
	@echo "--- Starting production environment (CPU) ---"
	$(COMPOSE_PROD_CMD) up --build -d

test: ##> 🧪 Run all integration tests (CPU, uses cache).
	@echo "--- Running integration tests (CPU) ---"
	HOST_UID=$(HOST_UID) HOST_GID=$(HOST_GID) $(COMPOSE_TEST_CMD) up \
		--build \
		--abort-on-container-exit \
		--exit-code-from $(GO_RUNNER_SERVICE)

logs: ##> 📜 View logs from the development environment.
	@echo "--- Tailing logs for development environment ---"
	$(COMPOSE_DEV_CMD) logs -f

##@ GPU Aliases
dev-gpu: ##> 🚀 (GPU) Start the development environment.
	@$(MAKE) dev GPU=1

prod-gpu: ##> 🚢 (GPU) Start the production environment.
	@$(MAKE) prod GPU=1

test-gpu: ##> 🧪 (GPU) Run integration tests.
	@$(MAKE) test GPU=1

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
test-ci: ##> 🤖 Run tests for CI (no cache, ensures clean build).
	@echo "--- Running CI integration tests (no cache) ---"
	HOST_UID=$(HOST_UID) HOST_GID=$(HOST_GID) $(COMPOSE_TEST_CMD) up \
		--build --no-cache \
		--abort-on-container-exit \
		--exit-code-from $(GO_RUNNER_SERVICE)

##@ Code Quality & Docs
swag: ##> 📄 Regenerate Swagger/OpenAPI documentation.
	@echo "--- Regenerating Swagger documentation ---"
	@$(COMPOSE_DEV_CMD) run --rm $(GO_RUNNER_SERVICE) sh -c \
		"swag init -g backend/cmd/server/main.go && chown -R $(HOST_UID):$(HOST_GID) docs"
	@echo "Swagger docs generated in backend/docs/"

lint: ##> 🔍 Run the Go linter (golangci-lint) to check code quality.
	@echo "--- Running Go linter ---"
	@$(COMPOSE_DEV_CMD) run --rm $(GO_RUNNER_SERVICE) golangci-lint run ./...

format: ##> ✨ Automatically format all Go source code.
	@echo "--- Formatting Go code ---"
	@$(COMPOSE_DEV_CMD) run --rm $(GO_RUNNER_SERVICE) goimports -w .