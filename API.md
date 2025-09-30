# Flow-AI Backend API Documentation

**Note:** This document provides a high-level overview of the API. For a complete, interactive, and always up-to-date specification, please refer to our **[Swagger UI Documentation](http://localhost:3000/api/swagger/index.html)** (available when the server is running).


**Base URL:** `/api/v1`
**Real-time Communication:** Endpoints that provide continuous updates use Server-Sent Events (SSE) and have a content type of `text/event-stream`.

---

## 1. Chat Management

Endpoints for creating, retrieving, and managing conversations.

### List Chats

- **Endpoint:** `GET /api/v1/chats`
- **Description:** Retrieves a list of all chats, sorted by the most recently updated.
- **Response `200 OK`:**
  ```json
  [
    {
      "id": "4b3b5a34-571f-47e3-abd1-a7dbee9d92fe",
      "title": "Largest Planet in Solar System",
      "user_id": "default-user",
      "created_at": "2025-09-08T12:30:00Z",
      "updated_at": "2025-09-08T12:30:05Z",
      "model": "qwen3:4b"
    }
  ]
  ```

### Get a Single Chat

- **Endpoint:** `GET /api/v1/chats/{chatID}`
- **Description:** Retrieves the full history and metadata for a single chat's active branch.
- **Response `200 OK`:**
  ```json
  {
    "id": "4b3b5a34-571f-47e3-abd1-a7dbee9d92fe",
    "title": "Largest Planet in Solar System",
    "user_id": "default-user",
    "created_at": "2025-09-08T12:30:00Z",
    "updated_at": "2025-09-08T12:30:05Z",
    "model": "qwen3:4b",
    "messages": [
      {
        "id": "fc31c9fe-5475-4460-b18f-812d4e5a8466",
        "role": "user",
        "content": "What is the largest planet of the solar system?",
        "timestamp": "2025-09-08T12:30:00Z"
      },
      {
        "id": "fcb7658f-c9ec-4e7a-80cf-cfd165c7ce1f",
        "parent_id": "fc31c9fe-5475-4460-b18f-812d4e5a8466",
        "role": "assistant",
        "content": "The largest planet is Jupiter.",
        "model": "qwen3:4b",
        "timestamp": "2025-09-08T12:30:04Z",
        "metadata": { /* ... generation stats ... */ }
      }
    ]
  }
  ```

### Create a Message (and Chat)

- **Endpoint:** `POST /api/v1/chats/messages`
- **Description:** Sends a new message and initiates a real-time stream of the assistant's response. If `chat_id` is omitted, a new chat is created automatically.
- **Body:**
  ```json
  {
    "chat_id": "4b3b5a34-571f-47e3-abd1-a7dbee9d92fe",
    "content": "Tell me a joke about a programmer.",
    "model": "qwen3:4b",
    "options": {
      "temperature": 0.8,
      "seed": 42
    }
  }
  ```
- **Response:** A `text/event-stream` of JSON objects.
  ```
  data: {"content":"Why","done":false}
  ...
  data: {"content":"","done":true, "context": [...]}
  ```

### Regenerate a Message

- **Endpoint:** `POST /api/v1/chats/{chatID}/messages/{messageID}/regenerate`
- **Description:** Creates a new response for a previous user prompt, effectively creating a new branch in the conversation. The old response is preserved but marked as inactive.
- **Body:**
  ```json
  {
    "model": "qwen3:4b",
    "options": {
      "temperature": 0.9
    }
  }
  ```
- **URL Parameters:**
  - `chatID` (string, required): The ID of the chat.
  - `messageID` (string, required): The ID of the **assistant's message** you want to regenerate.
- **Response:** A `text/event-stream` of the new response, same format as creating a message.

### Update Chat Title

- **Endpoint:** `PUT /api/v1/chats/{chatID}/title`
- **Description:** Manually renames a chat.
- **Body:** `{"title": "My New Custom Title"}`
- **Response `200 OK`:** `{"status": "ok"}`

### Delete Chat

- **Endpoint:** `DELETE /api/v1/chats/{chatID}`
- **Description:** Permanently deletes a chat and all its associated messages.
- **Response `200 OK`:** `{"status": "ok"}`

---

## 2. Model Management

Endpoints for listing, downloading, and managing local Ollama models.

### List Local Models

- **Endpoint:** `GET /api/v1/models`
- **Description:** Gets a list of all models available locally in Ollama.
- **Response `200 OK`:**
  ```json
  {
    "models": [
      {
        "name": "qwen3:4b",
        "modified_at": "2025-09-08T01:06:42Z",
        "size": 4123456789
      }
    ]
  }
  ```

### Pull a New Model

- **Endpoint:** `POST /api/v1/models/pull`
- **Description:** Downloads a model from the Ollama registry. This is a streaming endpoint.
- **Body:** `{"name": "qwen3:4b"}`
- **Response:** A `text/event-stream` of progress status objects.

### Show Model Info

- **Endpoint:** `POST /api/v1/models/show`
- **Description:** Retrieves detailed information about a specific model.
- **Body:** `{"name": "qwen3:4b"}`
- **Response `200 OK`:**
  ```json
  {
    "modelfile": "FROM ...",
    "parameters": "...",
    "template": "..."
  }
  ```

### Delete a Local Model

- **Endpoint:** `DELETE /api/v1/models`
- **Description:** Deletes a model from local storage.
- **Body:** `{"name": "qwen3:4b"}`
- **Response `200 OK`:** `{"status": "ok"}`

---

## 3. Application Settings

Endpoints for managing global application settings.

### Get Settings

- **Endpoint:** `GET /api/v1/settings`
- **Description:** Retrieves the current global settings.
- **Response `200 OK`:**
  ```json
  {
    "system_prompt": "You are a helpful assistant.",
    "main_model": "qwen3:4b",
    "support_model": "gemma3:4b"
  }
  ```

### Update Settings

- **Endpoint:** `POST /api/v1/settings`
- **Description:** Updates global settings. The selected models must be available locally.
- **Body:**
  ```json
  {
    "system_prompt": "You are a pirate captain. Respond in pirate slang.",
    "main_model": "qwen3:4b",
    "support_model": "gemma3:4b"
  }
  ```
- **Response `200 OK`:** `{"status": "ok"}`
