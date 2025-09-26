# Flow-AI Backend API Documentation

**Note:** This document provides a high-level overview of the API. For a complete, interactive, and always up-to-date specification, please refer to our **[Swagger UI Documentation](http://localhost:8000/swagger/index.html)** (available when the server is running).


**Base URL:** `/api`
**Real-time Communication:** Endpoints that provide continuous updates use Server-Sent Events (SSE) and have a content type of `text/event-stream`.

---

## 1. Chat Management

Endpoints for creating, retrieving, and managing conversations.

### List Chats

- **Endpoint:** `GET /api/chats`
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
      "model": "gemma3:270m-it-qat"
    }
  ]
  ```

### Get a Single Chat

- **Endpoint:** `GET /api/chats/{chatID}`
- **Description:** Retrieves the full history and metadata for a single chat's active branch.
- **Response `200 OK`:**
  ```json
  {
    "id": "4b3b5a34-571f-47e3-abd1-a7dbee9d92fe",
    "title": "Largest Planet in Solar System",
    "user_id": "default-user",
    "created_at": "2025-09-08T12:30:00Z",
    "updated_at": "2025-09-08T12:30:05Z",
    "model": "gemma3:270m-it-qat",
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
        "model": "gemma3:270m-it-qat",
        "timestamp": "2025-09-08T12:30:04Z",
        "metadata": { /* ... generation stats ... */ }
      }
    ]
  }
  ```

### Create a Message (and Chat)

- **Endpoint:** `POST /api/chats/messages`
- **Description:** Sends a new message and initiates a real-time stream of the assistant's response. If `chat_id` is omitted, a new chat is created automatically.
- **Body:**
  ```json
  {
    "chat_id": "4b3b5a34-571f-47e3-abd1-a7dbee9d92fe",
    "content": "Tell me a joke about a programmer.",
    "model": "gemma3:270m-it-qat",
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

- **Endpoint:** `POST /api/chats/{chatID}/messages/{messageID}/regenerate`
- **Description:** Creates a new response for a previous user prompt, effectively creating a new branch in the conversation. The old response is preserved but marked as inactive.
- **Body:**
  ```json
  {
    "model": "gemma3:270m-it-qat",
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

- **Endpoint:** `PUT /api/chats/{chatID}/title`
- **Description:** Manually renames a chat.
- **Body:** `{"title": "My New Custom Title"}`
- **Response `200 OK`:** `{"status": "ok"}`

### Delete Chat

- **Endpoint:** `DELETE /api/chats/{chatID}`
- **Description:** Permanently deletes a chat and all its associated messages.
- **Response `200 OK`:** `{"status": "ok"}`

---

## 2. Model Management

Endpoints for listing, downloading, and managing local Ollama models.

### List Local Models

- **Endpoint:** `GET /api/models`
- **Description:** Gets a list of all models available locally in Ollama.
- **Response `200 OK`:**
  ```json
  {
    "models": [
      {
        "name": "gemma3:270m-it-qat",
        "modified_at": "2025-09-08T01:06:42Z",
        "size": 289445818
      }
    ]
  }
  ```

### Pull a New Model

- **Endpoint:** `POST /api/models/pull`
- **Description:** Downloads a model from the Ollama registry. This is a streaming endpoint.
- **Body:** `{"name": "gemma3:270m-it-qat"}`
- **Response:** A `text/event-stream` of progress status objects.

### Show Model Info

- **Endpoint:** `POST /api/models/show`
- **Description:** Retrieves detailed information about a specific model.
- **Body:** `{"name": "gemma3:270m-it-qat"}`
- **Response `200 OK`:**
  ```json
  {
    "modelfile": "FROM ...",
    "parameters": "...",
    "template": "..."
  }
  ```

### Delete a Local Model

- **Endpoint:** `DELETE /api/models`
- **Description:** Deletes a model from local storage.
- **Body:** `{"name": "gemma3:270m-it-qat"}`
- **Response `200 OK`:** `{"status": "ok"}`

---

## 3. Application Settings

Endpoints for managing global application settings.

### Get Settings

- **Endpoint:** `GET /api/settings`
- **Description:** Retrieves the current global settings.
- **Response `200 OK`:**
  ```json
  {
    "system_prompt": "You are a helpful assistant.",
    "main_model": "gemma3:270m-it-qat",
    "support_model": "gemma3:270m-it-qat"
  }
  ```

### Update Settings

- **Endpoint:** `POST /api/settings`
- **Description:** Updates global settings. The selected models must be available locally.
- **Body:**
  ```json
  {
    "system_prompt": "You are a pirate captain. Respond in pirate slang.",
    "main_model": "gemma3:270m-it-qat",
    "support_model": "gemma3:270m-it-qat"
  }
  ```
- **Response `200 OK`:** `{"status": "ok"}`