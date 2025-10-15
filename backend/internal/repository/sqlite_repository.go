package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"flow-ai/backend/internal/model"
)

// queryable is a helper interface that abstracts over `*sql.DB` and `*sql.Tx`.
// This allows methods like `getActiveMessagesByChatID` to be reused both inside
// and outside of explicit transactions.
type queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// sqliteRepository is the concrete implementation of the Repository interface for SQLite.
type sqliteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository instance.
func NewSQLiteRepository(db *sql.DB) Repository {
	return &sqliteRepository{db: db}
}

// BeginTx starts a new database transaction.
func (r *sqliteRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// --- Chat Methods ---

func (r *sqliteRepository) CreateChat(ctx context.Context, chat *model.Chat) error {
	query := "INSERT INTO chats (id, title, model, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"
	_, err := r.db.ExecContext(ctx, query, chat.ID, chat.Title, chat.Model, chat.CreatedAt, chat.UpdatedAt)
	return err
}

func (r *sqliteRepository) GetChat(ctx context.Context, chatID string) (*model.Chat, error) {
	query := "SELECT id, title, model, created_at, updated_at FROM chats WHERE id = ?"
	row := r.db.QueryRowContext(ctx, query, chatID)
	var chat model.Chat
	err := row.Scan(&chat.ID, &chat.Title, &chat.Model, &chat.CreatedAt, &chat.UpdatedAt)
	if err != nil {
		// Abstract away the driver-specific error.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &chat, nil
}

func (r *sqliteRepository) GetChats(ctx context.Context) ([]*model.Chat, error) {
	// In the current single-user model, this fetches all chats without filtering.
	// The query is intentionally simple.
	query := "SELECT id, title, model, created_at, updated_at FROM chats ORDER BY updated_at DESC"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	// `defer rows.Close()` is crucial to prevent leaking database connections.
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("Failed to close rows in GetChats", "error", err)
		}
	}()

	var chats []*model.Chat
	for rows.Next() {
		var chat model.Chat
		if err := rows.Scan(&chat.ID, &chat.Title, &chat.Model, &chat.CreatedAt, &chat.UpdatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}
	return chats, nil
}

func (r *sqliteRepository) UpdateChatTitle(ctx context.Context, chatID, newTitle string) error {
	query := "UPDATE chats SET title = ?, updated_at = ? WHERE id = ?"
	res, err := r.db.ExecContext(ctx, query, newTitle, time.Now().UTC(), chatID)
	if err != nil {
		return err
	}
	// For UPDATE or DELETE operations, it's good practice to check if any rows
	// were actually affected. If not, the target entity likely didn't exist.
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqliteRepository) DeleteChat(ctx context.Context, chatID string) error {
	query := "DELETE FROM chats WHERE id = ?"
	res, err := r.db.ExecContext(ctx, query, chatID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// --- Message Methods ---

// AddMessage wraps the core logic in a transaction to ensure atomicity.
// Adding a message and updating the parent chat's timestamp should succeed or fail together.
func (r *sqliteRepository) AddMessage(ctx context.Context, message *model.Message, chatID string) error {
	tx, err := r.BeginTx(ctx)
	if err != nil {
		return err
	}
	// `defer tx.Rollback()` is a safeguard. If `tx.Commit()` is called, the
	// rollback becomes a no-op. If any error occurs, the rollback is executed.
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("Failed to rollback AddMessage transaction", "error", err)
		}
	}()

	if err := r.AddMessageTx(ctx, tx, message, chatID); err != nil {
		return err
	}
	if err := r.UpdateChatTimestampTx(ctx, tx, chatID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *sqliteRepository) GetMessageByID(ctx context.Context, messageID string) (*model.Message, error) {
	query := `
		SELECT id, chat_id, parent_id, role, content, model, timestamp, metadata, context
		FROM messages
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, messageID)
	var msg model.Message
	var chatID string
	// Use `sql.NullString` and `sql.Null` types for columns that can be NULL
	// to prevent `Scan` from failing.
	var metadata, context, parentID, modelName sql.NullString

	err := row.Scan(&msg.ID, &chatID, &parentID, &msg.Role, &msg.Content, &modelName, &msg.Timestamp, &metadata, &context)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Safely assign values from nullable columns to the struct fields.
	if parentID.Valid {
		msg.ParentID = &parentID.String
	}
	if modelName.Valid {
		msg.Model = &modelName.String
	}
	if metadata.Valid {
		msg.Metadata = json.RawMessage(metadata.String)
	}
	if context.Valid {
		msg.Context = json.RawMessage(context.String)
	}

	return &msg, nil
}

func (r *sqliteRepository) GetActiveMessagesByChatID(ctx context.Context, chatID string) ([]model.Message, error) {
	return r.getActiveMessagesByChatID(ctx, r.db, chatID)
}

func (r *sqliteRepository) GetActiveMessagesByChatIDTx(ctx context.Context, tx *sql.Tx, chatID string) ([]model.Message, error) {
	return r.getActiveMessagesByChatID(ctx, tx, chatID)
}

// getActiveMessagesByChatID is a private helper that can run on either a `*sql.DB` or `*sql.Tx`.
func (r *sqliteRepository) getActiveMessagesByChatID(ctx context.Context, q queryable, chatID string) ([]model.Message, error) {
	query := `
		SELECT id, parent_id, role, content, model, timestamp, metadata, context
		FROM messages
		WHERE chat_id = ? AND is_active = TRUE
		ORDER BY timestamp ASC
	`
	rows, err := q.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("Failed to close rows in getActiveMessagesByChatID", "error", err)
		}
	}()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		var metadata, context, parentID, modelName sql.NullString

		if err := rows.Scan(&msg.ID, &parentID, &msg.Role, &msg.Content, &modelName, &msg.Timestamp, &metadata, &context); err != nil {
			return nil, err
		}

		if parentID.Valid {
			msg.ParentID = &parentID.String
		}
		if modelName.Valid {
			msg.Model = &modelName.String
		}
		if metadata.Valid {
			msg.Metadata = json.RawMessage(metadata.String)
		}
		if context.Valid {
			msg.Context = json.RawMessage(context.String)
		}

		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *sqliteRepository) GetLastActiveMessage(ctx context.Context, chatID string) (*model.Message, error) {
	query := `
		SELECT id, context
		FROM messages
		WHERE chat_id = ? AND is_active = TRUE
		ORDER BY timestamp DESC LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, chatID)

	var msg model.Message
	var context sql.NullString
	err := row.Scan(&msg.ID, &context)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// In this specific case, returning `ErrNotFound` is more semantically
			// correct than `nil, nil` for the service layer to interpret.
			return nil, ErrNotFound
		}
		return nil, err
	}

	if context.Valid {
		msg.Context = json.RawMessage(context.String)
	}

	return &msg, nil
}

func (r *sqliteRepository) UpdateMessageContext(ctx context.Context, messageID string, ollamaContext []byte) error {
	query := "UPDATE messages SET context = ? WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, ollamaContext, messageID)
	return err
}

// --- Transactional Methods ---
// These methods expect to be passed an existing transaction `*sql.Tx` and do not commit or rollback.
// This allows them to be composed into larger atomic operations.

func (r *sqliteRepository) AddMessageTx(ctx context.Context, tx *sql.Tx, message *model.Message, chatID string) error {
	// Handle empty or "null" JSON from the model layer gracefully.
	var metadata sql.NullString
	if len(message.Metadata) > 0 && string(message.Metadata) != "null" {
		metadata.String = string(message.Metadata)
		metadata.Valid = true
	}

	insertMsgQuery := `
		INSERT INTO messages (id, chat_id, parent_id, role, content, model, timestamp, metadata, context, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := tx.ExecContext(ctx, insertMsgQuery,
		message.ID,
		chatID,
		message.ParentID,
		message.Role,
		message.Content,
		message.Model,
		message.Timestamp,
		metadata,
		message.Context,
		true, // New messages are always active.
	)
	return err
}

// DeactivateBranchTx performs a recursive update to mark a message and all its
// descendants as inactive. This is the core of the "regeneration" logic.
func (r *sqliteRepository) DeactivateBranchTx(ctx context.Context, tx *sql.Tx, messageID string) error {
	// Common Table Expression (CTE) for recursive traversal is efficient in SQLite.
	query := `
		WITH RECURSIVE branch_ids(id) AS (
			VALUES(?)
			UNION ALL
			SELECT m.id FROM messages m JOIN branch_ids b ON m.parent_id = b.id
		)
		UPDATE messages SET is_active = FALSE WHERE id IN (SELECT id FROM branch_ids);
	`
	_, err := tx.ExecContext(ctx, query, messageID)
	return err
}

func (r *sqliteRepository) UpdateChatTimestampTx(ctx context.Context, tx *sql.Tx, chatID string) error {
	query := "UPDATE chats SET updated_at = ? WHERE id = ?"
	_, err := tx.ExecContext(ctx, query, time.Now().UTC(), chatID)
	return err
}
