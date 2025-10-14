# Flow-AI: An Open-Source, Conversation-Driven WebUI for LLMs

Flow-AI is a powerful, self-hostable, and user-friendly web interface for interacting with local Large Language Models (LLMs) through Ollama. It is designed to be a lightweight, feature-rich, and conversation-focused alternative to other popular WebUIs.

## ✨ Core Features

-   **Advanced Conversations:** Go beyond simple chat. **Regenerate** assistant responses to explore different answers and create **branches** in your conversation.
-   **Real-time Chat:** Get streaming responses from your local models instantly.
-   **SQLite Powered:** No external database needed. All chat history, including every conversation branch, is stored in a simple, persistent SQLite database.
-   **Full Model Management:** Pull, delete, and inspect Ollama models directly from the UI.
-   **Smart Configuration:** On first launch, the app intelligently detects your latest Ollama model and sets it as the default.
-   **Containerized & Simple Deployment:** Get up and running in minutes with a single `make` command.
-   **Live API Docs:** Interactive API documentation is available via Swagger UI.

## 🛠️ Tech Stack

-   **Backend:** Go (1.23+)
    -   **Web Framework:** Chi
    -   **Database:** SQLite (with WAL mode)
    -   **API Documentation:** OpenAPI (Swagger) via `swaggo`
-   **LLM Engine:** [Ollama](https://ollama.com/)
-   **Frontend:** React (Vite)
    -   **UI:** Material Web Components (`@material/web`)
    -   **State Management:** Zustand
-   **Deployment:** Docker & Docker Compose, managed via `Makefile`.

---

## 🚀 Getting Started

This guide will help you launch the application quickly and reliably.

### Prerequisites

-   [Docker](https://www.docker.com/get-started) & [Docker Compose](https://docs.docker.com/compose/install/)
-   `make` command-line tool
-   For GPU acceleration: An NVIDIA GPU with appropriate drivers installed.

### Installation

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/ykvit/flow-ai.git
    cd flow-ai
    ```

2.  **Create your configuration file:**
    Copy the example configuration file. You can customize it later if needed.
    ```sh
    cp .env.example .env
    ```

3.  **(First Time Only) Create the Ollama volume:**
    This external volume stores your downloaded LLM models, keeping them safe even if you remove the Flow-AI containers.
    ```sh
    docker volume create ollama
    ```

4.  **Build and run the application:**
    This command builds the optimized production images and starts the containers in the background.
    ```sh
    make prod
    ```
    The first launch might take a few minutes as Docker downloads and builds the images.

5.  **Access the application:**
    You're all set! Open your browser and navigate to:
    -   **Web Interface:** `http://localhost:3000` (or the port set in your `.env` file)
    -   **API Base URL:** `http://localhost:3000/api/v1/`

---

## 🧑‍💻 Development

This section is for developers who want to contribute to the project.

### Running the Development Environment

1.  Ensure you have completed steps 1, 2, and 3 from the "Getting Started" guide.
2.  Run the following command:
    ```sh
    make dev
    ```
    This will build the necessary images and start the containers with logs attached to your terminal.

### Accessing Services in Development

-   **Frontend (with hot-reload):** `http://localhost:5173`
-   **Backend API:** `http://localhost:8000/api/v1/`
-   **Ollama API:** `http://localhost:11434`

### Useful Makefile Commands

Our `Makefile` provides simple commands for all common tasks. Run `make` or `make help` to see a full, categorized list of available commands.

#### Main Commands

| Command | Description |
| :--- | :--- |
| `make dev` | 🚀 Starts the development environment (CPU). |
| `make prod` | 🚢 Starts the production environment (CPU). |
| `make logs` | 📜 Tails the logs from the running development environment. |

#### Testing

| Command | Description |
| :--- | :--- |
| `make test-backend` | 🧪 Runs backend integration tests and generates a coverage report. |
| `make test-frontend` | 🧪 Runs frontend tests (structure is ready). |

#### GPU-Accelerated Commands

| Command | Description |
| :--- | :--- |
| `make dev-gpu` | 🚀 Starts the development environment with GPU support. |
| `make prod-gpu`| 🚢 Starts the production environment with GPU support. |

#### Code Quality & Docs

| Command | Description |
| :--- | :--- |
| `make swag` | 📄 Regenerates the Swagger/OpenAPI documentation. |
| `make lint` | 🔍 Runs the Go linter (`golangci-lint`) to check code quality. |
| `make format`| ✨ Automatically formats all Go source code using `goimports`. |

#### Cleanup Commands

| Command | Description |
| :--- | :--- |
| `make down-dev` | ⏹️ Stops and cleans up only the **DEV** environment. |
| `make down-prod`| ⏹️ Stops and cleans up only the **PROD** environment. |
| `make down` | ☢️ **Stops and cleans up ALL** project containers and volumes. |

#### CI-Specific Commands

| Command | Description |
| :--- | :--- |
| `make test-ci`| 🤖 Runs backend tests for CI (no cache, clean build). |

#### Database Migrations

| Command | Description |
| :--- | :--- |
| `make migrate-create` | 📦 Creates new SQL migration files. Usage: `make migrate-create name=...` |
| `make migrate-up` | 📈 Applies all pending database migrations. |
| `make migrate-down` | 📉 Rolls back the last applied database migration. |

## 📚 API Documentation

The backend includes interactive API documentation powered by Swagger UI. It's the best way to explore and test the API endpoints.

-   **How to access:** Once the application is running, go to:
    **[http://localhost:3000/api/swagger/index.html](http://localhost:3000/api/swagger/index.html)**

For a high-level overview of the API, see [API.md](./API.md). For a detailed explanation of the project architecture, see [DOCUMENTATION.md](./DOCUMENTATION.md).

## 🎯 Project Goals

The goal of Flow-AI is to provide an interface that treats conversations not as a simple linear log, but as a **tree of possibilities**. We aim to give users the tools to explore, refine, and direct their interactions with LLMs in a way that feels intuitive and powerful.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.