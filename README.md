# Web UI for LLMs - Backend

This directory contains the Go backend for the custom Web UI application. It serves as an API server that communicates with a frontend client, a Redis database, and various Large Language Models (LLMs) like Ollama.

## Features (Version 1.0)

*   **RESTful API:** Provides endpoints for managing chats, messages, and settings.
*   **Real-time Streaming:** Uses Server-Sent Events (SSE) to stream responses from the LLM directly to the client for a real-time "typing" effect.
*   **Ollama Integration:** Connects to an Ollama instance to generate text.
*   **Persistent Chat History:** Stores all chat and message data in a Redis database.
*   **Automatic Chat Titles:** After the first exchange in a new chat, it automatically generates a concise title in a background process using a separate "support" model.
*   **Conversation Context:** Correctly manages and re-uses Ollama's `context` field to ensure conversations are continuous.
*   **Configuration Management:** Easily configured using a `config.json` file and environment variables.

## Tech Stack

*   **Language:** Go 1.22
*   **Database:** Redis
*   **API:** REST + Server-Sent Events (SSE)
*   **Containerization:** Docker & Docker Compose

## How to Run

1.  **Prerequisites:**
    *   Docker and Docker Compose must be installed.
    *   Make sure you have an Ollama model pulled, e.g., `docker exec -it my-webui-ollama-1 ollama pull llama3`.

2.  **Configuration:**
    *   Create a file `backend/config.json` with the following structure:
    ```json
    {
      "redis_addr": "redis:6379",
      "ollama_url": "http://ollama:11434",
      "system_prompt": "You are a helpful assistant.",
      "main_model": "llama3:latest",
      "support_model": "llama3:latest"
    }
    ```
    *   The values for `redis_addr` and `ollama_url` are service names from `compose.yaml` and should generally not be changed.

3.  **Build and Run:**
    *   From the root directory (`my-webui/`), run the command:
    ```bash
    docker-compose up --build
    ```
    *   The backend server will be available at `http://localhost:8080`.

## API Endpoints (MVP)

*   `GET /api/settings`: Get current application settings.
*   `POST /api/settings`: Update application settings.
*   `GET /api/chats`: Get a list of all chats (metadata only).
*   `GET /api/chats/{chatID}`: Get the full message history for a specific chat.
*   `POST /api/chats/messages`: Send a new message. This endpoint initiates an SSE stream for the response. It can be used for both creating a new chat (if `chat_id` is empty) and adding to an existing one.
    *   **Request Body:**
        ```json
        {
          "chat_id": "optional-uuid-of-existing-chat",
          "content": "Your message here",
          "model": "llama3:latest" 
        }
        ```
    *   **Response:** A stream of `text/event-stream` events.

## Project Structure

*   `cmd/server/main.go`: The main entry point of the application.
*   `internal/api/`: Contains HTTP handlers and router setup.
*   `internal/service/`: Contains the core business logic.
*   `internal/repository/`: Handles all communication with the database (Redis).
*   `internal/llm/`: Provides an abstraction layer for communicating with LLMs (Ollama).
*   `internal/model/`: Defines the core data structures (Chat, Message).
*   `internal/config/`: Manages application configuration.