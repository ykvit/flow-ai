package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"flow-ai/backend/internal/api"
	app_errors "flow-ai/backend/internal/errors"
	"flow-ai/backend/internal/interfaces/mocks"
	"flow-ai/backend/internal/llm"
)

func setupModelHandler(t *testing.T) (*api.ModelHandler, *mocks.MockModelService) {
	mockModelSvc := mocks.NewMockModelService(t)
	handler := api.NewModelHandler(mockModelSvc)
	return handler, mockModelSvc
}

func TestModelHandler_HandleListModels(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		handler, mockSvc := setupModelHandler(t)
		expectedResp := &llm.ListModelsResponse{Models: []llm.Model{{Name: "test-model"}}}
		mockSvc.On("List", mock.Anything).Return(expectedResp, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		rr := httptest.NewRecorder()

		handler.HandleListModels(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp llm.ListModelsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test-model", resp.Models[0].Name)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Failure", func(t *testing.T) {
		handler, mockSvc := setupModelHandler(t)
		mockSvc.On("List", mock.Anything).Return(nil, errors.New("internal error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		rr := httptest.NewRecorder()

		handler.HandleListModels(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestModelHandler_HandleDeleteModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		handler, mockSvc := setupModelHandler(t)
		reqBody := `{"name": "test-model"}`
		mockSvc.On("Delete", mock.Anything, mock.AnythingOfType("*llm.DeleteModelRequest")).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/v1/models", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		handler.HandleDeleteModel(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Failure - Invalid JSON", func(t *testing.T) {
		handler, _ := setupModelHandler(t)
		reqBody := `{"name":`
		req := httptest.NewRequest(http.MethodDelete, "/v1/models", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		handler.HandleDeleteModel(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestModelHandler_HandleShowModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		handler, mockSvc := setupModelHandler(t)
		reqBody := `{"name": "test-model"}`
		expectedInfo := &llm.ModelInfo{Modelfile: "FROM scratch"}

		mockSvc.On("Show", mock.Anything, mock.MatchedBy(func(r *llm.ShowModelRequest) bool {
			return r.Name == "test-model"
		})).Return(expectedInfo, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/v1/models/show", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()
		handler.HandleShowModel(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp llm.ModelInfo
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, expectedInfo, &resp)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Failure - Service returns not found", func(t *testing.T) {
		handler, mockSvc := setupModelHandler(t)
		reqBody := `{"name": "test-model"}`

		mockSvc.On("Show", mock.Anything, mock.Anything).Return(nil, app_errors.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodPost, "/v1/models/show", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()
		handler.HandleShowModel(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestModelHandler_HandlePullModel(t *testing.T) {
	t.Run("Success - Service is called", func(t *testing.T) {
		handler, mockSvc := setupModelHandler(t)
		reqBody := `{"name": "test-model"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/models/pull", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		mockSvc.On("Pull", mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				streamChan := args.Get(2).(chan<- llm.PullStatus)
				close(streamChan)
			}).Return(nil).Once()

		handler.HandlePullModel(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Failure - Invalid JSON", func(t *testing.T) {
		handler, _ := setupModelHandler(t)
		reqBody := `{"name":`
		req := httptest.NewRequest(http.MethodPost, "/v1/models/pull", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		handler.HandlePullModel(rr, req)

		assert.Contains(t, rr.Body.String(), "Invalid request body")
	})
}
