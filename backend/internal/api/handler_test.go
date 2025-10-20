// The `_test` suffix creates a "black box" test package.
// This means the test code lives outside the `api` package and can only access
// its exported identifiers (functions, types, etc.). This is the preferred
// approach for testing the public API of a package.
package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"flow-ai/backend/internal/api"
	app_errors "flow-ai/backend/internal/errors"

	// We import the generated mocks for our service interfaces.
	"flow-ai/backend/internal/interfaces/mocks"
	"flow-ai/backend/internal/model"
	"flow-ai/backend/internal/service"
)

// setupChatHandler is a test helper function (or "test fixture").
//
// WHY: It encapsulates the repetitive setup logic for creating a handler with
// its dependencies mocked. This keeps our test cases clean, readable, and focused
// on the specific behavior being tested, adhering to the DRY principle.
func setupChatHandler(t *testing.T) (*api.ChatHandler, *mocks.MockChatService, *mocks.MockSettingsService) {
	mockChatSvc := mocks.NewMockChatService(t)
	mockSettingsSvc := mocks.NewMockSettingsService(t)
	handler := api.NewChatHandler(mockChatSvc, mockSettingsSvc)
	return handler, mockChatSvc, mockSettingsSvc
}

// addChiURLParams is a helper to simulate how the chi router injects URL
// parameters (e.g., `{chatID}`) into the request's context.
//
// WHY: Our handlers rely on `chi.URLParam` to extract these values. Without this
// helper, `chi.URLParam` would return an empty string, and our tests for routes
// like `/chats/{chatID}` would fail.
func addChiURLParams(req *http.Request, params map[string]string) *http.Request {
	chiCtx := chi.NewRouteContext()
	for key, value := range params {
		chiCtx.URLParams.Add(key, value)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
}

// TestChatHandler_GetSettings tests the GET /v1/settings endpoint.
//
// GOAL: Verify the handler correctly calls the SettingsService and translates
// its responses (both success and error) into the appropriate HTTP status codes.
func TestChatHandler_GetSettings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// ARRANGE: Set up the handler and define the mock's behavior.
		handler, _, mockSettingsSvc := setupChatHandler(t)
		expectedSettings := &service.Settings{MainModel: "test"}
		mockSettingsSvc.On("Get", mock.Anything).Return(expectedSettings, nil).Once()

		// ACT: Create a simulated HTTP request and record the response.
		req := httptest.NewRequest(http.MethodGet, "/v1/settings", nil)
		rr := httptest.NewRecorder()
		handler.GetSettings(rr, req)

		// ASSERT: Check the HTTP status code and verify the mock was called as expected.
		assert.Equal(t, http.StatusOK, rr.Code)
		mockSettingsSvc.AssertExpectations(t)
	})

	t.Run("Failure", func(t *testing.T) {
		// ARRANGE: Configure the mock to return a generic internal error.
		handler, _, mockSettingsSvc := setupChatHandler(t)
		mockSettingsSvc.On("Get", mock.Anything).Return(nil, app_errors.ErrInternal).Once()

		// ACT
		req := httptest.NewRequest(http.MethodGet, "/v1/settings", nil)
		rr := httptest.NewRecorder()
		handler.GetSettings(rr, req)

		// ASSERT: Verify the handler correctly maps the internal error to a 500 status.
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockSettingsSvc.AssertExpectations(t)
	})
}

// TestChatHandler_GetChats tests the GET /v1/chats endpoint.
func TestChatHandler_GetChats(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// ARRANGE
		handler, mockChatSvc, _ := setupChatHandler(t)
		expectedChats := []*model.Chat{{ID: "chat1", Title: "Test Chat"}}
		mockChatSvc.On("ListChats", mock.Anything).Return(expectedChats, nil).Once()

		// ACT
		req := httptest.NewRequest(http.MethodGet, "/v1/chats", nil)
		rr := httptest.NewRecorder()
		handler.GetChats(rr, req)

		// ASSERT
		assert.Equal(t, http.StatusOK, rr.Code)
		// Also assert that the JSON body of the response matches what the service returned.
		var returnedChats []*model.Chat
		err := json.Unmarshal(rr.Body.Bytes(), &returnedChats)
		assert.NoError(t, err)
		assert.Equal(t, expectedChats, returnedChats)

		mockChatSvc.AssertExpectations(t)
	})

	t.Run("Failure - Service returns error", func(t *testing.T) {
		// ARRANGE
		handler, mockChatSvc, _ := setupChatHandler(t)
		mockChatSvc.On("ListChats", mock.Anything).Return(nil, errors.New("internal error")).Once()

		// ACT
		req := httptest.NewRequest(http.MethodGet, "/v1/chats", nil)
		rr := httptest.NewRecorder()
		handler.GetChats(rr, req)

		// ASSERT
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "internal server error")
	})
}

// TestChatHandler_GetChat tests the GET /v1/chats/{chatID} endpoint.
func TestChatHandler_GetChat(t *testing.T) {
	chatID := "test-chat-id"

	t.Run("Success", func(t *testing.T) {
		// ARRANGE
		handler, mockChatSvc, _ := setupChatHandler(t)
		expectedChat := &model.FullChat{Chat: model.Chat{ID: chatID}}
		mockChatSvc.On("GetFullChat", mock.Anything, chatID).Return(expectedChat, nil).Once()

		// ACT
		req := httptest.NewRequest(http.MethodGet, "/v1/chats/"+chatID, nil)
		req = addChiURLParams(req, map[string]string{"chatID": chatID})
		rr := httptest.NewRecorder()
		handler.GetChat(rr, req)

		// ASSERT
		assert.Equal(t, http.StatusOK, rr.Code)
		mockChatSvc.AssertExpectations(t)
	})

	t.Run("Failure - Not Found", func(t *testing.T) {
		// ARRANGE: Simulate the service returning a specific sentinel error.
		handler, mockChatSvc, _ := setupChatHandler(t)
		mockChatSvc.On("GetFullChat", mock.Anything, chatID).Return(nil, app_errors.ErrNotFound).Once()

		// ACT
		req := httptest.NewRequest(http.MethodGet, "/v1/chats/"+chatID, nil)
		req = addChiURLParams(req, map[string]string{"chatID": chatID})
		rr := httptest.NewRecorder()
		handler.GetChat(rr, req)

		// ASSERT: Verify the handler correctly maps `ErrNotFound` to a 404 status.
		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockChatSvc.AssertExpectations(t)
	})
}

// TestChatHandler_UpdateSettings tests the POST /v1/settings endpoint.
// GOAL: Verify JSON parsing, validation logic, and service invocation.
func TestChatHandler_UpdateSettings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		handler, _, mockSettingsSvc := setupChatHandler(t)
		settingsJSON := `{"system_prompt":"new prompt","main_model":"model1"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/settings", strings.NewReader(settingsJSON))
		rr := httptest.NewRecorder()

		// `mock.MatchedBy` allows for more complex argument matching. Here, we
		// verify that the `Save` method is called with a struct that has the correct field values.
		mockSettingsSvc.On("Save", mock.Anything, mock.MatchedBy(func(s *service.Settings) bool {
			return s.MainModel == "model1" && s.SystemPrompt == "new prompt"
		})).Return(nil).Once()

		handler.UpdateSettings(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockSettingsSvc.AssertExpectations(t)
	})

	t.Run("Failure - Invalid JSON", func(t *testing.T) {
		// GOAL: Ensure malformed JSON bodies are rejected with a 400 Bad Request.
		handler, _, _ := setupChatHandler(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/settings", strings.NewReader(`{invalid`))
		rr := httptest.NewRecorder()
		handler.UpdateSettings(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Failure - Validation Error", func(t *testing.T) {
		// GOAL: Ensure that valid JSON that fails business rules (defined by
		// `validate` tags) is rejected with a 400 Bad Request.
		handler, _, _ := setupChatHandler(t)
		settingsJSON := `{"system_prompt":"new prompt","main_model":""}`
		req := httptest.NewRequest(http.MethodPost, "/v1/settings", strings.NewReader(settingsJSON))
		rr := httptest.NewRecorder()

		handler.UpdateSettings(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Field 'MainModel' failed on the 'required' tag")
	})
}

// TestChatHandler_UpdateChatTitle tests the PUT /v1/chats/{chatID}/title endpoint.
func TestChatHandler_UpdateChatTitle(t *testing.T) {
	chatID := "test-chat-id"

	t.Run("Success", func(t *testing.T) {
		handler, mockChatSvc, _ := setupChatHandler(t)
		newTitle := "A valid title"
		reqBody := `{"title": "` + newTitle + `"}`
		mockChatSvc.On("UpdateChatTitle", mock.Anything, chatID, newTitle).Return(nil).Once()
		req := httptest.NewRequest(http.MethodPut, "/v1/chats/"+chatID+"/title", strings.NewReader(reqBody))
		req = addChiURLParams(req, map[string]string{"chatID": chatID})
		rr := httptest.NewRecorder()
		handler.UpdateChatTitle(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		mockChatSvc.AssertExpectations(t)
	})

	t.Run("Failure - Validation Error (empty title)", func(t *testing.T) {
		handler, _, _ := setupChatHandler(t)
		reqBody := `{"title": ""}`
		req := httptest.NewRequest(http.MethodPut, "/v1/chats/"+chatID+"/title", strings.NewReader(reqBody))
		req = addChiURLParams(req, map[string]string{"chatID": chatID})
		rr := httptest.NewRecorder()
		handler.UpdateChatTitle(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Field 'Title' failed on the 'required' tag")
	})

	t.Run("Failure - Bad JSON", func(t *testing.T) {
		handler, _, _ := setupChatHandler(t)
		reqBody := `{"title":`
		req := httptest.NewRequest(http.MethodPut, "/v1/chats/"+chatID+"/title", strings.NewReader(reqBody))
		req = addChiURLParams(req, map[string]string{"chatID": chatID})
		rr := httptest.NewRecorder()
		handler.UpdateChatTitle(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestChatHandler_HandleDeleteChat tests the DELETE /v1/chats/{chatID} endpoint.
func TestChatHandler_HandleDeleteChat(t *testing.T) {
	chatID := "test-chat-id"

	t.Run("Success", func(t *testing.T) {
		handler, mockChatSvc, _ := setupChatHandler(t)
		mockChatSvc.On("DeleteChat", mock.Anything, chatID).Return(nil).Once()
		req := httptest.NewRequest(http.MethodDelete, "/v1/chats/"+chatID, nil)
		req = addChiURLParams(req, map[string]string{"chatID": chatID})
		rr := httptest.NewRecorder()
		handler.HandleDeleteChat(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		mockChatSvc.AssertExpectations(t)
	})

	t.Run("Failure - Not Found", func(t *testing.T) {
		handler, mockChatSvc, _ := setupChatHandler(t)
		mockChatSvc.On("DeleteChat", mock.Anything, chatID).Return(app_errors.ErrNotFound).Once()
		req := httptest.NewRequest(http.MethodDelete, "/v1/chats/"+chatID, nil)
		req = addChiURLParams(req, map[string]string{"chatID": chatID})
		rr := httptest.NewRecorder()
		handler.HandleDeleteChat(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockChatSvc.AssertExpectations(t)
	})
}

// TestChatHandler_HandleStreamMessage tests the streaming POST /v1/chats/messages endpoint.
//
// GOAL: Verify that the handler correctly sets up the stream, validates the
// input, and calls the service. We are NOT testing the streaming logic itself,
// only the handler's responsibilities.
func TestChatHandler_HandleStreamMessage(t *testing.T) {
	t.Run("Success - Service is called", func(t *testing.T) {
		handler, mockChatSvc, _ := setupChatHandler(t)
		reqBody := `{"content": "hello"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/chats/messages", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		// The service method `HandleNewMessage` runs in a goroutine. To prevent
		// the test from blocking forever, we must simulate the service closing
		// the channel when it's done. The `.Run` function is perfect for this.
		mockChatSvc.On("HandleNewMessage", mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				streamChan := args.Get(2).(chan<- model.StreamResponse)
				close(streamChan)
			}).Once()

		handler.HandleStreamMessage(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/event-stream", rr.Header().Get("Content-Type"))
		mockChatSvc.AssertExpectations(t)
	})

	t.Run("Failure - Invalid JSON", func(t *testing.T) {
		handler, _, _ := setupChatHandler(t)
		reqBody := `{"content":`
		req := httptest.NewRequest(http.MethodPost, "/v1/chats/messages", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		handler.HandleStreamMessage(rr, req)

		// For streaming endpoints, errors are sent over the stream itself.
		// We assert that the response body contains the error event.
		assert.Contains(t, rr.Body.String(), "Invalid request body")
	})

	t.Run("Failure - Validation Error", func(t *testing.T) {
		handler, _, _ := setupChatHandler(t)
		reqBody := `{"content": ""}`
		req := httptest.NewRequest(http.MethodPost, "/v1/chats/messages", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		handler.HandleStreamMessage(rr, req)

		assert.Contains(t, rr.Body.String(), "Field 'Content' failed on the 'required' tag")
	})
}
