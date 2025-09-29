# Docker Infrastructure for Flow-AI

This directory contains all Docker-related configurations for running the Flow-AI application. We use a flexible "base and override" model with Docker Compose to manage different environments easily.

## Core Concept: The "Base and Override" Model

We follow the **DRY (Don't Repeat Yourself)** principle. Instead of having large, almost identical configuration files, we split them:

- **`compose.base.yaml`**: This file is the skeleton of our application. It defines the core services (`flow-ai`, `ollama`), the network they share, and the data volumes. It is **always** used with an override file.
- **Override Files (`compose.prod.yaml`, `compose.dev.yaml`, etc.)**: These files are small and contain **only the differences** for a specific environment. For example, `compose.dev.yaml` adds volume mounts for live code reloading, while `compose.prod.yaml` sets restart policies.

This approach makes the configuration clean, easy to maintain, and flexible.

---

## How to Run

All commands should be run from the **project's root directory**.

### 1. Production Mode

This is the standard way to run the application for daily use. It builds a single, optimized container for the frontend and backend, served by Nginx.

```bash
# Build and start the application in the background
docker compose -f docker/compose.base.yaml -f docker/compose.prod.yaml up --build -d
```

The application will be available at `http://localhost:3000` (or the port you set in the `.env` file).

#### Production with an NVIDIA GPU

If you have an NVIDIA GPU, you can accelerate Ollama by adding the `compose.gpu.yaml` override file.

```bash
docker compose \
  -f docker/compose.base.yaml \
  -f docker/compose.prod.yaml \
  -f docker/compose.gpu.yaml \
  up --build -d
```

### 2. Development Mode

This mode is for developers. It enables hot-reloading for the backend, allowing you to see code changes instantly without rebuilding the container.

```bash
# Start the development environment in the foreground
docker compose -f docker/compose.base.yaml -f docker/compose.dev.yaml up --build
```

- The backend API will be available at `http://localhost:8000`.
- **Note:** The frontend is not served in this mode. You should run it separately if needed.

### 3. Running Automated Tests

This is the most important command for contributors. It runs the entire suite of integration tests in a clean, isolated environment and generates a code coverage report.

```bash
# This single command builds, runs the tests, and cleans up automatically.
HOST_UID=$(id -u) HOST_GID=$(id -g) docker compose -f docker/compose.base.yaml -f docker/compose.test.yaml up --build --abort-on-container-exit
```

After the command finishes, a `coverage/` directory will appear in the project root. Open `coverage/coverage.html` in your browser to see a detailed report.

#### A Note on `HOST_UID` and `HOST_GID`

When Docker creates files inside a volume (like our coverage reports), those files are owned by the `root` user by default. This can cause permission problems on your host machine.

By passing `HOST_UID=$(id -u)` and `HOST_GID=$(id -g)`, we tell our test runner script to change the owner of the generated files to match your current user. This avoids all permission issues.

### Stopping the Application

To stop an environment, use the `down` command with the same files you used to start it.

```bash
# Stop the production containers
docker compose -f docker/compose.base.yaml -f docker/compose.prod.yaml down

# To stop AND delete all data (database and models), add the -v flag
docker compose -f docker/compose.base.yaml -f docker/compose.prod.yaml down -v
```