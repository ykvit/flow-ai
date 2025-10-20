// backend/internal/api/handler_test.go
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
	"flow-ai/backend/internal/interfaces/mocks"
	"flow-ai/backend/internal/model"
	"flow-ai/backend/internal/service"
)

func setupChatHandler(t *testing.T) (*api.ChatHandler, *mocks.MockChatService, *mocks.MockSettingsService) {
	mockChatSvc := mocks.NewMockChatService(t)
	mockSettingsSvc := mocks.NewMockSettingsService(t)
	handler := api.NewChatHandler(mockChatSvc, mockSettingsSvc)
	return handler, mockChatSvc, mockSettingsSvc
}

func addChiURLParams(req *http.Request, params map[string]string) *http.Request {
	chiCtx := chi.NewRouteContext()
	for key, value := range params {
		chiCtx.URLParams.Add(key, value)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
}

func TestChatHandler_GetSettings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		handler, _, mockSettingsSvc := setupChatHandler(t)
		expectedSettings := &service.Settings{MainModel: "test"}
		mockSettingsSvc.On("Get", mock.Anything).Return(expectedSettings, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/settings", nil)
		rr := httptest.NewRecorder()

		handler.GetSettings(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockSettingsSvc.AssertExpectations(t)
	})

	t.Run("Failure", func(t *testing.T) {
		handler, _, mockSettingsSvc := setupChatHandler(t)
		mockSettingsSvc.On("Get", mock.Anything).Return(nil, app_errors.ErrInternal).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/settings", nil)
		rr := httptest.NewRecorder()

		handler.GetSettings(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockSettingsSvc.AssertExpectations(t)
	})
}

func TestChatHandler_GetChats(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		handler, mockChatSvc, _ := setupChatHandler(t)

		expectedChats := []*model.Chat{{ID: "chat1", Title: "Test Chat"}}
		mockChatSvc.On("ListChats", mock.Anything).Return(expectedChats, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/chats", nil)
		rr := httptest.NewRecorder()

		handler.GetChats(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var returnedChats []*model.Chat
		err := json.Unmarshal(rr.Body.Bytes(), &returnedChats)
		assert.NoError(t, err)
		assert.Equal(t, expectedChats, returnedChats)

		mockChatSvc.AssertExpectations(t)
	})

	t.Run("Failure - Service returns error", func(t *testing.T) {
		handler, mockChatSvc, _ := setupChatHandler(t)

		mockChatSvc.On("ListChats", mock.Anything).Return(nil, errors.New("internal error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/chats", nil)
		rr := httptest.NewRecorder()

		handler.GetChats(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "internal server error")
	})
}

func TestChatHandler_GetChat(t *testing.T) {
	chatID := "test-chat-id"

	t.Run("Success", func(t *testing.T) {
		handler, mockChatSvc, _ := setupChatHandler(t)
		expectedChat := &model.FullChat{Chat: model.Chat{ID: chatID}}
		mockChatSvc.On("GetFullChat", mock.Anything, chatID).Return(expectedChat, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/chats/"+chatID, nil)
		req = addChiURLParams(req, map[string]string{"chatID": chatID})

		rr := httptest.NewRecorder()
		handler.GetChat(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockChatSvc.AssertExpectations(t)
	})

	t.Run("Failure - Not Found", func(t *testing.T) {
		handler, mockChatSvc, _ := setupChatHandler(t)
		mockChatSvc.On("GetFullChat", mock.Anything, chatID).Return(nil, app_errors.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/chats/"+chatID, nil)
		req = addChiURLParams(req, map[string]string{"chatID": chatID})

		rr := httptest.NewRecorder()
		handler.GetChat(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockChatSvc.AssertExpectations(t)
	})
}

func TestChatHandler_UpdateSettings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		handler, _, mockSettingsSvc := setupChatHandler(t)

		settingsJSON := `{"system_prompt":"new prompt","main_model":"model1"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/settings", strings.NewReader(settingsJSON))
		rr := httptest.NewRecorder()

		mockSettingsSvc.On("Save", mock.Anything, mock.MatchedBy(func(s *service.Settings) bool {
			return s.MainModel == "model1" && s.SystemPrompt == "new prompt"
		})).Return(nil).Once()

		handler.UpdateSettings(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockSettingsSvc.AssertExpectations(t)
	})

	t.Run("Failure - Invalid JSON", func(t *testing.T) {
		handler, _, _ := setupChatHandler(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/settings", strings.NewReader(`{invalid`))
		rr := httptest.NewRecorder()
		handler.UpdateSettings(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Failure - Validation Error", func(t *testing.T) {
		handler, _, _ := setupChatHandler(t)
		settingsJSON := `{"system_prompt":"new prompt","main_model":""}`
		req := httptest.NewRequest(http.MethodPost, "/v1/settings", strings.NewReader(settingsJSON))
		rr := httptest.NewRecorder()

		handler.UpdateSettings(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Field 'MainModel' failed on the 'required' tag")
	})
}

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

func TestChatHandler_HandleStreamMessage(t *testing.T) {
	t.Run("Success - Service is called", func(t *testing.T) {
		handler, mockChatSvc, _ := setupChatHandler(t)
		reqBody := `{"content": "hello"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/chats/messages", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

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
