package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"flow-ai/backend/internal/config"
)

// TestNewApp is a unit test for the application's main initialization logic.
//
// GOAL: The primary goal of this test is to verify that the entire dependency
// injection and component wiring process in `NewApp` can complete without panicking
// or returning an error. It's a high-level "smoke test" for the application's startup sequence.
//
// WHY IT'S IMPORTANT:
// Thanks to the refactoring of `app.go`, we can now test this critical path.
// This single test provides immense value by ensuring that all components
// (database, services, handlers) can be instantiated correctly. It catches a
// wide range of potential issues, such as incorrect dependency wiring, invalid
// initial SQL in migrations, or configuration problems.
func TestNewApp(t *testing.T) {
	// ARRANGE: Set up the minimal required external dependencies as mocks.

	// 1. Mock the Ollama Server.
	// `NewApp` contains a blocking call `waitForOllama`. To prevent this test
	// from hanging or relying on a real Ollama instance, we create a fake HTTP
	// server using `httptest` that immediately returns a 200 OK status.
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ollamaServer.Close() // Ensure the mock server is shut down after the test.

	// 2. Create a temporary database file.
	// Our application requires a real SQLite database file to initialize.
	// We create a temporary, empty file for this purpose. The file will be
	// automatically removed at the end of the test.
	dbFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Remove(dbFile.Name())) }()

	// 3. Create a minimal valid configuration.
	// We point our config to the mock Ollama server and the temporary DB file.
	cfg := &config.Config{
		DatabasePath: dbFile.Name(),
		OllamaURL:    ollamaServer.URL,
		LogLevel:     "DEBUG",
	}

	// ACT: Call the function we are testing.
	app, err := NewApp(cfg)

	// ASSERT: Verify the outcome.

	// The most important check: `NewApp` must not return an error.
	require.NoError(t, err)
	// As a sanity check, ensure the returned app object is not nil.
	require.NotNil(t, app)

	// Clean up the resources created by `NewApp`.
	// This is crucial to prevent leaking database connections.
	defer func() { require.NoError(t, app.DB.Close()) }()

	// Finally, assert that the core components within the App struct were initialized.
	assert.NotNil(t, app.DB)
	assert.NotNil(t, app.Server)
}
