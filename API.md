# Flow-AI Backend API Documentation

This document provides a complete specification for the Flow-AI backend API.

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
      "model": "qwen:0.5b"
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
    "model": "qwen:0.5b",
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
        "model": "qwen:0.5b",
        "timestamp": "2025-09-08T12:30:04Z",
        "metadata": {
          "total_duration": 2098615719,
          "prompt_eval_count": 36,
          "eval_count": 45,
          "eval_duration": 1500000000
        }
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
    "model": "tinyllama",
    "options": {
      "temperature": 0.8,
      "seed": 42,
      "top_k": 50,
      "top_p": 0.9,
      "system": "You are a witty assistant who loves puns."
    }
  }
  ```
- **Body Fields:**
  - `chat_id` (string, optional): The ID of the chat to add the message to. If empty, a new chat is created.
  - `content` (string, required): The user's message.
  - `model` (string, optional): Override the default model for this specific request.
  - `options` (object, optional): A key-value object of Ollama parameters to override for this request. See Ollama documentation for all possible options.
- **Response:** A `text/event-stream` of JSON objects.
  ```
  data: {"content":"Why","done":false}
  data: {"content":" do","done":false}
  ...
  data: {"content":"","done":true, "context": [...]}
  ```

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
        "name": "qwen:0.5b",
        "modified_at": "2025-09-08T01:06:42Z",
        "size": 637707251
      }
    ]
  }
  ```

### Pull a New Model

- **Endpoint:** `POST /api/models/pull`
- **Description:** Downloads a model from the Ollama registry. This is a streaming endpoint.
- **Body:** `{"name": "tinyllama"}`
- **Response:** A `text/event-stream` of progress status objects.
  ```
  data: {"status":"pulling manifest"}
  data: {"status":"pulling layer...", "digest":"...", "total":123456, "completed":1024}
  ...
  data: {"status":"success"}
  ```

### Show Model Info

- **Endpoint:** `POST /api/models/show`
- **Description:** Retrieves detailed information about a specific model, including its Modelfile and parameters.
- **Body:** `{"name": "qwen:0.5b"}`
- **Response `200 OK`:**
  ```json
  {
    "modelfile": "FROM ...",
    "parameters": "stop <|system|>\nstop <|user|>\n...",
    "template": "..."
  }
  ```

### Delete a Local Model

- **Endpoint:** `DELETE /api/models`
- **Description:** Deletes a model from local storage.
- **Body:** `{"name": "tinyllama"}`
- **Response `200 OK`:** `{"status": "ok"}`

---

## 3. Application Settings

Endpoints for managing global application settings. These settings are persisted in the database.

### Get Settings

- **Endpoint:** `GET /api/settings`
- **Description:** Retrieves the current global settings used by the application.
- **Response `200 OK`:**
  ```json
  {
    "system_prompt": "You are a helpful assistant.",
    "main_model": "qwen:0.5b",
    "support_model": "qwen:0.5b"
  }
  ```

### Update Settings

- **Endpoint:** `POST /api/settings`
- **Description:** Updates global settings. The selected models must be available in Ollama at the time of the request.
- **Body:**
  ```json
  {
    "system_prompt": "You are a pirate captain. Respond in pirate slang.",
    "main_model": "tinyllama",
    "support_model": "qwen:0.5b"
  }
  ```
- **Response `200 OK`:** `{"status": "ok"}`
- **Response `400 Bad Request`:** If a specified model is not found in Ollama.
  ```json
  {
    "error": "main model 'non-existent-model' is not available in Ollama"
  }
  ```