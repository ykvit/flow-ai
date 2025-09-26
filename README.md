# Flow-AI: An Open-Source, Conversation-Driven WebUI for LLMs

Flow-AI is a powerful, self-hostable, and user-friendly web interface for interacting with local Large Language Models (LLMs) through Ollama. It is designed to be a lightweight, feature-rich, and conversation-focused alternative to other popular WebUIs.

## ✨ Core Features

- **Advanced Conversations:** Go beyond simple chat. **Regenerate** assistant responses to explore different answers and create **branches** in your conversation.
- **Real-time Chat:** Get streaming responses from your local models instantly.
- **SQLite Powered:** No external database needed. All chat history, including every conversation branch, is stored in a simple, persistent SQLite database.
- **Full Model Management:** Pull, delete, and inspect Ollama models directly from the UI.
- **Smart Configuration:** On first launch, the app intelligently detects your latest Ollama model and sets it as the default.
- **Containerized & Simple Deployment:** Get up and running in minutes with a single Docker Compose command.
- **Live API Docs:** Interactive API documentation is available via Swagger UI.

## 🎯 Project Goals

The goal of Flow-AI is to provide an interface that treats conversations not as a simple linear log, but as a **tree of possibilities**. We aim to give users the tools to explore, refine, and direct their interactions with LLMs in a way that feels intuitive and powerful.

## 🛠️ Tech Stack

- **Backend:** Go (1.22+)
  - **Web Framework:** Chi
  - **Database:** SQLite (with WAL mode)
  - **API Documentation:** OpenAPI (Swagger) via `swaggo`
- **LLM Engine:** [Ollama](https://ollama.com/)
- **Frontend:** (To be developed)
- **Deployment:** Docker & Docker Compose

## 🚀 Getting Started

### Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- For GPU acceleration: NVIDIA GPU with drivers.

### Installation

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/ykvit/flow-ai.git
    cd flow-ai
    ```

2.  **(First Time Only) Create the Ollama volume:**
    This external volume will store your downloaded LLM models, keeping them safe even if you remove the Flow-AI containers.
    ```sh
    docker volume create ollama
    ```

3.  **Build and run the application:**
    Choose the command that matches your hardware.

    - **For users with an NVIDIA GPU (Recommended):**
      ```sh
      docker compose -f compose.gpu.yaml up --build -d
      ```
    - **For CPU-only users:**
      ```sh
      docker compose up --build -d
      ```
    The `-d` flag runs the containers in the background. The first launch might take a few minutes as Docker builds the images.

4.  **Access the application:**
    - The frontend will be available at `http://localhost:3000`.
    - The backend API is accessible at `http://localhost:8000`.

## 📚 API Documentation

The backend includes interactive API documentation powered by Swagger UI. It's the best way to explore and test the API.

-   **How to access:** Once the application is running, open your browser and go to:
    **[http://localhost:8000/swagger/index.html](http://localhost:8000/swagger/index.html)**

For a high-level overview of the API, see [API.md](./API.md).
For a detailed explanation of the backend architecture, see [DOCUMENTATION.md](./DOCUMENTATION.md).

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.