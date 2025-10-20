package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOllamaProvider is a unit test for our Ollama HTTP client implementation.
//
// GOAL: To verify that our `ollamaProvider` correctly constructs and sends HTTP
// requests to the Ollama API, and correctly parses the responses.
//
// TECHNIQUE: We use the `net/http/httptest` package to create a mock HTTP server.
// This server acts as a stand-in for the real Ollama API. It listens on a local
// port and allows us to define exactly how it should respond to incoming requests.
// This lets us test our client's logic in complete isolation, without making any
// real network calls, making the test fast, reliable, and free of external dependencies.
func TestOllamaProvider(t *testing.T) {
	// These variables will capture the details of the HTTP request received by our mock server.
	var capturedMethod, capturedPath string

	// Create a new mock HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inside this handler function, we define the server's behavior.
		// We capture the method and path of the incoming request for later assertions.
		capturedMethod = r.Method
		capturedPath = r.URL.Path

		// A simple router to handle different Ollama API endpoints.
		switch r.URL.Path {
		case "/api/delete":
			// For a DELETE request, Ollama returns a 200 OK with no body on success.
			w.WriteHeader(http.StatusOK)
		case "/api/show":
			// For a "show" request, it returns a JSON object.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"modelfile": "FROM scratch"}`))
			assert.NoError(t, err) // It's good practice to check errors even in test helpers.
		default:
			// If our client tries to access an unknown endpoint, we return a 404.
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	// `defer server.Close()` ensures the mock server is shut down cleanly after the test completes.
	defer server.Close()

	// ARRANGE: Create an instance of our ollamaProvider, pointing it to the URL
	// of our mock server instead of a real Ollama instance.
	provider := NewOllamaProvider(server.URL)
	ctx := context.Background()

	t.Run("DeleteModel", func(t *testing.T) {
		// ACT: Call the method we want to test.
		err := provider.DeleteModel(ctx, &DeleteModelRequest{Name: "test-model"})

		// ASSERT:
		// 1. Check that the method completed without an error.
		require.NoError(t, err)
		// 2. Verify that our client sent a request with the correct HTTP method and path.
		assert.Equal(t, http.MethodDelete, capturedMethod)
		assert.Equal(t, "/api/delete", capturedPath)
	})

	t.Run("ShowModelInfo", func(t *testing.T) {
		// ACT
		info, err := provider.ShowModelInfo(ctx, &ShowModelRequest{Name: "test-model"})

		// ASSERT:
		// 1. Check for no errors and that a valid response object was returned.
		require.NoError(t, err)
		require.NotNil(t, info)
		// 2. Verify that the JSON response from the server was correctly parsed into the struct.
		assert.Equal(t, "FROM scratch", info.Modelfile)
		// 3. Verify that the correct HTTP method and path were used.
		assert.Equal(t, http.MethodPost, capturedMethod)
		assert.Equal(t, "/api/show", capturedPath)
	})
}
