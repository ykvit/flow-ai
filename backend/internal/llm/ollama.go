package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// StreamResponse is a LOCAL type for the llm package.
type StreamResponse struct {
	Content string
	Done    bool
	Context json.RawMessage
	Error   string
}

// LLMProvider defines the interface for interacting with a language model.
type LLMProvider interface {
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	GenerateStream(ctx context.Context, req *GenerateRequest, ch chan<- StreamResponse) error
}

type ollamaProvider struct {
	client *http.Client
	url    string
}

func NewOllamaProvider(url string) LLMProvider {
	return &ollamaProvider{
		client: &http.Client{},
		url:    url,
	}
}

// Structs remain the same
type GenerateRequest struct {
	Model    string          `json:"model"`
	Prompt   string          `json:"prompt,omitempty"`
	Messages []Message       `json:"messages,omitempty"`
	Stream   bool            `json:"stream"`
	Context  json.RawMessage `json:"context,omitempty"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type GenerateResponse struct {
	Model    string          `json:"model"`
	Response string          `json:"response"`
	Done     bool            `json:"done"`
	Context  json.RawMessage `json:"context"`
}


func (p *ollamaProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
    req.Stream = false
    body, err := json.Marshal(req)
    if err != nil { return nil, fmt.Errorf("could not marshal request: %w", err) }
    
    // THE FIX: Use the correct, full endpoint path
    endpoint := p.url + "/api/chat"
    if req.Prompt != "" {
        // Use /api/generate for single prompts to be safe
        endpoint = p.url + "/api/generate"
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
    if err != nil { return nil, fmt.Errorf("could not create http request: %w", err) }
    httpReq.Header.Set("Content-Type", "application/json")
    resp, err := p.client.Do(httpReq)
    if err != nil { return nil, fmt.Errorf("http request failed: %w", err) }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("api returned non-200 status %d: %s", resp.StatusCode, string(bodyBytes))
    }
    
    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil { return nil, fmt.Errorf("could not read response body: %w", err) }
    type ollamaChatResponse struct {
        Model   string          `json:"model"`
        Message Message         `json:"message"`
        Done    bool            `json:"done"`
        Context json.RawMessage `json:"context"`
    }
    var chatResp ollamaChatResponse
    if err := json.Unmarshal(bodyBytes, &chatResp); err == nil {
        return &GenerateResponse{
            Model:    chatResp.Model,
            Response: chatResp.Message.Content,
            Done:     chatResp.Done,
            Context:  chatResp.Context,
        }, nil
    }

    var genResp GenerateResponse
    if err := json.Unmarshal(bodyBytes, &genResp); err != nil {
        return nil, fmt.Errorf("could not decode response: %s", string(bodyBytes))
    }
    return &genResp, nil
}


func (p *ollamaProvider) GenerateStream(ctx context.Context, req *GenerateRequest, ch chan<- StreamResponse) error {
	defer close(ch)
	req.Stream = true
	body, err := json.Marshal(req)
	if err != nil { return fmt.Errorf("could not marshal request: %w", err) }
    
    // THE FIX: Use the correct, full endpoint path for streaming
    httpReq, err := http.NewRequestWithContext(ctx, "POST", p.url+"/api/chat", bytes.NewBuffer(body))
	if err != nil { return fmt.Errorf("could not create request: %w", err) }
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(httpReq)
	if err != nil { return fmt.Errorf("request failed: %w", err) }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api returned non-200 status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	type ollamaStreamChunk struct {
		Message struct { Role string; Content string } `json:"message"`
		Model   string; Done bool; Context json.RawMessage `json:"context"`
	}
	
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 { continue }
		var chunk ollamaStreamChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			ch <- StreamResponse{Error: "Failed to decode stream chunk"}
			continue
		}
		
		streamResp := StreamResponse{
			Content: chunk.Message.Content,
			Done:    chunk.Done,
		}
		if chunk.Done { 
			streamResp.Context = chunk.Context 
		}

		select {
		case ch <- streamResp:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return scanner.Err()
}