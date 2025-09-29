# ====================================================================================
# Makefile for the Flow-AI Project
#
# Provides a simple and consistent interface for development, testing, and deployment tasks.
# ====================================================================================

# Use bash for all shell commands
SHELL := /bin/bash

# Get the host user's UID and GID to fix file permissions from Docker.
# This ensures that generated files (like test coverage) are owned by you, not root.
HOST_UID := $(shell id -u)
HOST_GID := $(shell id -g)

# Define the Docker Compose file configurations for each environment.
# This makes the commands below cleaner and easier to manage.
COMPOSE_BASE_FILE := -f docker/compose.base.yaml
COMPOSE_DEV_FILE  := -f docker/compose.dev.yaml
COMPOSE_PROD_FILE := -f docker/compose.prod.yaml
COMPOSE_TEST_FILE := -f docker/compose.test.yaml

# Define the full docker compose commands.
COMPOSE_DEV_CMD  := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_DEV_FILE)
COMPOSE_PROD_CMD := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_PROD_FILE)
COMPOSE_TEST_CMD := docker compose $(COMPOSE_BASE_FILE) $(COMPOSE_TEST_FILE)

# The 'flow-ai' service in our compose files is used as a runner for Go commands.
# It's based on the 'builder-backend' stage which contains all necessary tools.
GO_RUNNER_SERVICE := flow-ai

# Phony targets are not files. This tells Make to always run the command.
.PHONY: help dev prod test down logs swag lint format

# Set the default goal to 'help' so that running 'make' without arguments shows the help message.
.DEFAULT_GOAL := help

##@ Main Commands
help: ##> 📖 Show this help message.
	@awk 'BEGIN {FS = ":.*?##> "} /^[a-zA-Z_-]+:.*?##> / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ##> 🚀 Start the development environment (with hot-reloading for frontend).
	@echo "--- Starting development environment ---"
	$(COMPOSE_DEV_CMD) up --build

prod: ##> 🚢 Start the production environment.
	@echo "--- Starting production environment ---"
	$(COMPOSE_PROD_CMD) up --build -d

test: ##> 🧪 Run all integration tests in an isolated environment.
	@echo "--- Running integration tests ---"
	HOST_UID=$(HOST_UID) HOST_GID=$(HOST_GID) $(COMPOSE_TEST_CMD) up \
		--build \
		--abort-on-container-exit \
		--exit-code-from $(GO_RUNNER_SERVICE)

down: ##> ⏹️ Stop and remove all project containers, networks, and volumes.
	@echo "--- Stopping all environments ---"
	$(COMPOSE_DEV_CMD) down -v --remove-orphans
	$(COMPOSE_PROD_CMD) down -v --remove-orphans
	$(COMPOSE_TEST_CMD) down -v --remove-orphans

logs: ##> 📜 View logs from the development environment.
	@echo "--- Tailing logs for development environment ---"
	$(COMPOSE_DEV_CMD) logs -f

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