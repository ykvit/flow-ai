-- Down migration for initial schema
DROP TABLE IF EXISTS chats;

DROP INDEX IF EXISTS idx_messages_chat_id_active_timestamp;
DROP INDEX IF EXISTS idx_messages_parent_id;
DROP TABLE IF EXISTS messages;

DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS settings;
