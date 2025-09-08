package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"flow-ai/backend/internal/model"
)

type sqliteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) Repository {
	return &sqliteRepository{db: db}
}

// ... (методи CreateChat, GetChat, GetChats, UpdateChatTitle, DeleteChat залишаються без змін) ...
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

// REFACTORED: This method now uses a transaction to ensure data integrity.
func (r *sqliteRepository) AddMessage(ctx context.Context, message *model.Message, chatID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	// Ensure transaction is rolled back on error
	defer tx.Rollback()

	var metadata sql.NullString
	if len(message.Metadata) > 0 && string(message.Metadata) != "null" {
		metadata.String = string(message.Metadata)
		metadata.Valid = true
	}

	insertMsgQuery := `
		INSERT INTO messages (id, chat_id, parent_id, role, content, model, timestamp, metadata, context, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, insertMsgQuery,
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
	if err != nil {
		return fmt.Errorf("could not insert message: %w", err)
	}
	
	updateChatQuery := "UPDATE chats SET updated_at = ? WHERE id = ?"
	_, err = tx.ExecContext(ctx, updateChatQuery, time.Now().UTC(), chatID)
	if err != nil {
		return fmt.Errorf("could not update chat timestamp: %w", err)
	}

	return tx.Commit()
}

// ... (решта методів залишаються без змін) ...
func (r *sqliteRepository) GetActiveMessagesByChatID(ctx context.Context, chatID string) ([]model.Message, error) {
	query := `
		SELECT id, parent_id, role, content, model, timestamp, metadata, context
		FROM messages
		WHERE chat_id = ? AND is_active = TRUE
		ORDER BY timestamp ASC
	`
	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		var metadata sql.NullString
		var context sql.NullString
		var parentID sql.NullString
		var modelName sql.NullString

		if err := rows.Scan(&msg.ID, &parentID, &msg.Role, &msg.Content, &modelName, &msg.Timestamp, &metadata, &context); err != nil {
			return nil, err
		}
		
		if parentID.Valid { msg.ParentID = &parentID.String }
		if modelName.Valid { msg.Model = &modelName.String }
		if metadata.Valid { msg.Metadata = json.RawMessage(metadata.String) }
		if context.Valid { msg.Context = json.RawMessage(context.String) }
		
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
			return nil, nil
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