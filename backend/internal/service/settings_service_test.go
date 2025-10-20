package service_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/llm/mocks"
	"flow-ai/backend/internal/service"
)

func setupSettingsService(t *testing.T) (*service.SettingsService, *sql.DB, sqlmock.Sqlmock, *mocks.MockLLMProvider) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)

	mockLLM := mocks.NewMockLLMProvider(t)
	settingsService := service.NewSettingsService(db, mockLLM)

	return settingsService, db, mockDB, mockLLM
}

func TestSettingsService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Get existing settings", func(t *testing.T) {
		settingsService, db, mockDB, _ := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		rows := sqlmock.NewRows([]string{"key", "value"}).
			AddRow("system_prompt", "test prompt").
			AddRow("main_model", "test-model").
			AddRow("support_model", "support-model")

		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnRows(rows)

		settings, err := settingsService.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, "test prompt", settings.SystemPrompt)
		assert.Equal(t, "test-model", settings.MainModel)
		assert.Equal(t, "support-model", settings.SupportModel)

		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Success - Self-heal when main model is empty", func(t *testing.T) {
		settingsService, db, mockDB, mockLLM := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		rows := sqlmock.NewRows([]string{"key", "value"}).
			AddRow("system_prompt", "test prompt").
			AddRow("main_model", "").
			AddRow("support_model", "")

		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnRows(rows)

		mockLLM.On("ListModels", ctx).Return(&llm.ListModelsResponse{
			Models: []llm.Model{{Name: "discovered-model", ModifiedAt: time.Now().String()}},
		}, nil).Once()

		mockDB.ExpectBegin()
		prep := mockDB.ExpectPrepare("INSERT INTO settings")
		prep.ExpectExec().WithArgs("main_model", "discovered-model").WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().WithArgs("support_model", "discovered-model").WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().WithArgs("system_prompt", "test prompt").WillReturnResult(sqlmock.NewResult(1, 1))
		mockDB.ExpectCommit()

		settings, err := settingsService.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, "discovered-model", settings.MainModel)
		assert.Equal(t, "discovered-model", settings.SupportModel)

		assert.NoError(t, mockDB.ExpectationsWereMet())
		mockLLM.AssertExpectations(t)
	})

	t.Run("Failure - DB error on get", func(t *testing.T) {
		settingsService, db, mockDB, _ := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		expectedErr := errors.New("db error")
		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnError(expectedErr)

		settings, err := settingsService.Get(ctx)
		require.Error(t, err)
		assert.Nil(t, settings)
		assert.Contains(t, err.Error(), expectedErr.Error())
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})
}

func TestSettingsService_InitAndGet(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Settings already exist, just get them", func(t *testing.T) {
		settingsService, db, mockDB, _ := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		rows := sqlmock.NewRows([]string{"key", "value"}).
			AddRow("main_model", "existing-model")
		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnRows(rows)
		rows2 := sqlmock.NewRows([]string{"key", "value"}).
			AddRow("main_model", "existing-model")
		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnRows(rows2)

		settings, err := settingsService.InitAndGet(ctx, "default")
		require.NoError(t, err)
		assert.Equal(t, "existing-model", settings.MainModel)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Success - No settings, initialize with discovered model", func(t *testing.T) {
		settingsService, db, mockDB, mockLLM := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnError(sql.ErrNoRows)
		mockLLM.On("ListModels", ctx).Return(&llm.ListModelsResponse{
			Models: []llm.Model{{Name: "discovered-model"}},
		}, nil).Once()

		mockDB.ExpectBegin()
		prep := mockDB.ExpectPrepare("INSERT INTO settings")
		prep.ExpectExec().WithArgs("main_model", "discovered-model").WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().WithArgs("support_model", "discovered-model").WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().WithArgs("system_prompt", "default prompt").WillReturnResult(sqlmock.NewResult(1, 1))
		mockDB.ExpectCommit()

		settings, err := settingsService.InitAndGet(ctx, "default prompt")
		require.NoError(t, err)
		assert.Equal(t, "discovered-model", settings.MainModel)
		assert.NoError(t, mockDB.ExpectationsWereMet())
		mockLLM.AssertExpectations(t)
	})

	t.Run("Success - No settings, no models available", func(t *testing.T) {
		settingsService, db, mockDB, mockLLM := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnError(sql.ErrNoRows)
		mockLLM.On("ListModels", ctx).Return(&llm.ListModelsResponse{Models: []llm.Model{}}, nil).Once()

		mockDB.ExpectBegin()
		prep := mockDB.ExpectPrepare("INSERT INTO settings")
		prep.ExpectExec().WithArgs("main_model", "").WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().WithArgs("support_model", "").WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().WithArgs("system_prompt", "default").WillReturnResult(sqlmock.NewResult(1, 1))
		mockDB.ExpectCommit()

		settings, err := settingsService.InitAndGet(ctx, "default")
		require.NoError(t, err)
		assert.Equal(t, "", settings.MainModel)
		assert.NoError(t, mockDB.ExpectationsWereMet())
		mockLLM.AssertExpectations(t)
	})
}

func TestSettingsService_Save(t *testing.T) {
	ctx := context.Background()
	settingsToSave := &service.Settings{
		SystemPrompt: "new prompt",
		MainModel:    "model1",
		SupportModel: "model2",
	}

	t.Run("Success - Save valid settings", func(t *testing.T) {
		settingsService, db, mockDB, mockLLM := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		mockLLM.On("ListModels", ctx).Return(&llm.ListModelsResponse{
			Models: []llm.Model{{Name: "model1"}, {Name: "model2"}},
		}, nil).Once()

		mockDB.ExpectBegin()
		prep := mockDB.ExpectPrepare(regexp.QuoteMeta("INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value"))
		prep.ExpectExec().WithArgs("main_model", "model1").WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().WithArgs("support_model", "model2").WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().WithArgs("system_prompt", "new prompt").WillReturnResult(sqlmock.NewResult(1, 1))
		mockDB.ExpectCommit()

		err := settingsService.Save(ctx, settingsToSave)
		require.NoError(t, err)
		assert.NoError(t, mockDB.ExpectationsWereMet())
		mockLLM.AssertExpectations(t)
	})

	t.Run("Failure - Main model not available", func(t *testing.T) {
		settingsService, db, mockDB, mockLLM := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		mockLLM.On("ListModels", ctx).Return(&llm.ListModelsResponse{
			Models: []llm.Model{{Name: "another-model"}},
		}, nil).Once()

		err := settingsService.Save(ctx, settingsToSave)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "main model 'model1' is not available")
		assert.NoError(t, mockDB.ExpectationsWereMet())
		mockLLM.AssertExpectations(t)
	})

	t.Run("Failure - LLM provider returns error", func(t *testing.T) {
		settingsService, db, mockDB, mockLLM := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		expectedErr := errors.New("ollama is down")
		mockLLM.On("ListModels", ctx).Return(nil, expectedErr).Once()

		err := settingsService.Save(ctx, settingsToSave)
		require.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
		assert.NoError(t, mockDB.ExpectationsWereMet())
		mockLLM.AssertExpectations(t)
	})
}
