package tests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func TestMain(m *testing.M) {
	log.Println("--- Setting up test environment ---")

	// Start Docker Compose
	cmdUp := exec.Command("docker", "compose", "-f", composeFile, "up", "-d", "--build")
	if err := runCommand(cmdUp); err != nil {
		log.Printf("Failed to start docker compose: %v. Cleaning up...", err)
		cleanup()
		os.Exit(1)
	}

	if err := waitForBackend(); err != nil {
		log.Printf("Backend not ready: %v. Cleaning up.", err)
		cleanup()
		os.Exit(1)
	}
	log.Println("Backend is ready.")

	if err := pullTestModel(); err != nil {
		log.Printf("Failed to pull test model: %v. Cleaning up.", err)
		cleanup()
		os.Exit(1)
	}
	log.Printf("Test model '%s' pulled successfully.", testModel)

	exitCode := m.Run()

	log.Println("--- Tearing down test environment ---")
	cleanup()

	os.Exit(exitCode)
}

func cleanup() {
	cmdDown := exec.Command("docker", "compose", "-f", composeFile, "down", "-v")
	if err := runCommand(cmdDown); err != nil {
		log.Printf("WARN: Failed to stop docker-compose: %v", err)
	}
}

func runCommand(cmd *exec.Cmd) error {
    // This helper now finds the project root to run docker-compose from the correct directory.
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
	wd, err := os.Getwd() // .../backend/tests
	if err != nil {
		return "", err
	}
	// Go up two levels to get the project root
	return filepath.Abs(filepath.Join(wd, "..", ".."))
}

func waitForBackend() error {
	client := &http.Client{Timeout: 3 * time.Second}
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second) // Wait before check
		resp, err := client.Get(baseAPIURL + "/settings")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if err != nil {
			log.Printf("Waiting for backend... attempt %d failed: %v", i+1, err)
		} else if resp != nil {
			log.Printf("Waiting for backend... attempt %d got status: %s", i+1, resp.Status)
			resp.Body.Close()
		}
	}
	return fmt.Errorf("backend did not become ready in time")
}

func pullTestModel() error {
	pullReq := map[string]string{"name": testModel}
	body, _ := json.Marshal(pullReq)
	resp, err := http.Post(baseAPIURL+"/models/pull", "application/json", bytes.NewBuffer(body))
	if err != nil { return fmt.Errorf("failed to send pull request: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("model pull returned non-200 status: %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}
	io.Copy(io.Discard, resp.Body)
	return nil
}


func TestFullChatWorkflow(t *testing.T) {
	var chatID string
	initialContent := "What is 2+2?"

	t.Run("CreateNewChat", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{"content": "%s", "model": "%s"}`, initialContent, testModel)
		resp, err := http.Post(baseAPIURL+"/chats/messages", "application/json", strings.NewReader(reqBody))
		if err != nil { t.Fatalf("Failed to create message: %v", err) }
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK { t.Fatalf("Expected status 200 for chat creation, got %d", resp.StatusCode) }

		scanner := bufio.NewScanner(resp.Body)
		foundDone := false
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data:") {
				if strings.Contains(line, `"done":true`) {
					foundDone = true
					break
				}
			}
		}
		if err := scanner.Err(); err != nil { t.Fatalf("Error reading stream: %v", err) }
		if !foundDone { t.Fatal("Stream finished without a 'done:true' message") }
	})

	t.Run("ListChatsAndWaitForTitle", func(t *testing.T) {
		var chatFound bool
		for i := 0; i < 20; i++ {
			resp, err := http.Get(baseAPIURL + "/chats")
			if err != nil { t.Fatalf("Failed to list chats: %v", err) }

			var chats []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&chats); err != nil {
				resp.Body.Close()
				t.Fatalf("Failed to decode chats list: %v", err)
			}
			resp.Body.Close()

			if len(chats) == 1 {
				chatID = chats[0]["id"].(string)
				title := chats[0]["title"].(string)
				// Check if title has been updated from the initial content
				if title != "" && title != truncate(initialContent, 50) {
					log.Printf("Found new chat with ID: %s and generated Title: %s", chatID, title)
					chatFound = true
					break
				}
			}
			time.Sleep(500 * time.Millisecond) // Wait before retrying
		}

		if !chatFound {
			t.Fatal("Timed out waiting for generated title. Chat was not found or title was not updated.")
		}
	})

	// ... The rest of the test functions remain the same
	t.Run("GetChatByID", func(t *testing.T) {
		if chatID == "" { t.Fatal("Chat ID not set from previous step") }

		resp, err := http.Get(baseAPIURL + "/chats/" + chatID)
		if err != nil {
			t.Fatalf("Failed to get chat by ID: %v", err)
		}
		defer resp.Body.Close()

		var fullChat map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&fullChat); err != nil {
			t.Fatalf("Failed to decode full chat: %v", err)
		}
		messages, ok := fullChat["messages"].([]interface{})
		if !ok || len(messages) < 2 {
			t.Fatalf("Expected at least 2 messages in the chat, got %d", len(messages))
		}
	})
    
	t.Run("UpdateTitle", func(t *testing.T) {
		if chatID == "" { t.Fatal("Chat ID not set from previous step") }
		
		reqBody := `{"title": "Simple Math Question"}`
		req, _ := http.NewRequest(http.MethodPut, baseAPIURL+"/chats/"+chatID+"/title", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		resp, err := http.DefaultClient.Do(req)
		if err != nil { t.Fatalf("Failed to update title: %v", err) }
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK { t.Fatalf("Expected status 200 for title update, got %d", resp.StatusCode) }
	})

	t.Run("DeleteChat", func(t *testing.T) {
		if chatID == "" { t.Fatal("Chat ID not set from previous step") }

		req, _ := http.NewRequest(http.MethodDelete, baseAPIURL+"/chats/"+chatID, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil { t.Fatalf("Failed to delete chat: %v", err) }
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK { t.Fatalf("Expected status 200 for chat deletion, got %d", resp.StatusCode) }
	})

	t.Run("VerifyDeletion", func(t *testing.T) {
		resp, err := http.Get(baseAPIURL + "/chats")
		if err != nil { t.Fatalf("Failed to list chats after deletion: %v", err) }
		defer resp.Body.Close()

		var chats []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&chats); err != nil { t.Fatalf("Failed to decode chats list: %v", err) }
		if len(chats) != 0 { t.Fatalf("Expected 0 chats after deletion, got %d", len(chats)) }
	})
}

// Copied from chat_service for the test
func truncate(s string, n int) string {
	if len(s) <= n { return s }; runes := []rune(s); if len(runes) <= n { return s }; return string(runes[:n])
}