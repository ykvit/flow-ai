# Docker Infrastructure for Flow-AI

This directory contains all Docker-related configurations for running the Flow-AI application. We use a flexible "base and override" model with Docker Compose to manage different environments easily and safely.

## Core Concept: The "Base and Override" Model

We follow the **DRY (Don't Repeat Yourself)** principle. Instead of having large, almost identical configuration files, we split them:

- **`compose.base.yaml`**: This file is the skeleton of our application. It defines the core services (`flow-ai`, `ollama`), the network they share, and the data volumes. It is **never used alone** and always requires an override file.
- **Override Files (`compose.prod.yaml`, `compose.dev.yaml`, etc.)**: These files are small and contain **only the differences** for a specific environment. For example:
  - `compose.dev.yaml` uses a builder stage for live code reloading and exposes individual service ports.
  - `compose.prod.yaml` builds a final, optimized image and exposes a single port through an Nginx reverse proxy.
  - `compose.test.yaml` re-purposes the `flow-ai` service to act as a test runner instead of a web server.

This approach makes the configuration clean, easy to maintain, and flexible.

---

## How to Run

All commands should be run from the **project's root directory** using the provided `Makefile`, which simplifies the Docker Compose commands.

### 1. Production Mode

This is the standard way to run the application for daily use. It builds a single, optimized container for the frontend and backend, served by Nginx.

```bash
# Build and start the application in the background
make prod
```

The application will be available at `http://localhost:3000`.

### 2. Development Mode

This mode is for developers. It enables hot-reloading for the frontend and allows you to restart the backend to see changes.

```bash
# Start the development environment in the foreground
make dev
```

- **Backend API:** `http://localhost:8000/api/v1/`
- **Frontend Vite Server:** `http://localhost:5173/`

### 3. Running Automated Tests

This command runs the entire suite of integration tests in a clean, isolated environment and generates a code coverage report.

```bash
# This single command builds, runs the tests, and cleans up automatically.
make test
```

After the command finishes, a `coverage/` directory will appear in the project root. Open `coverage/coverage.html` in your browser to see a detailed report.

#### A Note on `HOST_UID` and `HOST_GID`

When Docker creates files inside a volume (like our coverage reports), those files are owned by the `root` user by default. This can cause permission problems on your host machine.

Our `Makefile` and `test_runner.sh` script automatically handle this by passing your local user's UID and GID into the container. The script then uses `chown` to change the owner of the generated files to match your user, avoiding all permission issues.

### Stopping the Application

To stop all environments, remove all containers, and clean up associated volumes, use the `down` command.

```bash
# Stop and remove everything
make down
```