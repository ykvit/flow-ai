# Flow-AI: An Open-Source WebUI for LLMs

Flow-AI is a powerful, self-hostable, and user-friendly web interface for interacting with local Large Language Models (LLMs) through Ollama. It is designed to be a lightweight and feature-rich alternative to other popular WebUIs.


## ✨ Core Features

- **Real-time Chat:** Get streaming responses from your local models instantly.
- **SQLite Powered:** No external database needed. All chat history is stored in a simple, persistent SQLite database.
- **Dynamic Model Management:** Pull, delete, and inspect Ollama models directly from the UI.
- **Dynamic Configuration:** On first launch, the app intelligently detects your latest Ollama model and sets it as the default.
- **Per-Request Options:** Customize generation parameters like temperature, seed, and system prompts for each message.
- **Containerized & Simple Deployment:** Get up and running in minutes with a single `docker-compose up` command.

## 🛠️ Tech Stack

- **Backend:** Go (Golang)
  - **Web Framework:** Chi
  - **Database:** SQLite
  - **LLM Engine:** [Ollama](https://ollama.com/)
- **Frontend:** (To be developed - currently a placeholder)
- **Deployment:** Docker & Docker Compose

## 🚀 Getting Started

### Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- NVIDIA GPU with drivers (for GPU acceleration with Ollama)

### Installation

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/ykvit/flow-ai.git
    cd flow-ai
    ```

2.  **Build and run the application:**
    ```sh
    docker compose up --build
    ```
    The first launch might take a few minutes as Docker builds the images.

3.  **Access the application:**
    - The frontend will be available at `http://localhost:3000`.
    - The backend API is accessible at `http://localhost:8000/api`.

### Post-Installation: Pull a Model

On the first launch, your Ollama instance will be empty. You need to pull a model before you can start chatting. You can do this via the API.

**Example: Pulling the `gemma3:1b` model:**
```sh
curl -N -X POST http://localhost:8000/api/models/pull \
-H "Content-Type: application/json" \
-d '{"name": "gemma3:1b"}'
```

After the model is downloaded, the application will automatically use it as the default.

## 📚 Documentation

- **[Backend Architecture](backend/backend.md):** A detailed explanation of the backend design and key decisions.
- **[API Specification](API.md):** Complete documentation for the backend API endpoints.

## ⚙️ Configuration

The application uses a hierarchical configuration system, making it flexible for both Docker and local development.

**The order of priority is:**

1.  **Environment Variables (Highest Priority):** This is the primary method for configuring the application in Docker. Variables are set in the `compose.yaml` file (e.g., `OLLAMA_URL`, `DATABASE_PATH`).

2.  **`config.json` file (Fallback):** If an environment variable is not set, the application will look for the corresponding value in the `backend/config.json` file. This file is included in the repository to provide a seamless setup for local development without Docker.

3.  **Default Values (Lowest Priority):** If a setting is found in neither of the above, a hardcoded default value will be used.

This means that when you run `docker compose up`, the settings from `compose.yaml` will **always** be used, regardless of what is in `config.json`.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
