package tests

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"flow-ai/backend/internal/api"
	"flow-ai/backend/internal/config"
	"flow-ai/backend/internal/database"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/repository"
	"flow-ai/backend/internal/service"
)

const (
	baseAPIURL        = "http://localhost:8000/api"
	ollamaInternalURL = "http://ollama:11434"
	testModel         = "gemma3:270m-it-qat"
	testDBPath        = "/tmp/flow-ai-test.db" // Use an in-memory or temp-file DB for tests
)

var testServer *http.Server

// TestMain now orchestrates the entire test lifecycle:
// 1. Sets up the server programmatically.
// 2. Runs it in a goroutine.
// 3. Waits for dependent services.
// 4. Pulls the test model.
// 5. Runs the tests.
// 6. Gracefully shuts down the server.
func TestMain(m *testing.M) {
	// Clean up any previous test database file.
	_ = os.Remove(testDBPath)

	fmt.Println("--- [TestMain] Setting up test environment ---")
	if err := setupTestServer(); err != nil {
		fmt.Printf("[TestMain] ERROR: Failed to set up test server: %v\n", err)
		os.Exit(1)
	}

	// Run the server in a background goroutine.
	go func() {
		fmt.Println("[TestMain] Starting in-process server...")
		if err := testServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[TestMain] ERROR: In-process server failed: %v\n", err)
			os.Exit(1)
		}
	}()

	if err := waitForServices(); err != nil {
		fmt.Printf("[TestMain] ERROR: Services did not become ready: %v\n", err)
		shutdownServer()
		os.Exit(1)
	}
	fmt.Println("[TestMain] All services are ready.")

	if err := pullTestModel(); err != nil {
		fmt.Printf("[TestMain] ERROR: Failed to pull test model: %v\n", err)
		shutdownServer()
		os.Exit(1)
	}
	fmt.Printf("[TestMain] Test model '%s' pulled successfully.\n", testModel)

	// Run all tests.
	exitCode := m.Run()

	// Teardown.
	shutdownServer()
	_ = os.Remove(testDBPath)

	os.Exit(exitCode)
}

// setupTestServer initializes all dependencies and creates an http.Server instance.
// This mirrors the logic from `internal/app/app.go` but is adapted for a test context.
func setupTestServer() error {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	bootstrapCfg, err := config.LoadBootstrapConfig()
	if err != nil {
		return fmt.Errorf("failed to load bootstrap config: %w", err)
	}

	// Override the database path for tests to ensure isolation.
	db, err := database.InitDB(testDBPath)
	if err != nil {
		return fmt.Errorf("failed to init test DB: %w", err)
	}

	repo := repository.NewSQLiteRepository(db)
	ollamaProvider := llm.NewOllamaProvider(bootstrapCfg.OllamaURL)
	settingsService := service.NewSettingsService(db, ollamaProvider)
	_, _ = settingsService.InitAndGet(context.Background(), bootstrapCfg.SystemPrompt)
	chatService := service.NewChatService(repo, ollamaProvider, settingsService)
	modelService := service.NewModelService(ollamaProvider)
	chatHandler := api.NewChatHandler(chatService, settingsService)
	modelHandler := api.NewModelHandler(modelService)
	router := api.NewRouter(chatHandler, modelHandler)

	testServer = &http.Server{
		Addr:    ":8000",
		Handler: router,
	}
	return nil
}

func shutdownServer() {
	fmt.Println("--- [TestMain] Tearing down test environment ---")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := testServer.Shutdown(ctx); err != nil {
		fmt.Printf("[TestMain] WARN: Server shutdown failed: %v\n", err)
	}
}

// --- Test Suites (Unchanged) ---
func TestFullChatWorkflow(t *testing.T) {
	var chatID string
	initialContent := "What is the result of 10+10?"

	t.Run("CreateNewChat", func(t *testing.T) {
		resp, err := createMessage(t, "", initialContent)
		if err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}
		defer resp.Body.Close()
		drainStream(t, resp.Body)
	})

	t.Run("WaitForChatCompletion", func(t *testing.T) {
		var chatIsComplete bool
		for i := 0; i < 60; i++ {
			chats := listChats(t)
			if len(chats) == 1 {
				chatID = chats[0].ID
				fullChat := getFullChat(t, chatID)

				titleIsGenerated := fullChat.Title != "" && fullChat.Title != truncate(initialContent, 50)
				if len(fullChat.Messages) == 2 && titleIsGenerated {
					t.Logf("Chat %s is complete: 2 messages saved and title updated to '%s'.", chatID, fullChat.Title)
					chatIsComplete = true
					break
				}
			}
			time.Sleep(500 * time.Millisecond)
		}

		if !chatIsComplete {
			t.Fatal("Timed out waiting for the chat to be fully processed.")
		}
	})

	t.Run("GetChatByID", func(t *testing.T) {
		if chatID == "" {
			t.Fatal("Chat ID not set from previous step")
		}
		fullChat := getFullChat(t, chatID)
		if len(fullChat.Messages) < 2 {
			t.Fatalf("Expected at least 2 messages, got %d", len(fullChat.Messages))
		}
	})

	t.Run("UpdateTitle", func(t *testing.T) {
		if chatID == "" {
			t.Fatal("Chat ID not set from previous step")
		}
		reqBody := `{"title": "Simple Math Question"}`
		req, _ := http.NewRequest(http.MethodPut, baseAPIURL+"/chats/"+chatID+"/title", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to update title: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200 for title update, got %d", resp.StatusCode)
		}
	})

	t.Run("DeleteChat", func(t *testing.T) {
		if chatID == "" {
			t.Fatal("Chat ID not set from previous step")
		}
		req, _ := http.NewRequest(http.MethodDelete, baseAPIURL+"/chats/"+chatID, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to delete chat: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200 for chat deletion, got %d", resp.StatusCode)
		}
	})

	t.Run("VerifyDeletion", func(t *testing.T) {
		chats := listChats(t)
		if len(chats) != 0 {
			t.Fatalf("Expected 0 chats after deletion, got %d", len(chats))
		}
	})
}

func TestRegenerationWorkflow(t *testing.T) {
	resp, err := createMessage(t, "", "Suggest a name for a pet robot.")
	if err != nil {
		t.Fatal(err)
	}
	drainStream(t, resp.Body)

	var chatID string
	var initialAssistantMessage messageTest

	requireCondition(t, 10*time.Second, "chat creation", func() bool {
		chats := listChats(t)
		if len(chats) == 1 {
			chatID = chats[0].ID
			return true
		}
		return false
	})

	resp, err = createMessage(t, chatID, "Give me another idea.")
	if err != nil {
		t.Fatal(err)
	}
	drainStream(t, resp.Body)

	requireCondition(t, 20*time.Second, "second assistant message", func() bool {
		chat := getFullChat(t, chatID)
		if len(chat.Messages) == 4 {
			initialAssistantMessage = chat.Messages[3]
			return true
		}
		return false
	})

	t.Logf("Initial assistant message ID: %s, Length: %d", initialAssistantMessage.ID, len(initialAssistantMessage.Content))

	url := fmt.Sprintf("%s/chats/%s/messages/%s/regenerate", baseAPIURL, chatID, initialAssistantMessage.ID)
	reqBody := `{}`
	regenResp, err := http.Post(url, "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Regenerate request failed: %v", err)
	}
	drainStream(t, regenResp.Body)

	var regeneratedAssistantMessage messageTest
	requireCondition(t, 20*time.Second, "message regeneration", func() bool {
		chat := getFullChat(t, chatID)
		if len(chat.Messages) == 4 && chat.Messages[3].ID != initialAssistantMessage.ID {
			regeneratedAssistantMessage = chat.Messages[3]
			return true
		}
		return false
	})

	t.Logf("Regenerated message ID: %s, Length: %d", regeneratedAssistantMessage.ID, len(regeneratedAssistantMessage.Content))

	if regeneratedAssistantMessage.Content == initialAssistantMessage.Content {
		t.Logf("Warning: Regenerated content was the same as the original. This can happen occasionally.")
	}

	req, _ := http.NewRequest(http.MethodDelete, baseAPIURL+"/chats/"+chatID, nil)
	http.DefaultClient.Do(req)
}
// --- Helper Functions & Types (Unchanged) ---
type messageTest struct{ ID, Content, Role string }
type fullChatTest struct{ ID, Title string; Messages []messageTest }
type chatInfoTest struct{ ID string }

func requireCondition(t *testing.T, timeout time.Duration, name string, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("Timed out waiting for condition: %s", name)
}

func createMessage(t *testing.T, chatID, content string) (*http.Response, error) {
	t.Helper()
	reqMap := map[string]string{"content": content, "model": testModel}
	if chatID != "" {
		reqMap["chat_id"] = chatID
	}
	body, _ := json.Marshal(reqMap)
	return http.Post(baseAPIURL+"/chats/messages", "application/json", bytes.NewBuffer(body))
}

func getFullChat(t *testing.T, chatID string) fullChatTest {
	t.Helper()
	resp, err := http.Get(baseAPIURL + "/chats/" + chatID)
	if err != nil {
		t.Fatalf("Failed to get chat by ID %s: %v", chatID, err)
	}
	defer resp.Body.Close()
	var chat fullChatTest
	if err := json.NewDecoder(resp.Body).Decode(&chat); err != nil {
		t.Fatalf("Failed to decode full chat: %v", err)
	}
	return chat
}

func listChats(t *testing.T) []chatInfoTest {
	t.Helper()
	resp, err := http.Get(baseAPIURL + "/chats")
	if err != nil {
		t.Fatalf("Failed to list chats: %v", err)
	}
	defer resp.Body.Close()
	var chats []chatInfoTest
	if err := json.NewDecoder(resp.Body).Decode(&chats); err != nil {
		t.Fatalf("Failed to decode chat list: %v", err)
	}
	return chats
}

func drainStream(t *testing.T, body io.ReadCloser) {
	t.Helper()
	scanner := bufio.NewScanner(body)
	foundDone := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") && strings.Contains(line, `"done":true`) {
			foundDone = true
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading stream: %v", err)
	}
	if !foundDone {
		t.Fatal("Stream finished without a 'done:true' message")
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

func waitForServices() error {
	services := map[string]string{
		"Backend": baseAPIURL + "/settings",
		"Ollama":  ollamaInternalURL,
	}
	client := &http.Client{Timeout: 2 * time.Second}

	for name, url := range services {
		fmt.Printf("Waiting for %s at %s...\n", name, url)
		ready := false
		for i := 0; i < 30; i++ {
			resp, err := client.Get(url)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				fmt.Printf("%s is ready.\n", name)
				ready = true
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(2 * time.Second)
		}
		if !ready {
			return fmt.Errorf("%s service did not become ready in time", name)
		}
	}
	return nil
}

func pullTestModel() error {
	fmt.Printf("Requesting backend to pull model: %s\n", testModel)
	pullReq := map[string]string{"name": testModel}
	body, _ := json.Marshal(pullReq)
	client := &http.Client{Timeout: 0}
	resp, err := client.Post(baseAPIURL+"/models/pull", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send pull request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("model pull returned non-200 status: %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}
	fmt.Println("Pulling model (this may take a while)...")
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading model pull stream: %w", err)
	}
	fmt.Println("Model pull stream finished.")
	return nil
}