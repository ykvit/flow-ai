# Backend Service for Flow-AI

This directory contains the Go backend service for the Flow-AI application. It acts as a robust API server that handles all business logic, communicates with the Ollama service for AI generation, and uses **SQLite** for data persistence.

For a high-level overview of the project's architecture, please see the main [Architecture Documentation](../DOCUMENTATION.md).

## Project Structure

The backend follows the principles of Clean Architecture to ensure a clear separation of concerns.

-   **/cmd/server**: The main entry point of the application.
-   **/internal/app**: The core application logic, responsible for initializing and wiring all components together.
-   **/internal/api**: Contains HTTP handlers, routing (`go-chi`), input validation, and API response models.
-   **/internal/service**: Holds the core business logic.
-   **/internal/repository**: Manages all data persistence via a `Repository` interface, with a concrete implementation for **SQLite**.
-   **/internal/database**: Contains logic for initializing the SQLite connection and running **schema migrations**.
-   **/internal/errors**: Defines **custom sentinel errors** used across the service layer.
-   **/internal/llm**: Abstracts communication with Ollama via an `LLMProvider` interface.
-   **/internal/model**: Defines the core data structures (`Chat`, `Message`).
-   **/internal/config**: Handles loading of bootstrap configuration using **Viper** from a `.env` file and environment variables.
-   **/tests**: Contains integration tests for the API.

## Core Features & Logic

-   **Real-time Chat Streaming**: Uses Server-Sent Events (SSE) to stream responses from the LLM to the client.
-   **Advanced Chat Persistence**: Saves all chat history in a local SQLite database. The schema supports a tree-like structure, enabling features like **response regeneration** and conversation branching.
-   **Version-Controlled Database Schema**: Manages all database changes through `golang-migrate`, ensuring predictable and automated schema evolution.
-   **Input Validation**: Enforces strict validation rules on all incoming API requests using `go-playground/validator`.
-   **Dynamic Title Generation**: Automatically creates a concise title for new conversations by prompting a support model.
-   **Full Model Management**: Provides a complete API to list, pull, and delete local Ollama models.
-   **Dynamic & Self-Healing Configuration**: On first launch, the `SettingsService` discovers available Ollama models, selects one as a default, and saves the configuration to the database. It can self-heal if models are added later.

## API Documentation

The API is documented using OpenAPI (Swagger). When the backend is running, you can access the interactive UI at:
**[http://localhost:8000/api/swagger/index.html](http://localhost:8000/api/swagger/index.html)** (in `dev` mode).

After making changes to the API handlers (adding or modifying routes/parameters), you must regenerate the documentation. Run this command from the **project root**:

```sh
make swag
```

## Development Workflow

We use a suite of tools to maintain high code quality and streamline development, all conveniently wrapped in `make` commands. All commands should be run from the **project root**.

-   **To format all Go code:**
    ```sh
    make format
    ```
-   **To run the linter and check for issues:**
    ```sh
    make lint
    ```
-   **To manage database migrations:**
    ```sh
    # Create new migration files
    make migrate-create name=your_migration_name

    # Apply migrations (usually not needed, as `make dev` does this automatically)
    make migrate-up
    ```

## Testing

The project contains a suite of integration tests that run inside a fully containerized, isolated environment.

To run all backend tests and generate an HTML coverage report, run the following command from the **project root**:

```sh
make test-backend
```
The report will be available at `coverage/coverage.html`.

## Running Locally (Without Docker)

While the primary method is using Docker Compose (`make dev` from the project root), you can run the backend as a standalone service for quick debugging.

1.  Ensure you have Go (1.23+) installed and an Ollama instance running locally.
2.  **Create a configuration file:** In the **project root**, copy `.env.example` to `.env` and modify it if needed. The backend will automatically find and use this file. A minimal `.env` for local execution would be:

    ```dotenv
    DATABASE_PATH="./flow-dev.db"
    OLLAMA_URL="http://localhost:11434"
    LOG_LEVEL="DEBUG"
    ```

3.  **Apply database migrations:** Before the first run, you must create the database schema. Ensure you have the `migrate` CLI installed (see `Dockerfile` for installation method) and run from the **project root**:
    ```sh
    make migrate-up
    ```

4.  **Run the application:** From the `backend/` directory, execute:
    ```sh
    go run ./cmd/server
    ```