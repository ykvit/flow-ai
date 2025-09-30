# Docker Infrastructure for Flow-AI

This directory contains all Docker configurations for running the Flow-AI application. The setup is designed following modern DevOps principles to be consistent, reliable, and easy to use across different environments.

## Core Philosophy

Our infrastructure is built on three key principles:

1.  **DRY (Don't Repeat Yourself) with the "Base and Override" Model:** We avoid duplicating configuration. A central `compose.base.yaml` file defines the shared skeleton of the application, while small `override` files (`dev`, `prod`, `test`) add or change only what is necessary for a specific environment.
2.  **Explicitness over Implicitness:** To ensure clarity and prevent unexpected behavior, we explicitly declare all structural connections, such as network attachments, in each override file. While this may look slightly repetitive, it makes each environment's configuration self-contained and easy to understand without relying on implicit merging logic.
3.  **Environment Isolation:** The test environment is completely isolated. It uses its own ephemeral database and model storage volumes (`ollama_test_data`), ensuring that automated tests never interfere with development data.

## File Breakdown

-   **`compose.base.yaml`**: The foundation. Defines the `flow-ai` and `ollama` services, the shared `flow-ai-net` network, and the common `ollama` external volume. It is never used alone.
-   **`compose.dev.yaml`**: The override for local development. Mounts source code for live updates, enables debugging features, and exposes ports for direct access.
-   **`compose.prod.yaml`**: The override for production. Builds from the final, optimized stage of the `Dockerfile` and runs the application behind an Nginx reverse proxy.
-   **`compose.test.yaml`**: The override for automated testing. Re-purposes the services to act as test runners and uses fully isolated, temporary volumes.
-   **`compose.gpu.yaml`**: An optional override that can be layered on top of any environment to add NVIDIA GPU acceleration to the `ollama` service.
-   **`Dockerfile`**: A multi-stage build file that creates optimized, reproducible images. It pins versions of all tools (`Go`, `golangci-lint`, etc.) to guarantee that builds are identical everywhere.

## Key Design Patterns

-   **Configuration via `.env` File:** The application is configured via a `.env` file in the project root. The `compose.base.yaml` uses the `env_file` directive to inject these variables into the containers at runtime. For security, the `.env` file itself is listed in `.dockerignore` and is never copied into the image.
-   **Reproducible Builds:** By pinning the exact versions of Go and all build tools in the `Dockerfile`, we ensure that the application builds the same way today, tomorrow, and on any developer's machine or CI server.

---

## How to Run

All commands should be run from the **project's root directory**. The provided `Makefile` is the single, unified interface for interacting with the Docker environment.

### 1. Production Mode

This is the standard way to run the application for daily use.

```bash
# Build and start the application in the background
make prod
```

The application will be available at `http://localhost:3000` (or the `APP_PORT` you set in `.env`).

### 2. Development Mode

This mode is for developers, enabling features like live code reloading.

```bash
# Start the development environment in the foreground
make dev
```

-   **Backend API:** `http://localhost:8000/api/v1/`
-   **Frontend Vite Server:** `http://localhost:5173/`

### 3. Running Automated Tests

This command runs the backend integration tests in a clean, isolated environment.

```bash
# This command builds, runs the tests, and cleans up automatically.
make test-backend
```

After the command finishes, a `coverage/` directory will appear in the project root. Open `coverage/coverage.html` in your browser to see a detailed report.

#### A Note on File Permissions (`HOST_UID`/`HOST_GID`)

When Docker creates files inside a volume (like our coverage reports), those files are owned by the `root` user by default. This can cause permission problems on your host machine. Our `Makefile` and `test_runner.sh` script automatically solve this by passing your local user's ID into the container, which then changes the ownership of the generated files to match your user.

### Stopping the Application

To stop all environments, remove all containers, and clean up associated volumes, use the `down` command.

```bash
# Stop and remove everything
make down
```