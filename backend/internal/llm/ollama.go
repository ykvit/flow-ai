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

// --- NEW STRUCT for response stats ---
// GenerationStats holds the statistics returned by Ollama after generation.
type GenerationStats struct {
	TotalDuration      int64 `json:"total_duration"`
	LoadDuration       int64 `json:"load_duration"`
	PromptEvalCount    int   `json:"prompt_eval_count"`
	PromptEvalDuration int64 `json:"prompt_eval_duration"`
	EvalCount          int   `json:"eval_count"`
	EvalDuration       int64 `json:"eval_duration"`
}

// StreamResponse is updated to include the final stats.
type StreamResponse struct {
	Content string
	Done    bool
	Context json.RawMessage
	Error   string
	Stats   *GenerationStats `json:"stats,omitempty"` // NEW FIELD
}

// LLMProvider defines the interface for interacting with a language model.
type LLMProvider interface {
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	GenerateStream(ctx context.Context, req *GenerateRequest, ch chan<- StreamResponse) error
	ListModels(ctx context.Context) (*ListModelsResponse, error)
	PullModel(ctx context.Context, req *PullModelRequest, ch chan<- PullStatus) error
	DeleteModel(ctx context.Context, req *DeleteModelRequest) error
	ShowModelInfo(ctx context.Context, req *ShowModelRequest) (*ModelInfo, error)
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

// --- Chat Structs ---

// RequestOptions holds optional parameters for a generation request.
type RequestOptions struct {
	Temperature *float32 `json:"temperature,omitempty"`
	TopK        *int     `json:"top_k,omitempty"`
	TopP        *float32 `json:"top_p,omitempty"`
	System      *string  `json:"system,omitempty"`
	RepeatPenalty *float32 `json:"repeat_penalty,omitempty"`
	Seed        *int     `json:"seed,omitempty"`
}

type GenerateRequest struct {
	Model    string          `json:"model"`
	Prompt   string          `json:"prompt,omitempty"`
	Messages []Message       `json:"messages,omitempty"`
	Stream   bool            `json:"stream"`
	Context  json.RawMessage `json:"context,omitempty"`
	Options  *RequestOptions `json:"options,omitempty"`
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

// --- Model Management Structs ---
type ListModelsResponse struct {
	Models []Model `json:"models"`
}
type Model struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at"`
	Size       int64  `json:"size"`
}
type PullModelRequest struct {
	Name   string `json:"name"`
	Stream bool   `json:"stream"`
}
type PullStatus struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
	Error     string `json:"error,omitempty"`
}
type DeleteModelRequest struct {
	Name string `json:"name"`
}
type ShowModelRequest struct {
	Name string `json:"name"`
}
type ModelInfo struct {
	Modelfile  string `json:"modelfile"`
	Parameters string `json:"parameters"`
	Template   string `json:"template"`
}

// --- ollamaProvider methods ---

func (p *ollamaProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
    req.Stream = false
    body, err := json.Marshal(req)
    if err != nil { return nil, fmt.Errorf("could not marshal request: %w", err) }
    
    endpoint := p.url + "/api/chat"
    // Use /api/generate only if there's a single prompt and no messages.
    if len(req.Messages) == 0 && req.Prompt != "" {
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
    
    // --- ROBUST RESPONSE PARSING ---
    
    // Attempt to parse as a chat response first
    type ollamaChatResponse struct {
        Message Message `json:"message"`
        // Other fields we don't need here...
    }
    var chatResp ollamaChatResponse
    if err := json.Unmarshal(bodyBytes, &chatResp); err == nil && chatResp.Message.Content != "" {
        return &GenerateResponse{
            Response: chatResp.Message.Content,
        }, nil
    }

    // If chat parsing fails, attempt to parse as a generate response
    type ollamaGenerateResponse struct {
        Response string `json:"response"`
    }
    var genResp ollamaGenerateResponse
    if err := json.Unmarshal(bodyBytes, &genResp); err == nil {
        return &GenerateResponse{
            Response: genResp.Response,
        }, nil
    }

    return nil, fmt.Errorf("could not decode response from Ollama: %s", string(bodyBytes))
}

func (p *ollamaProvider) GenerateStream(ctx context.Context, req *GenerateRequest, ch chan<- StreamResponse) error {
	defer close(ch)
	req.Stream = true
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("could not marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.url+"/api/chat", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api returned non-200 status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// This struct helps decode both streaming content and the final stats block.
	type ollamaStreamChunk struct {
		Message            struct { Content string } `json:"message"`
		Model              string                   `json:"model"`
		Done               bool                     `json:"done"`
		Context            json.RawMessage          `json:"context"`
		TotalDuration      int64                    `json:"total_duration"`
		LoadDuration       int64                    `json:"load_duration"`
		PromptEvalCount    int                      `json:"prompt_eval_count"`
		PromptEvalDuration int64                    `json:"prompt_eval_duration"`
		EvalCount          int                      `json:"eval_count"`
		EvalDuration       int64                    `json:"eval_duration"`
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var chunk ollamaStreamChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			ch <- StreamResponse{Error: "Failed to decode stream chunk"}
			continue
		}

		streamResp := StreamResponse{
			Content: chunk.Message.Content,
			Done:    chunk.Done,
		}

		// If the stream is done, capture all the stats.
		if chunk.Done {
			streamResp.Context = chunk.Context
			streamResp.Stats = &GenerationStats{
				TotalDuration:      chunk.TotalDuration,
				LoadDuration:       chunk.LoadDuration,
				PromptEvalCount:    chunk.PromptEvalCount,
				PromptEvalDuration: chunk.PromptEvalDuration,
				EvalCount:          chunk.EvalCount,
				EvalDuration:       chunk.EvalDuration,
			}
		}

		select {
		case ch <- streamResp:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return scanner.Err()
}


func (p *ollamaProvider) ListModels(ctx context.Context) (*ListModelsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.url+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var listResp ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &listResp, nil
}

func (p *ollamaProvider) PullModel(ctx context.Context, req *PullModelRequest, ch chan<- PullStatus) error {
	defer close(ch)
	req.Stream = true
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("could not marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.url+"/api/pull", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var status PullStatus
		if err := json.Unmarshal(scanner.Bytes(), &status); err != nil {
			ch <- PullStatus{Error: "Failed to decode stream chunk"}
			continue
		}
		select {
		case ch <- status:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return scanner.Err()
}

func (p *ollamaProvider) DeleteModel(ctx context.Context, req *DeleteModelRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("could not marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", p.url+"/api/delete", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api returned non-200 status: %s", resp.Status)
	}
	return nil
}

func (p *ollamaProvider) ShowModelInfo(ctx context.Context, req *ShowModelRequest) (*ModelInfo, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.url+"/api/show", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var info ModelInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &info, nil
}