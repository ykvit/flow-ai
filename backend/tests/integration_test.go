package tests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	baseAPIURL  = "http://localhost:8088/api"
	testModel   = "gemma3:270m-it-qat"
	composeFile = "compose.test.yaml"
)

// --- Test Main Setup & Teardown ---

func TestMain(m *testing.M) {
	// Using t.Logf is not possible in TestMain, so we use standard log here.
	fmt.Println("--- Setting up test environment ---")

	cmdUp := exec.Command("docker", "compose", "-f", composeFile, "up", "-d", "--build")
	if err := runCommand(cmdUp); err != nil {
		fmt.Printf("Failed to start docker compose: %v. Cleaning up...\n", err)
		cleanup()
		os.Exit(1)
	}

	if err := waitForBackend(); err != nil {
		fmt.Printf("Backend not ready: %v. Cleaning up.\n", err)
		cleanup()
		os.Exit(1)
	}
	fmt.Println("Backend is ready.")

	if err := pullTestModel(); err != nil {
		fmt.Printf("Failed to pull test model: %v. Cleaning up.\n", err)
		cleanup()
		os.Exit(1)
	}
	fmt.Printf("Test model '%s' pulled successfully.\n", testModel)

	exitCode := m.Run()

	fmt.Println("--- Tearing down test environment ---")
	cleanup()

	os.Exit(exitCode)
}

// --- Test Suites ---

func TestFullChatWorkflow(t *testing.T) {
	var chatID string
	initialContent := "What is the result of 10+10?" // Use a different simple question

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
	// 1. Setup: Use a more creative prompt to ensure different responses.
	resp, err := createMessage(t, "", "Suggest a name for a pet robot.")
	if err != nil {
		t.Fatal(err)
	}
	drainStream(t, resp.Body)

	var chatID string
	var initialAssistantMessage messageTest

	// Wait for the chat to be created.
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

	// Wait for the second assistant message.
	requireCondition(t, 20*time.Second, "second assistant message", func() bool {
		chat := getFullChat(t, chatID)
		if len(chat.Messages) == 4 {
			initialAssistantMessage = chat.Messages[3]
			return true
		}
		return false
	})

	t.Logf("Initial assistant message ID: %s, Length: %d", initialAssistantMessage.ID, len(initialAssistantMessage.Content))

	// 2. Action: Regenerate the last message without a seed.
	url := fmt.Sprintf("%s/chats/%s/messages/%s/regenerate", baseAPIURL, chatID, initialAssistantMessage.ID)
	reqBody := `{}` // No options, let it be random
	regenResp, err := http.Post(url, "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Regenerate request failed: %v", err)
	}
	drainStream(t, regenResp.Body)

	// 3. Verification: Check that the message ID has changed.
	var regeneratedAssistantMessage messageTest
	requireCondition(t, 20*time.Second, "message regeneration", func() bool {
		chat := getFullChat(t, chatID)
		// The key check: the last message ID must be different.
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

	// 4. Cleanup: Delete the chat.
	req, _ := http.NewRequest(http.MethodDelete, baseAPIURL+"/chats/"+chatID, nil)
	http.DefaultClient.Do(req)
}

// --- Helper Functions & Types ---

type messageTest struct{ ID, Content, Role string }
type fullChatTest struct{ ID, Title string; Messages []messageTest }
type chatInfoTest struct{ ID string }

// requireCondition is a generic helper to wait for a state to be true.
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
func runCommand(cmd *exec.Cmd) error {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("could not find project root: %w", err)
	}
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func getProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Join(wd, "..", ".."))
}
func cleanup() {
	cmdDown := exec.Command("docker", "compose", "-f", composeFile, "down", "-v")
	if err := runCommand(cmdDown); err != nil {
		fmt.Printf("WARN: Failed to stop docker-compose: %v\n", err)
	}
}
func waitForBackend() error {
	client := &http.Client{Timeout: 3 * time.Second}
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		resp, err := client.Get(baseAPIURL + "/settings")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
	}
	return fmt.Errorf("backend did not become ready in time")
}
func pullTestModel() error {
	pullReq := map[string]string{"name": testModel}
	body, _ := json.Marshal(pullReq)
	resp, err := http.Post(baseAPIURL+"/models/pull", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send pull request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("model pull returned non-200 status: %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
	}
	return scanner.Err()
}