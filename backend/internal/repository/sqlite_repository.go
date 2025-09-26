package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"flow-ai/backend/internal/model"
)

// queryable is a helper interface that allows using both *sql.DB and *sql.Tx.
// This helps to avoid code duplication for read operations.
type queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type sqliteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) Repository {
	return &sqliteRepository{db: db}
}

func (r *sqliteRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// --- Chat Methods ---
func (r *sqliteRepository) CreateChat(ctx context.Context, chat *model.Chat) error {
	query := "INSERT INTO chats (id, user_id, title, model, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := r.db.ExecContext(ctx, query, chat.ID, chat.UserID, chat.Title, chat.Model, chat.CreatedAt, chat.UpdatedAt)
	return err
}

func (r *sqliteRepository) GetChat(ctx context.Context, chatID string) (*model.Chat, error) {
	query := "SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = ?"
	row := r.db.QueryRowContext(ctx, query, chatID)
	var chat model.Chat
	err := row.Scan(&chat.ID, &chat.UserID, &chat.Title, &chat.Model, &chat.CreatedAt, &chat.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chat not found")
		}
		return nil, err
	}
	return &chat, nil
}

func (r *sqliteRepository) GetChats(ctx context.Context, userID string) ([]*model.Chat, error) {
	query := "SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE user_id = ? ORDER BY updated_at DESC"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*model.Chat
	for rows.Next() {
		var chat model.Chat
		if err := rows.Scan(&chat.ID, &chat.UserID, &chat.Title, &chat.Model, &chat.CreatedAt, &chat.UpdatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}
	return chats, nil
}

func (r *sqliteRepository) UpdateChatTitle(ctx context.Context, chatID, newTitle string) error {
	query := "UPDATE chats SET title = ?, updated_at = ? WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, newTitle, time.Now().UTC(), chatID)
	return err
}

func (r *sqliteRepository) DeleteChat(ctx context.Context, chatID string) error {
	query := "DELETE FROM chats WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, chatID)
	return err
}

// --- Message Methods ---

func (r *sqliteRepository) AddMessage(ctx context.Context, message *model.Message, chatID string) error {
	tx, err := r.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := r.AddMessageTx(ctx, tx, message, chatID); err != nil {
		return fmt.Errorf("could not add message transactionally: %w", err)
	}
	if err := r.UpdateChatTimestampTx(ctx, tx, chatID); err != nil {
		return fmt.Errorf("could not update chat timestamp transactionally: %w", err)
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
	var metadata, context, parentID, modelName sql.NullString

	err := row.Scan(&msg.ID, &chatID, &parentID, &msg.Role, &msg.Content, &modelName, &msg.Timestamp, &metadata, &context)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message with ID %s not found", messageID)
		}
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

	return &msg, nil
}

// GetActiveMessagesByChatID is the public non-transactional method.
func (r *sqliteRepository) GetActiveMessagesByChatID(ctx context.Context, chatID string) ([]model.Message, error) {
	return r.getActiveMessagesByChatID(ctx, r.db, chatID)
}

// GetActiveMessagesByChatIDTx is the public transactional method.
func (r *sqliteRepository) GetActiveMessagesByChatIDTx(ctx context.Context, tx *sql.Tx, chatID string) ([]model.Message, error) {
	return r.getActiveMessagesByChatID(ctx, tx, chatID)
}

// getActiveMessagesByChatID is a private helper that works with both *sql.DB and *sql.Tx.
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
	defer rows.Close()

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
		if err == sql.ErrNoRows {
			return nil, nil // Not an error, just no messages
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

func (r *sqliteRepository) AddMessageTx(ctx context.Context, tx *sql.Tx, message *model.Message, chatID string) error {
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
		true,
	)
	return err
}

func (r *sqliteRepository) DeactivateBranchTx(ctx context.Context, tx *sql.Tx, messageID string) error {
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