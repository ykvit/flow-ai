# Flow-AI Backend API Documentation

## Single Source of Truth

This document provides a high-level overview of the Flow-AI API. It is intended for quick reference only.

For a complete, interactive, and always up-to-date specification, please use our **Swagger UI Documentation**. It is the single source of truth for all API endpoints, models, and examples.

-   **Development URL:** [http://localhost:8000/api/swagger/index.html](http://localhost:8000/api/swagger/index.html)
-   **Production URL:** [http://localhost:3000/api/swagger/index.html](http://localhost:3000/api/swagger/index.html) (when running via Docker Compose in production mode)

---

## Overview

The API is structured around three main resources: **Chats**, **Models**, and **Settings**.

-   **Base URL for API v1:** `/api/v1`
-   **Real-time Communication:** Endpoints that provide continuous updates (like generating messages or pulling models) use Server-Sent Events (SSE) and have a `Content-Type` of `text/event-stream`.

### 1. Chats

This group of endpoints allows you to manage the entire lifecycle of a conversation. You can list all chats, retrieve a specific chat with its full message history, create new messages (which can also create a new chat), regenerate responses, and delete chats.

-   `GET /api/v1/chats` - List all chats.
-   `GET /api/v1/chats/{chatID}` - Get a single chat with all messages.
-   `POST /api/v1/chats/messages` - Create a new message (and optionally a new chat).
-   `DELETE /api/v1/chats/{chatID}` - Delete a chat.
-   ... and more. See Swagger UI for details.

### 2. Models

These endpoints are used to interact with the local Ollama models. You can list all installed models, pull new models from a registry, view detailed information about a model, and delete them to free up space.

-   `GET /api/v1/models` - List local models.
-   `POST /api/v1/models/pull` - Download a new model.
-   `DELETE /api/v1/models` - Delete a local model.
-   ... and more. See Swagger UI for details.

### 3. Settings

A simple set of endpoints to manage global application settings, such as the default system prompt and the main model to be used for conversations.

-   `GET /api/v1/settings` - Get current settings.
-   `POST /api/v1/settings` - Update settings.

---

For detailed information on request/response bodies, URL parameters, and to try out the API live, please refer to the **[Swagger UI Documentation](http://localhost:8000/api/swagger/index.html)**.