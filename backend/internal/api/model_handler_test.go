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

// setupModelHandler is a test helper that provides a ModelHandler instance
// with its ModelService dependency already mocked.
//
// WHY: This pattern keeps tests DRY and focused. Each test can start with a
// clean, pre-configured handler without repeating the setup code.
func setupModelHandler(t *testing.T) (*api.ModelHandler, *mocks.MockModelService) {
	mockModelSvc := mocks.NewMockModelService(t)
	handler := api.NewModelHandler(mockModelSvc)
	return handler, mockModelSvc
}

// TestModelHandler_HandleListModels tests the GET /v1/models endpoint.
//
// GOAL: Verify the handler correctly calls the service to list models and
// translates both successful results and errors into the correct HTTP responses.
func TestModelHandler_HandleListModels(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// ARRANGE: Set up the handler and configure the mock service to return a sample response.
		handler, mockSvc := setupModelHandler(t)
		expectedResp := &llm.ListModelsResponse{Models: []llm.Model{{Name: "test-model"}}}
		mockSvc.On("List", mock.Anything).Return(expectedResp, nil).Once()

		// ACT: Simulate an incoming HTTP request.
		req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		rr := httptest.NewRecorder()
		handler.HandleListModels(rr, req)

		// ASSERT: Check for a 200 OK status and validate that the JSON body matches the data from the service.
		assert.Equal(t, http.StatusOK, rr.Code)
		var resp llm.ListModelsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test-model", resp.Models[0].Name)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Failure", func(t *testing.T) {
		// ARRANGE: Configure the mock service to return an error.
		handler, mockSvc := setupModelHandler(t)
		mockSvc.On("List", mock.Anything).Return(nil, errors.New("internal error")).Once()

		// ACT
		req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		rr := httptest.NewRecorder()
		handler.HandleListModels(rr, req)

		// ASSERT: Verify the handler returns a 500 Internal Server Error.
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockSvc.AssertExpectations(t)
	})
}

// TestModelHandler_HandleDeleteModel tests the DELETE /v1/models endpoint.
//
// GOAL: Verify the handler correctly parses the request body and calls the
// appropriate service method.
func TestModelHandler_HandleDeleteModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// ARRANGE
		handler, mockSvc := setupModelHandler(t)
		reqBody := `{"name": "test-model"}`
		// We expect the `Delete` method to be called once with any context and any pointer to a DeleteModelRequest.
		mockSvc.On("Delete", mock.Anything, mock.AnythingOfType("*llm.DeleteModelRequest")).Return(nil).Once()

		// ACT
		req := httptest.NewRequest(http.MethodDelete, "/v1/models", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()
		handler.HandleDeleteModel(rr, req)

		// ASSERT
		assert.Equal(t, http.StatusOK, rr.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Failure - Invalid JSON", func(t *testing.T) {
		// GOAL: Ensure the handler rejects requests with malformed JSON bodies.
		handler, _ := setupModelHandler(t)
		reqBody := `{"name":`
		req := httptest.NewRequest(http.MethodDelete, "/v1/models", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()
		handler.HandleDeleteModel(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestModelHandler_HandleShowModel tests the POST /v1/models/show endpoint.
func TestModelHandler_HandleShowModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// ARRANGE
		handler, mockSvc := setupModelHandler(t)
		reqBody := `{"name": "test-model"}`
		expectedInfo := &llm.ModelInfo{Modelfile: "FROM scratch"}
		// Use `mock.MatchedBy` for precise argument validation.
		mockSvc.On("Show", mock.Anything, mock.MatchedBy(func(r *llm.ShowModelRequest) bool {
			return r.Name == "test-model"
		})).Return(expectedInfo, nil).Once()

		// ACT
		req := httptest.NewRequest(http.MethodPost, "/v1/models/show", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()
		handler.HandleShowModel(rr, req)

		// ASSERT
		assert.Equal(t, http.StatusOK, rr.Code)
		var resp llm.ModelInfo
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, expectedInfo, &resp)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Failure - Service returns not found", func(t *testing.T) {
		// ARRANGE: Simulate the service returning a specific business error.
		handler, mockSvc := setupModelHandler(t)
		reqBody := `{"name": "test-model"}`
		mockSvc.On("Show", mock.Anything, mock.Anything).Return(nil, app_errors.ErrNotFound).Once()

		// ACT
		req := httptest.NewRequest(http.MethodPost, "/v1/models/show", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()
		handler.HandleShowModel(rr, req)

		// ASSERT: Verify the handler correctly maps the domain error to a 404 Not Found status.
		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockSvc.AssertExpectations(t)
	})
}

// TestModelHandler_HandlePullModel tests the streaming POST /v1/models/pull endpoint.
func TestModelHandler_HandlePullModel(t *testing.T) {
	t.Run("Success - Service is called", func(t *testing.T) {
		handler, mockSvc := setupModelHandler(t)
		reqBody := `{"name": "test-model"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/models/pull", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		// Similar to the chat stream test, we must simulate the service closing
		// the channel to prevent the handler from blocking forever and causing a timeout.
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

		// For streaming endpoints, errors are sent in the response body as an event.
		assert.Contains(t, rr.Body.String(), "Invalid request body")
	})
}
