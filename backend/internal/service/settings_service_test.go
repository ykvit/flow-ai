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

// setupSettingsService is a test helper for `SettingsService` tests.
//
// WHY: This service has two distinct dependencies: a database connection (`*sql.DB`)
// and an LLM provider. This fixture creates mocks for *both*.
//   - `go-sqlmock` provides a mock `*sql.DB` (`db`) and a controller (`mockDB`)
//     to define expectations for SQL queries.
//   - `mockery` provides a mock for our `LLMProvider` interface (`mockLLM`).
//
// This setup allows complete isolation of the `SettingsService` logic.
func setupSettingsService(t *testing.T) (*service.SettingsService, *sql.DB, sqlmock.Sqlmock, *mocks.MockLLMProvider) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)

	mockLLM := mocks.NewMockLLMProvider(t)
	settingsService := service.NewSettingsService(db, mockLLM)

	return settingsService, db, mockDB, mockLLM
}

// TestSettingsService_Get tests the retrieval of existing settings, including the "self-healing" logic.
func TestSettingsService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Get existing settings", func(t *testing.T) {
		// ARRANGE
		settingsService, db, mockDB, _ := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		// Define the rows that our mock database should return for the SELECT query.
		rows := sqlmock.NewRows([]string{"key", "value"}).
			AddRow("system_prompt", "test prompt").
			AddRow("main_model", "test-model").
			AddRow("support_model", "support-model")

		// We expect a specific SQL query to be executed.
		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnRows(rows)

		// ACT
		settings, err := settingsService.Get(ctx)

		// ASSERT
		require.NoError(t, err)
		assert.Equal(t, "test prompt", settings.SystemPrompt)
		assert.Equal(t, "test-model", settings.MainModel)
		assert.Equal(t, "support-model", settings.SupportModel)

		// `ExpectationsWereMet` verifies that all expected SQL queries were executed.
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Success - Self-heal when main model is empty", func(t *testing.T) {
		// GOAL: Test the critical self-healing logic where the service automatically
		// selects a model if the configured one is missing.
		// ARRANGE
		settingsService, db, mockDB, mockLLM := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		// 1. Simulate the DB returning settings with an empty model.
		rows := sqlmock.NewRows([]string{"key", "value"}).
			AddRow("system_prompt", "test prompt").
			AddRow("main_model", "").
			AddRow("support_model", "")
		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnRows(rows)

		// 2. Simulate the LLM provider discovering an available model.
		mockLLM.On("ListModels", ctx).Return(&llm.ListModelsResponse{
			Models: []llm.Model{{Name: "discovered-model", ModifiedAt: time.Now().String()}},
		}, nil).Once()

		// 3. Expect the service to open a transaction and save the updated settings.
		// Note the deterministic order of inserts due to our code change.
		mockDB.ExpectBegin()
		prep := mockDB.ExpectPrepare("INSERT INTO settings")
		prep.ExpectExec().WithArgs("main_model", "discovered-model").WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().WithArgs("support_model", "discovered-model").WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().WithArgs("system_prompt", "test prompt").WillReturnResult(sqlmock.NewResult(1, 1))
		mockDB.ExpectCommit()

		// ACT
		settings, err := settingsService.Get(ctx)

		// ASSERT
		require.NoError(t, err)
		assert.Equal(t, "discovered-model", settings.MainModel)
		assert.Equal(t, "discovered-model", settings.SupportModel)
		assert.NoError(t, mockDB.ExpectationsWereMet())
		mockLLM.AssertExpectations(t)
	})

	t.Run("Failure - DB error on get", func(t *testing.T) {
		// ARRANGE: Simulate a database failure.
		settingsService, db, mockDB, _ := setupSettingsService(t)
		defer func() { _ = db.Close() }()
		expectedErr := errors.New("db error")
		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnError(expectedErr)

		// ACT
		settings, err := settingsService.Get(ctx)

		// ASSERT
		require.Error(t, err)
		assert.Nil(t, settings)
		assert.Contains(t, err.Error(), expectedErr.Error())
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})
}

// TestSettingsService_InitAndGet tests the first-run initialization logic.
func TestSettingsService_InitAndGet(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Settings already exist, just get them", func(t *testing.T) {
		// GOAL: Verify that if settings are found, initialization is skipped.
		settingsService, db, mockDB, _ := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		// `InitAndGet` internally calls `getFromDB` twice in this path.
		rows1 := sqlmock.NewRows([]string{"key", "value"}).AddRow("main_model", "existing-model")
		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnRows(rows1)
		rows2 := sqlmock.NewRows([]string{"key", "value"}).AddRow("main_model", "existing-model")
		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnRows(rows2)

		settings, err := settingsService.InitAndGet(ctx, "default")
		require.NoError(t, err)
		assert.Equal(t, "existing-model", settings.MainModel)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Success - No settings, initialize with discovered model", func(t *testing.T) {
		// GOAL: Test the core initialization path where the database is empty.
		settingsService, db, mockDB, mockLLM := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		// 1. Simulate DB returning "not found". `sql.ErrNoRows` is the canonical error for this.
		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnError(sql.ErrNoRows)
		// 2. Simulate the LLM provider finding a model.
		mockLLM.On("ListModels", ctx).Return(&llm.ListModelsResponse{Models: []llm.Model{{Name: "discovered-model"}}}, nil).Once()
		// 3. Expect the service to save the newly created default settings.
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
		// GOAL: Test the edge case where the application starts for the first time
		// and Ollama has no models downloaded yet.
		settingsService, db, mockDB, mockLLM := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		mockDB.ExpectQuery("SELECT key, value FROM settings").WillReturnError(sql.ErrNoRows)
		mockLLM.On("ListModels", ctx).Return(&llm.ListModelsResponse{Models: []llm.Model{}}, nil).Once()

		mockDB.ExpectBegin()
		prep := mockDB.ExpectPrepare("INSERT INTO settings")
		prep.ExpectExec().WithArgs("main_model", "").WillReturnResult(sqlmock.NewResult(1, 1)) // Expect empty strings
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

// TestSettingsService_Save tests the logic for updating settings.
func TestSettingsService_Save(t *testing.T) {
	ctx := context.Background()
	settingsToSave := &service.Settings{
		SystemPrompt: "new prompt",
		MainModel:    "model1",
		SupportModel: "model2",
	}

	t.Run("Success - Save valid settings", func(t *testing.T) {
		// GOAL: Verify that valid settings (where models exist) are saved correctly.
		settingsService, db, mockDB, mockLLM := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		// 1. Simulate the LLM provider confirming that both models are available.
		mockLLM.On("ListModels", ctx).Return(&llm.ListModelsResponse{
			Models: []llm.Model{{Name: "model1"}, {Name: "model2"}},
		}, nil).Once()

		// 2. Expect an UPSERT transaction.
		mockDB.ExpectBegin()
		// `regexp.QuoteMeta` is used because the query string contains special characters like `(?)`
		// that would otherwise be interpreted as a regex. This ensures we match the exact SQL string.
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
		// GOAL: Verify that the service rejects settings if the specified model does not exist.
		settingsService, db, mockDB, mockLLM := setupSettingsService(t)
		defer func() { _ = db.Close() }()

		mockLLM.On("ListModels", ctx).Return(&llm.ListModelsResponse{
			Models: []llm.Model{{Name: "another-model"}}, // "model1" is missing
		}, nil).Once()

		err := settingsService.Save(ctx, settingsToSave)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "main model 'model1' is not available")

		// Crucially, no DB calls should be made if validation fails.
		assert.NoError(t, mockDB.ExpectationsWereMet())
		mockLLM.AssertExpectations(t)
	})

	t.Run("Failure - LLM provider returns error", func(t *testing.T) {
		// GOAL: Verify that errors from the LLM provider are handled gracefully.
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
