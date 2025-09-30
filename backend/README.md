# Backend Service for Flow-AI

This directory contains the Go backend service for the Flow-AI application. It acts as a robust API server that handles all business logic, communicates with the Ollama service for AI generation, and uses **SQLite** for data persistence.

For a high-level overview of the project's architecture, please see the main [Architecture Documentation](../DOCUMENTATION.md).

## Project Structure

The backend follows the principles of Clean Architecture to ensure a clear separation of concerns.

- **/cmd/server**: The main entry point of the application.
- **/internal/api**: Contains HTTP handlers, routing (`go-chi`), and API response models.
- **/internal/service**: Holds the core business logic.
- **/internal/repository**: Manages all data persistence via a `Repository` interface, with a concrete implementation for **SQLite**.
- **/internal/database**: Contains logic for initializing the SQLite connection and schema.
- **/internal/llm**: Abstracts communication with Ollama via an `LLMProvider` interface.
- **/internal/model**: Defines the core data structures (`Chat`, `Message`).
- **/internal/config**: Handles loading of configuration from environment variables and `config.json`.
- **/tests**: Contains integration tests for the API.

## Core Features & Logic

- **Real-time Chat Streaming**: Uses Server-Sent Events (SSE) to stream responses from the LLM to the client.
- **Advanced Chat Persistence**: Saves all chat history and messages in a local SQLite database. The schema supports a tree-like structure, enabling implemented features like **response regeneration** and future features like conversation branching.
- **Response Regeneration**: Allows users to regenerate the last assistant response, creating a new, active conversation branch while preserving the old one.
- **Dynamic Title Generation**: Automatically creates a concise title for new conversations by prompting a support model. The service includes "smart parsing" logic to extract a clean title from model outputs.
- **Full Model Management**: Provides a complete API to list, pull, and delete local Ollama models.
- **Dynamic & Self-Healing Configuration**: On first launch, the `SettingsService` discovers available Ollama models, selects one as a default, and saves the configuration. If no models are present, it will auto-configure itself later when a model becomes available.
- **Per-Request Generation Parameters**: Allows clients to override LLM options (like temperature, seed) for each message.

## API Documentation

The API is documented using OpenAPI (Swagger). When the backend is running, you can access the interactive UI at:
**[http://localhost:8000/api/swagger/index.html](http://localhost:8000/api/swagger/index.html)**

After making changes to the API handlers (adding or modifying routes/parameters), you must regenerate the documentation. Run this command from the **project root**:
```sh
make swag
```

## Code Quality

We use a suite of tools to maintain high code quality, all conveniently wrapped in `make` commands.

- **To format all Go code:**
  ```sh
  make format
  ```
- **To run the linter and check for issues:**
  ```sh
  make lint
  ```

## Testing

The project contains a suite of integration tests that run inside a fully containerized, isolated environment.

To run all tests and generate an HTML coverage report, simply run the following command from the **project root**:
```sh
make test
```
The report will be available at `coverage/coverage.html`.

## Running Locally (Without Docker)

While the primary method is using Docker Compose (`make dev`), you can run the backend as a standalone service.

1.  Ensure you have Go (1.22+) installed and an Ollama instance running.
2.  Set the required environment variables:
    ```sh
    export DATABASE_PATH="./flow-dev.db"
    export OLLAMA_URL="http://localhost:11434"
    export LOG_LEVEL="DEBUG"
    ```
3.  Run the application from the `backend/` directory:
    ```sh
    go run ./cmd/server
    ```