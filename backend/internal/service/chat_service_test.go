package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"flow-ai/backend/internal/llm"
	mock_llm "flow-ai/backend/internal/llm/mocks"
	"flow-ai/backend/internal/model"
	"flow-ai/backend/internal/repository"
	mock_repo "flow-ai/backend/internal/repository/mocks"
	"flow-ai/backend/internal/service"
)

type Mocks struct {
	repo   *mock_repo.MockRepository
	llm    *mock_llm.MockLLMProvider
	db     *sql.DB
	mockDB sqlmock.Sqlmock
}

func setupChatService(t *testing.T) (*service.ChatService, Mocks) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)

	mocks := Mocks{
		repo:   mock_repo.NewMockRepository(t),
		llm:    mock_llm.NewMockLLMProvider(t),
		db:     db,
		mockDB: mockDB,
	}

	settingsService := service.NewSettingsService(mocks.db, mocks.llm)
	chatService := service.NewChatService(mocks.repo, mocks.llm, settingsService)

	return chatService, mocks
}

func TestChatService_UpdateChatTitle(t *testing.T) {
	ctx := context.Background()
	chatID := "chat123"
	newTitle := "New Title"

	t.Run("Success", func(t *testing.T) {
		chatService, mocks := setupChatService(t)
		defer func() {
			_ = mocks.db.Close()
		}()

		mocks.repo.On("UpdateChatTitle", ctx, chatID, newTitle).Return(nil).Once()

		err := chatService.UpdateChatTitle(ctx, chatID, newTitle)
		assert.NoError(t, err)
	})

	t.Run("Failure - Repository returns not found", func(t *testing.T) {
		chatService, mocks := setupChatService(t)
		defer func() {
			_ = mocks.db.Close()
		}()
		mocks.repo.On("UpdateChatTitle", ctx, chatID, newTitle).Return(repository.ErrNotFound).Once()

		err := chatService.UpdateChatTitle(ctx, chatID, newTitle)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "not found")
	})
}

func TestChatService_ListChats(t *testing.T) {
	ctx := context.Background()
	chatService, mocks := setupChatService(t)
	defer func() {
		_ = mocks.db.Close()
	}()

	expectedChats := []*model.Chat{{ID: "chat1"}}
	mocks.repo.On("GetChats", ctx).Return(expectedChats, nil).Once()

	chats, err := chatService.ListChats(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedChats, chats)
}

func TestChatService_GetFullChat(t *testing.T) {
	ctx := context.Background()
	chatID := "chat123"

	t.Run("Success", func(t *testing.T) {
		chatService, mocks := setupChatService(t)
		defer func() {
			_ = mocks.db.Close()
		}()

		chat := &model.Chat{ID: chatID}
		messages := []model.Message{{ID: "msg1"}}

		mocks.repo.On("GetChat", ctx, chatID).Return(chat, nil).Once()
		mocks.repo.On("GetActiveMessagesByChatID", ctx, chatID).Return(messages, nil).Once()

		fullChat, err := chatService.GetFullChat(ctx, chatID)
		require.NoError(t, err)
		assert.Equal(t, chat, &fullChat.Chat)
		assert.Equal(t, messages, fullChat.Messages)
	})

	t.Run("Failure - GetChat returns error", func(t *testing.T) {
		chatService, mocks := setupChatService(t)
		defer func() {
			_ = mocks.db.Close()
		}()

		mocks.repo.On("GetChat", ctx, chatID).Return(nil, errors.New("db error")).Once()

		_, err := chatService.GetFullChat(ctx, chatID)
		assert.Error(t, err)
	})

	t.Run("Failure - GetActiveMessagesByChatID returns error", func(t *testing.T) {
		chatService, mocks := setupChatService(t)
		defer func() {
			_ = mocks.db.Close()
		}()

		mocks.repo.On("GetChat", ctx, chatID).Return(&model.Chat{}, nil).Once()
		mocks.repo.On("GetActiveMessagesByChatID", ctx, chatID).Return(nil, errors.New("db error")).Once()

		_, err := chatService.GetFullChat(ctx, chatID)
		assert.Error(t, err)
	})
}

func TestChatService_HandleNewMessage_NewChat(t *testing.T) {
	ctx := context.Background()
	req := &service.CreateMessageRequest{Content: "Hello"}

	t.Run("Success - Happy Path", func(t *testing.T) {
		chatService, mocks := setupChatService(t)
		defer func() {
			_ = mocks.db.Close()
		}()

		streamChan := make(chan model.StreamResponse, 5)

		rows := sqlmock.NewRows([]string{"key", "value"}).
			AddRow("system_prompt", "system").
			AddRow("main_model", "test-model").
			AddRow("support_model", "support-model")
		mocks.mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnRows(rows)

		mocks.repo.On("CreateChat", ctx, mock.AnythingOfType("*model.Chat")).Return(nil).Once()
		mocks.repo.On("GetLastActiveMessage", ctx, mock.AnythingOfType("string")).Return(nil, repository.ErrNotFound).Once()
		mocks.repo.On("AddMessage", ctx, mock.AnythingOfType("*model.Message"), mock.AnythingOfType("string")).Return(nil).Twice()
		mocks.repo.On("GetActiveMessagesByChatID", ctx, mock.AnythingOfType("string")).Return([]model.Message{}, nil).Once()
		mocks.repo.On("UpdateMessageContext", ctx, mock.Anything, mock.Anything).Return(nil).Once()
		mocks.repo.On("UpdateChatTitle", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Maybe()

		mocks.llm.On("Generate", mock.Anything, mock.Anything).Return(&llm.GenerateResponse{Response: `{"title": "Test"}`}, nil).Maybe()
		mocks.llm.On("GenerateStream", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).
			Run(func(args mock.Arguments) {
				outChan := args.Get(2).(chan<- llm.StreamResponse)
				outChan <- llm.StreamResponse{Content: "response"}
				outChan <- llm.StreamResponse{Done: true, Context: []byte(`"context"`)}
				close(outChan)
			}).Once()

		chatService.HandleNewMessage(ctx, req, streamChan)

		assert.Len(t, streamChan, 2)
		<-streamChan
		finalChunk := <-streamChan
		assert.True(t, finalChunk.Done)

		require.NoError(t, mocks.mockDB.ExpectationsWereMet())
		mocks.repo.AssertExpectations(t)
		mocks.llm.AssertExpectations(t)
	})

	t.Run("Failure - Settings service returns error", func(t *testing.T) {
		chatService, mocks := setupChatService(t)
		defer func() {
			_ = mocks.db.Close()
		}()

		streamChan := make(chan model.StreamResponse, 1)

		mocks.mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnError(errors.New("db error"))

		chatService.HandleNewMessage(ctx, req, streamChan)

		errChunk := <-streamChan
		assert.NotEmpty(t, errChunk.Error)
		assert.Contains(t, errChunk.Error, "Could not load application settings")
		require.NoError(t, mocks.mockDB.ExpectationsWereMet())
	})

	t.Run("Failure - CreateChat returns error", func(t *testing.T) {
		chatService, mocks := setupChatService(t)
		defer func() {
			_ = mocks.db.Close()
		}()

		streamChan := make(chan model.StreamResponse, 1)

		rows := sqlmock.NewRows([]string{"key", "value"}).
			AddRow("system_prompt", "system").
			AddRow("main_model", "test-model")
		mocks.mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnRows(rows)

		mocks.repo.On("CreateChat", ctx, mock.Anything).Return(errors.New("db error")).Once()

		chatService.HandleNewMessage(ctx, req, streamChan)

		errChunk := <-streamChan
		assert.NotEmpty(t, errChunk.Error)
		assert.Contains(t, errChunk.Error, "Could not create chat")
		require.NoError(t, mocks.mockDB.ExpectationsWereMet())
	})
}
