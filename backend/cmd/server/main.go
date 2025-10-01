package main

import (
	"os"

	"flow-ai/backend/internal/app"
)

// @title           Flow-AI API
// @version         0.0.1
// @description     This is the API server for the Flow-AI application. It provides endpoints for managing chats, models, and application settings.
//
// @contact.name   API Support
// @contact.url    https://github.com/ykvit/flow-ai/issues
//
// @license.name   MIT
// @license.url    https://opensource.org/licenses/MIT
//
// @BasePath  /api
//
// @tag.name        Chats
// @tag.description Endpoints for creating, retrieving, and managing conversations.
//
// @tag.name        Models
// @tag.description Endpoints for listing, downloading, and managing local Ollama models.
//
// @tag.name        Settings
// @tag.description Endpoints for managing global application settings.

func main() {
	// The main package is a thin wrapper around the app package,
	// making the core application logic importable and testable.
	os.Exit(app.Run())
}
