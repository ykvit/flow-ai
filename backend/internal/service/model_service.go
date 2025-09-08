package service

import (
	"context"
	"flow-ai/backend/internal/llm"
)

// ModelService handles the business logic for model management.
type ModelService struct {
	llm llm.LLMProvider
}

// NewModelService creates a new ModelService.
func NewModelService(llmProvider llm.LLMProvider) *ModelService {
	return &ModelService{llm: llmProvider}
}

// List returns a list of all locally available models.
func (s *ModelService) List(ctx context.Context) (*llm.ListModelsResponse, error) {
	return s.llm.ListModels(ctx)
}

// Pull downloads a model from a registry. It streams the progress.
func (s *ModelService) Pull(ctx context.Context, req *llm.PullModelRequest, ch chan<- llm.PullStatus) error {
	return s.llm.PullModel(ctx, req, ch)
}

// Delete removes a local model.
func (s *ModelService) Delete(ctx context.Context, req *llm.DeleteModelRequest) error {
	return s.llm.DeleteModel(ctx, req)
}

// Show retrieves detailed information about a model.
func (s *ModelService) Show(ctx context.Context, req *llm.ShowModelRequest) (*llm.ModelInfo, error) {
	return s.llm.ShowModelInfo(ctx, req)
}