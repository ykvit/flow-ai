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

func TestNewApp(t *testing.T) {
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ollamaServer.Close()

	dbFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Remove(dbFile.Name())) }()

	cfg := &config.Config{
		DatabasePath: dbFile.Name(),
		OllamaURL:    ollamaServer.URL,
		LogLevel:     "DEBUG",
	}

	app, err := NewApp(cfg)
	require.NoError(t, err)
	require.NotNil(t, app)

	defer func() { require.NoError(t, app.DB.Close()) }()

	assert.NotNil(t, app.DB)
	assert.NotNil(t, app.Server)
}
