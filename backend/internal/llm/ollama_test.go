package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOllamaProvider(t *testing.T) {
	var capturedMethod, capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path

		switch r.URL.Path {
		case "/api/delete":
			w.WriteHeader(http.StatusOK)
		case "/api/show":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"modelfile": "FROM scratch"}`))
			assert.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	provider := NewOllamaProvider(server.URL)
	ctx := context.Background()

	t.Run("DeleteModel", func(t *testing.T) {
		err := provider.DeleteModel(ctx, &DeleteModelRequest{Name: "test-model"})
		require.NoError(t, err)
		assert.Equal(t, http.MethodDelete, capturedMethod)
		assert.Equal(t, "/api/delete", capturedPath)
	})

	t.Run("ShowModelInfo", func(t *testing.T) {
		info, err := provider.ShowModelInfo(ctx, &ShowModelRequest{Name: "test-model"})
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, "FROM scratch", info.Modelfile)
		assert.Equal(t, http.MethodPost, capturedMethod)
		assert.Equal(t, "/api/show", capturedPath)
	})
}
