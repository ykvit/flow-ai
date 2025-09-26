
# Backend Service for Flow-AI

This directory contains the Go backend service for the Flow-AI application. It acts as a robust API server that handles all business logic, communicates with the Ollama service for AI generation, and uses **SQLite** for data persistence.

For a high-level overview of the project's architecture, please see the main [Architecture Documentation](../DOCUMENTATION.md).

## Project Structure

The backend follows the principles of Clean Architecture to ensure a clear separation of concerns.

- **/cmd/server**: The main entry point of the application. It initializes dependencies and starts the HTTP server.
- **/internal/api**: Contains HTTP handlers and the router (`go-chi`). This layer handles request/response logic.
- **/internal/service**: Holds the core business logic. Services coordinate tasks between the API, repository, and external services.
- **/internal/repository**: Manages all data persistence. It provides a `Repository` interface, with a concrete implementation for **SQLite**.
- **/internal/database**: Contains logic for initializing the SQLite connection and schema.
- **/internal/llm**: Abstracts communication with Ollama via an `LLMProvider` interface.
- **/internal/model**: Defines the core data structures (`Chat`, `Message`).
- **/internal/config**: Handles loading of bootstrap configuration (e.g., Ollama URL, database path) from environment variables and a fallback `config.json`.

## Core Features & Logic

- **Real-time Chat Streaming**: Uses Server-Sent Events (SSE) to stream responses from the LLM to the client.
- **Chat Persistence**: Saves all chat history, messages, and conversation branching metadata in a local SQLite database. The schema is designed to support future features like message regeneration.
- **Dynamic Title Generation**: Automatically creates a concise title for new conversations by prompting a support model with a strict JSON format requirement. The service includes "smart parsing" logic to extract the JSON even from noisy model outputs.
- **Full Model Management**: Provides a complete API to list, pull, and delete Ollama models.
- **Dynamic Configuration**: On first launch, the `SettingsService` discovers available Ollama models, selects the most recent one as a default, and saves this configuration to the database. Subsequent launches read settings from the DB.
- **Per-Request Generation Parameters**: Allows clients to specify LLM options (like temperature, seed) on a per-request basis.

## Running Locally (Without Docker)

While the primary method is using Docker Compose, you can run the backend as a standalone service for development.

1.  Ensure you have Go installed and an Ollama instance running.
2.  Set the required environment variables:
    ```sh
    export DATABASE_PATH="./flow-dev.db"
    export OLLAMA_URL="http://localhost:11434"
    ```
3.  Run the application from the root `backend` directory:
    ```sh
    go run ./cmd/server/main.go
    ```
