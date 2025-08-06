# Flow-AI Backend API Documentation

This document provides a complete specification for the Flow-AI backend API. Frontend developers should use this guide to integrate with the backend services.

**Base URL:** All API endpoints are prefixed with `/api`. When running through the project's NGINX proxy, the full path will be `/backend-api/...`.

**Real-time Communication:** The API uses Server-Sent Events (SSE) for streaming chat responses and model download progress. The content type for these endpoints is `text/event-stream`.

---

## 1. Chat Management

Endpoints for creating, retrieving, and managing chats.

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
      "created_at": "2025-08-06T12:39:27.57Z",
      "updated_at": "2025-08-06T12:39:29.87Z",
      "model": "hf.co/unsloth/gemma-3-1b-it-GGUF:IQ4_NL"
    }
  ]
  ```

### Get a Single Chat

- **Endpoint:** `GET /api/chats/{chatID}`
- **Description:** Retrieves the full history and metadata for a single chat.
- **Response `200 OK`:**
  ```json
  {
    "id": "4b3b5a34-571f-47e3-abd1-a7dbee9d92fe",
    "title": "Largest Planet in Solar System",
    "model": "hf.co/unsloth/gemma-3-1b-it-GGUF:IQ4_NL",
    "messages": [
      {
        "id": "fc31c9fe-5475-4460-b18f-812d4e5a8466",
        "role": "user",
        "content": "What is the largest planet of the solar system?",
        "timestamp": "2025-08-06T12:39:27.57Z"
      },
      {
        "id": "fcb7658f-c9ec-4e7a-80cf-cfd165c7ce1f",
        "role": "assistant",
        "content": "The largest planet is Jupiter.",
        "timestamp": "2025-08-06T12:39:29.67Z",
        "metadata": {
          "total_duration": 2098615719,
          "prompt_eval_count": 36,
          "eval_count": 45
        }
      }
    ]
  }
  ```

### Create a Message (and Chat)

- **Endpoint:** `POST /api/chats/messages`
- **Description:** Sends a new message and initiates a real-time stream of the assistant's response. If `chat_id` is omitted, a new chat is created.
- **Body:**
  ```json
  {
    "chat_id": "4b3b5a34-571f-47e3-abd1-a7dbee9d92fe", // Omit to create a new chat
    "content": "Tell me a joke.",
    "model": "hf.co/unsloth/gemma-3-1b-it-GGUF:IQ4_NL",
    "system_prompt": "You are a helpful assistant.", // Default system prompt
    "options": { // Optional: override model parameters for this request
      "temperature": 0.8,
      "top_k": 50,
      "system": "You are a sarcastic assistant." // Per-request system prompt override
    }
  }
  ```
- **Response:** This is a `text/event-stream` response. The client will receive a series of JSON objects.
  ```
  data: {"content":"Why","done":false}
  data: {"content":" don't","done":false}
  data: {"content":" scientists","done":false}
  ...
  data: {"content":"","done":true, "context": [...]}
  ```

### Update Chat Title

- **Endpoint:** `PUT /api/chats/{chatID}/title`
- **Description:** Manually renames a chat.
- **Body:**
  ```json
  {
    "title": "My New Custom Title"
  }
  ```
- **Response `200 OK`:** `{"status": "ok"}`

### Delete Chat

- **Endpoint:** `DELETE /api/chats/{chatID}`
- **Description:** Permanently deletes a chat and all its messages.
- **Response `200 OK`:** `{"status": "ok"}`

---

## 2. Model Management
****
Endpoints for listing, downloading, and managing local Ollama models.

### List Local Models

- **Endpoint:** `GET /api/models`
- **Description:** Gets a list of all models available locally in Ollama.
- **Response `200 OK`:**
  ```json
  {
    "models": [
      {
        "name": "tinyllama:latest",
        "modified_at": "2025-08-06T01:06:42.34Z",
        "size": 637707251
      }
    ]
  }
  ```

### Pull a New Model

- **Endpoint:** `POST /api/models/pull`
- **Description:** Downloads a model from the Ollama registry. This is a streaming endpoint.
- **Body:**
  ```json
  {
    "name": "qwen:0.6b"
  }
  ```
- **Response:** A `text/event-stream` of progress status objects.
  ```
  data: {"status":"pulling manifest"}
  data: {"status":"pulling layer...", "digest":"...", "total":12345, "completed":100}
  ...
  data: {"status":"success"}
  ```

### Show Model Info

- **Endpoint:** `POST /api/models/show`
- **Description:** Retrieves detailed information about a specific model, including its Modelfile and parameters.
- **Body:**
  ```json
  {
    "name": "tinyllama:latest"
  }
  ```
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
- **Description:** Deletes a model from the local Ollama storage.
- **Body:**
  ```json
  {
    "name": "tinyllama:latest"
  }
  ```
- **Response `200 OK`:** `{"status": "ok"}`

---

## 3. Application Settings

### Get Settings

- **Endpoint:** `GET /api/settings`
- **Description:** Retrieves current global settings.
- **Response `200 OK`:**
  ```json
  {
    "redis_addr": "redis:6379",
    "ollama_url": "http://ollama:11434",
    "system_prompt": "You are a helpful assistant.",
    "main_model": "hf.co/unsloth/gemma-3-1b-it-GGUF:IQ4_NL",
    "support_model": "hf.co/unsloth/gemma-3-1b-it-GGUF:IQ4_NL"
  }
  ```

### Update Settings

- **Endpoint:** `POST /api/settings`
- **Description:** Updates global settings. Note: `redis_addr` and `ollama_url` cannot be updated at runtime.
- **Body:**
  ```json
  {
    "system_prompt": "You are a pirate captain.",
    "main_model": "qwen:0.6b"
  }
  ```
- **Response `200 OK`:** `{"status": "ok"}`
