# Flow-AI: A Self-Hosted Web UI for Ollama

Flow-AI is a modern, responsive, and feature-rich web interface for**** interacting with local Large Language Models through the Ollama service. It is designed to be a complete, self-contained solution that you can run on your own hardware with a single command.

The project consists of a high-performance Go backend and a (planned) reactive frontend, all orchestrated with Docker.

## Features

- **Real-time Chat**: Get streaming responses from your LLM for a smooth, natural conversation.
- **Full Chat History**: All your conversations are saved and can be revisited anytime.
- **Automatic Conversation Titling**: New chats are automatically named based on their content.
- **Complete Model Management**: List, pull, inspect, and delete Ollama models directly from the UI.
- **Dynamic Parameters**: Fine-tune model behavior (like creativity/temperature) for each individual message.
- **Response Metrics**: View detailed statistics for each AI response, such as generation time and token counts.

## Tech Stack

- **Backend**: Go with Chi (for routing) and go-redis.
- **Database**: Redis (for all data persistence).
- **LLM Service**: Ollama.
- **Frontend (Legacy)**: A placeholder React app. A new version is planned.
- **Orchestration**: Docker & Docker Compose.
- **Proxy**: NGINX (serves the frontend and proxies API requests).

## Getting Started

### Prerequisites

- Docker and Docker Compose must be installed on your machine.
- If you have an NVIDIA GPU, ensure the NVIDIA Container Toolkit is installed for GPU acceleration.
- An `ollama` Docker volume should exist to persist your models. If it doesn't, create it: `docker volume create ollama`.

### Installation

1.  **Clone the repository:**
    ```sh
    git clone https://your-repo-url/flow-ai.git
    cd flow-ai
    ```

2.  **Run with Docker Compose:**
    This command will build the frontend and backend, and start all the necessary services (NGINX, Go Backend, Redis, Ollama).
    ```sh
    docker compose up --build -d
    ```
    The `-d` flag runs the containers in detached mode.

3.  **Access the Application:**
    Open your web browser and navigate to `http://localhost:3000`.

## Project Structure

```
/
├── backend/          # Go backend source code
├── frontend/         # Frontend source code (React)
├── app/              # Legacy frontend (to be phased out)
├── k6/               # k6 performance test scripts
├── compose.yaml      # Docker Compose file for orchestration
├── Dockerfile        # Multi-stage Dockerfile for building the app
├── nginx.conf        # NGINX configuration
└── README.md         # This file
```

## API Documentation

The backend provides a comprehensive RESTful API for all its features. This is the primary interface for the frontend application.

For detailed specifications, endpoint descriptions, and examples, please see the **[API Documentation](API.md)**.

## Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request.