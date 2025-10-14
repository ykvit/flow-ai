# ==============================================================================
# Makefile for the Flow-AI Project
# Provides a single, unified interface for all development and operational tasks.
# ==============================================================================

# Use bash for all shell commands to ensure consistent scripting behavior.
SHELL := /bin/bash

# Capture the host user's UID and GID to fix file permission issues when
# Docker creates files in mounted volumes (e.g., test reports).
HOST_UID := $(shell id -u)
HOST_GID := $(shell id -g)

# Define base Docker Compose files for a DRY (Don't Repeat Yourself) setup.
COMPOSE_BASE_FILE := -f docker/compose.base.yaml
COMPOSE_DEV_FILE  := -f docker/compose.dev.yaml
COMPOSE_PROD_FILE := -f docker/compose.prod.yaml
COMPOSE_TEST_FILE := -f docker/compose.test.yaml

# Allow optional GPU support by layering an additional compose file.
# Usage: `make dev GPU=1`
COMPOSE_GPU :=
ifeq ($(GPU),1)
	COMPOSE_GPU := -f docker/compose.gpu.yaml
endif

# Define reusable command snippets to keep targets clean and consistent.
# One-off tool commands (lint, format, migrate) reuse the DEV environment for convenience.
COMPOSE_DEV_CMD  := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_DEV_FILE) $(COMPOSE_GPU)
COMPOSE_PROD_CMD := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_PROD_FILE) $(COMPOSE_GPU)
COMPOSE_TEST_CMD := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_TEST_FILE)

# Define service names as variables for easy reference and modification.
BACKEND_SERVICE_NAME := flow-ai
FRONTEND_SERVICE_NAME := frontend

# Declare all targets as .PHONY to prevent conflicts with files of the same name.
.PHONY: help dev prod down logs swag lint format format-check dev-gpu prod-gpu down-dev down-prod down-test test test-backend test-frontend test-ci migrate-create migrate-up migrate-down

# Set the default goal to 'help' so running `make` without arguments shows the help message.
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
	# Pass host UID/GID to the test container to ensure reports are owned by the host user.
	# `--abort-on-container-exit` and `--exit-code-from` ensure make fails if tests fail.
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
	# `-v` removes volumes, `--remove-orphans` cleans up unused containers.
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
	# CI builds should always be clean to avoid stale cache issues.
	HOST_UID=$(HOST_UID) HOST_GID=$(HOST_GID) $(COMPOSE_TEST_CMD) build --no-cache $(BACKEND_SERVICE_NAME)
	@echo "--- Running CI integration tests and capturing logs ---"
	# `set -o pipefail` ensures that the exit code of `docker compose` is preserved
	# even when its output is piped to `tee`.
	set -o pipefail; \
	HOST_UID=$(HOST_UID) HOST_GID=$(HOST_GID) $(COMPOSE_TEST_CMD) up \
		--abort-on-container-exit \
		--exit-code-from $(BACKEND_SERVICE_NAME) \
		$(BACKEND_SERVICE_NAME) 2>&1 | tee test-logs.txt

format-check: ##> 🧐 Check if Go code is formatted (for CI).
	@echo "--- Checking Go code formatting ---"
	# This command fails if `goimports -l` produces any output, making it ideal for CI checks.
	@if [ -n "$$($(COMPOSE_DEV_CMD) run --rm $(BACKEND_SERVICE_NAME) goimports -l .)" ]; then \
		echo "The following files are not formatted correctly:"; \
		$(COMPOSE_DEV_CMD) run --rm $(BACKEND_SERVICE_NAME) goimports -l .; \
		exit 1; \
	fi
	@echo "All files are correctly formatted."

##@ Code Quality & Docs
swag: ##> 📄 Regenerate Swagger/OpenAPI documentation.
	@echo "--- Regenerating Swagger documentation ---"
	# Run as the host user to ensure generated docs have correct file permissions.
	# We set GOCACHE=/tmp to a world-writable directory to prevent permission
	# errors when swag tries to use the Go build cache as a non-root user.
	@$(COMPOSE_DEV_CMD) run --rm \
		--user $(HOST_UID):$(HOST_GID) \
		-e GOCACHE=/tmp \
		$(BACKEND_SERVICE_NAME) \
		swag init -g ./cmd/server/main.go --output ./docs --parseDependency

lint: ##> 🔍 Run the Go linter (golangci-lint).
	@echo "--- Running Go linter ---"
	@$(COMPOSE_DEV_CMD) run --rm $(BACKEND_SERVICE_NAME) golangci-lint run -v ./...

format: ##> ✨ Automatically format all Go source code.
	@echo "--- Formatting Go code ---"
	@$(COMPOSE_DEV_CMD) run --rm $(BACKEND_SERVICE_NAME) goimports -w .

##@ Database Migrations
migrate-create: ##> 📦 Create new SQL migration files. Usage: make migrate-create name=add_users_table
	@echo "--- Creating migration files for [$(name)] ---"
	# Run as host user so the new migration files aren't owned by root.
	# The `name` argument is passed from the `make` command line.
	@$(COMPOSE_DEV_CMD) run --rm --user $(HOST_UID):$(HOST_GID) $(BACKEND_SERVICE_NAME) \
		migrate create -ext sql -dir internal/database/migrations -seq $(name)

migrate-up: ##> 📈 Apply all pending database migrations.
	@echo "--- Applying database migrations ---"
	# The DATABASE_PATH env var is automatically picked up by the container from compose.dev.yaml.
	@$(COMPOSE_DEV_CMD) run --rm $(BACKEND_SERVICE_NAME) \
		migrate -path internal/database/migrations -database "sqlite3://$(DATABASE_PATH)" up

migrate-down: ##> 📉 Roll back the last database migration.
	@echo "--- Rolling back last migration ---"
	@$(COMPOSE_DEV_CMD) run --rm $(BACKEND_SERVICE_NAME) \
		migrate -path internal/database/migrations -database "sqlite3://$(DATABASE_PATH)" down