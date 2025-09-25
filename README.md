# Flow-AI: An Open-Source WebUI for LLMs

Flow-AI is a powerful, self-hostable, and user-friendly web interface for interacting with local Large Language Models (LLMs) through Ollama. It is designed to be a lightweight and feature-rich alternative to other popular WebUIs.

## ✨ Core Features

- **Real-time Chat:** Get streaming responses from your local models instantly.
- **SQLite Powered:** No external database needed. All chat history is stored in a simple, persistent SQLite database.
- **Dynamic Model Management:** Pull, delete, and inspect Ollama models directly from the UI.
- **Dynamic Configuration:** On first launch, the app intelligently detects your latest Ollama model and sets it as the default.
- **Auto-Generated API Docs:** Interactive API documentation is available via Swagger UI.
- **Containerized & Simple Deployment:** Get up and running in minutes with a single `docker-compose up` command.

## 🛠️ Tech Stack

- **Backend:** Go (Golang)
  - **API Documentation:** OpenAPI (Swagger)
  - **Web Framework:** Chi
  - **Database:** SQLite
  - **LLM Engine:** [Ollama](https://ollama.ai/)
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

## 📚 API Documentation

**The backend includes interactive API documentation powered by Swagger UI.**

-   **How to access:** Once the application is running, open your web browser and go to:
    **[http://localhost:8000/swagger/index.html](http://localhost:8000/swagger/index.html)**

This interface allows you to explore all available endpoints, view their parameters and responses, and even execute test requests directly from your browser.

For a high-level overview of the API, you can also consult the static [API.md](./API.md) file. For a detailed explanation of the backend architecture, see [DOCUMENTATION.md](./DOCUMENTATION.md).

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
