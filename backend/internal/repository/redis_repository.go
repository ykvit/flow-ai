package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"flow-ai/backend/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

// Repository interface now includes DeleteChat
type Repository interface {
	CreateChat(ctx context.Context, chat *model.Chat) error
	GetChat(ctx context.Context, chatID string) (*model.Chat, error)
	GetChats(ctx context.Context, userID string) ([]*model.Chat, error)
	UpdateChat(ctx context.Context, chat *model.Chat) error
	UpdateChatTitle(ctx context.Context, chatID, newTitle string) error
	DeleteChat(ctx context.Context, chatID string) error // New method
	AddMessage(ctx context.Context, chatID string, message *model.Message) error
	GetMessages(ctx context.Context, chatID string) ([]model.Message, error)
	GetOllamaContext(ctx context.Context, chatID string) ([]byte, error)
	SetOllamaContext(ctx context.Context, chatID string, ollamaContext []byte) error
}

type redisRepository struct {
	rdb *redis.Client
}

func NewRedisRepository(rdb *redis.Client) Repository {
	return &redisRepository{rdb: rdb}
}

// Key Generation Helpers
func (r *redisRepository) chatKey(chatID string) string      { return fmt.Sprintf("chat:%s", chatID) }
func (r *redisRepository) messagesKey(chatID string) string  { return fmt.Sprintf("chat:%s:messages", chatID) }
func (r *redisRepository) contextKey(chatID string) string   { return fmt.Sprintf("chat:%s:context", chatID) }
func (r *redisRepository) messageKey(messageID string) string{ return fmt.Sprintf("message:%s", messageID) }
func (r *redisRepository) userChatsKey(userID string) string { return fmt.Sprintf("user:%s:chats", userID) }

// --- Chat Operations ---
func (r *redisRepository) CreateChat(ctx context.Context, chat *model.Chat) error {
	chatMap, err := structToMap(chat)
	if err != nil { return fmt.Errorf("could not convert chat to map: %w", err) }
	pipe := r.rdb.TxPipeline()
	pipe.HSet(ctx, r.chatKey(chat.ID), chatMap)
	pipe.ZAdd(ctx, r.userChatsKey(chat.UserID), redis.Z{Score: float64(-chat.UpdatedAt.UnixNano()), Member: chat.ID})
	_, err = pipe.Exec(ctx)
	return err
}

func (r *redisRepository) GetChat(ctx context.Context, chatID string) (*model.Chat, error) {
	chatMap, err := r.rdb.HGetAll(ctx, r.chatKey(chatID)).Result()
	if err != nil { return nil, err }
	if len(chatMap) == 0 { return nil, redis.Nil }
	var chat model.Chat
	return &chat, mapToStruct(chatMap, &chat)
}

func (r *redisRepository) GetChats(ctx context.Context, userID string) ([]*model.Chat, error) {
	chatIDs, err := r.rdb.ZRange(ctx, r.userChatsKey(userID), 0, -1).Result()
	if err != nil { return nil, err }
	chats := make([]*model.Chat, 0, len(chatIDs))
	for _, id := range chatIDs {
		chat, err := r.GetChat(ctx, id)
		if err == nil && chat != nil { chats = append(chats, chat) }
	}
	return chats, nil
}

func (r *redisRepository) UpdateChat(ctx context.Context, chat *model.Chat) error {
	return r.CreateChat(ctx, chat)
}

func (r *redisRepository) UpdateChatTitle(ctx context.Context, chatID, newTitle string) error {
    pipe := r.rdb.TxPipeline()
    key := r.chatKey(chatID)
    now := time.Now()
    
    pipe.HSet(ctx, key, "title", newTitle)
    pipe.HSet(ctx, key, "updated_at", now.Format(time.RFC3339Nano))

    userID, err := r.rdb.HGet(ctx, key, "user_id").Result()
    if err == nil && userID != "" {
        pipe.ZAdd(ctx, r.userChatsKey(userID), redis.Z{
            Score: float64(-now.UnixNano()),
            Member: chatID,
        })
    }

	_, err = pipe.Exec(ctx)
	return err
}

func (r *redisRepository) DeleteChat(ctx context.Context, chatID string) error {
	chat, err := r.GetChat(ctx, chatID)
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("could not get chat for deletion: %w", err)
	}

	msgIDs, err := r.rdb.ZRange(ctx, r.messagesKey(chatID), 0, -1).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("could not get message IDs for deletion: %w", err)
	}

	pipe := r.rdb.TxPipeline()

	if len(msgIDs) > 0 {
		messageKeys := make([]string, len(msgIDs))
		for i, id := range msgIDs {
			messageKeys[i] = r.messageKey(id)
		}
		pipe.Del(ctx, messageKeys...)
	}

	pipe.Del(ctx, r.chatKey(chatID))
	pipe.Del(ctx, r.messagesKey(chatID))
	pipe.Del(ctx, r.contextKey(chatID))
	pipe.ZRem(ctx, r.userChatsKey(chat.UserID), chatID)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute chat deletion pipeline: %w", err)
	}

	return nil
}

// --- Message Operations ---
func (r *redisRepository) AddMessage(ctx context.Context, chatID string, message *model.Message) error {
	msgMap, err := structToMap(message)
	if err != nil { return fmt.Errorf("could not convert message to map: %w", err) }
	pipe := r.rdb.TxPipeline()
	pipe.HSet(ctx, r.messageKey(message.ID), msgMap)
	pipe.ZAdd(ctx, r.messagesKey(chatID), redis.Z{Score: float64(message.Timestamp.UnixNano()), Member: message.ID})
	_, err = pipe.Exec(ctx)
	return err
}

func (r *redisRepository) GetMessages(ctx context.Context, chatID string) ([]model.Message, error) {
	msgIDs, err := r.rdb.ZRange(ctx, r.messagesKey(chatID), 0, -1).Result()
	if err != nil {
		if err == redis.Nil { return []model.Message{}, nil }
		return nil, err
	}
	if len(msgIDs) == 0 {
		return []model.Message{}, nil
	}
	messages := make([]model.Message, 0, len(msgIDs))
	for _, id := range msgIDs {
		msgMap, err := r.rdb.HGetAll(ctx, r.messageKey(id)).Result()
		if err != nil { continue }
		var msg model.Message
		if err := mapToStruct(msgMap, &msg); err == nil {
			messages = append(messages, msg)
		}
	}
	return messages, nil
}

// --- Context Operations ---
func (r *redisRepository) GetOllamaContext(ctx context.Context, chatID string) ([]byte, error) {
	return r.rdb.Get(ctx, r.contextKey(chatID)).Bytes()
}

func (r *redisRepository) SetOllamaContext(ctx context.Context, chatID string, ollamaContext []byte) error {
	return r.rdb.Set(ctx, r.contextKey(chatID), ollamaContext, 24*time.Hour).Err()
}

// --- Helper Functions ---
func structToMap(obj interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(obj)
	if err != nil { return nil, err }
	var mapData map[string]interface{}
	return mapData, json.Unmarshal(data, &mapData)
}

func mapToStruct(data map[string]string, obj interface{}) error {
	jsonStr, err := json.Marshal(data)
	if err != nil { return err }
	return json.Unmarshal(jsonStr, obj)
}