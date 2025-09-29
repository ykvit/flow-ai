package main

import (
	"os"

	"flow-ai/backend/internal/app"
)

// @title           Flow-AI Backend API
// @version         0.1
// @description     This is the API server for the Flow-AI application, providing endpoints for chat, model management, and settings.
// @contact.name    API Support
// @contact.url     https://github.com/ykvit/flow-ai
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
// @BasePath        /api
func main() {
	// The main package is now just a thin wrapper around the app package.
	// This makes the core application logic importable and testable.
	os.Exit(app.Run())
}