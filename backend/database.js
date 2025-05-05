const sqlite3 = require('sqlite3').verbose();
const path = require('path');
const fs = require('fs');

const dbDir = path.resolve(__dirname, 'data');
const dbPath = path.join(dbDir, 'database.db');


if (!fs.existsSync(dbDir)) {
    try {
        fs.mkdirSync(dbDir, { recursive: true });
        console.log(`Created directory: ${dbDir}`);
    } catch (mkdirError) {
        console.error(`!!! Error creating directory ${dbDir}:`, mkdirError.message);
        process.exit(1); 
    }
}

console.log(`Database path: ${dbPath}`);

const db = new sqlite3.Database(dbPath, (err) => {
    if (err) {
        console.error("!!! Error opening database", err.message);
    } else {
        console.log("Connected to the SQLite database.");
        db.exec('PRAGMA foreign_keys = ON;', (pragmaErr) => {
            if (pragmaErr) {
                console.error("!!! Error enabling foreign keys", pragmaErr.message);
            } else {
                console.log("Foreign key support enabled.");
                createTables();
            }
        });
    }
});

function createTables() {
    const createChatsTable = `
    CREATE TABLE IF NOT EXISTS chats (
        id TEXT PRIMARY KEY,
        title TEXT NOT NULL,
        createdAt TEXT NOT NULL,
        lastModified TEXT NOT NULL,
        model TEXT NOT NULL,           -- Зроблено NOT NULL, модель важлива
        systemPrompt TEXT NULL,        -- Нове поле: Системний промпт
        ollamaOptions TEXT NULL        -- Нове поле: Параметри Ollama (JSON рядок)
    );`;

    const createChatsIndex = `
    CREATE INDEX IF NOT EXISTS idx_chats_lastModified ON chats (lastModified DESC);
    `;

    const createMessagesTable = `
    CREATE TABLE IF NOT EXISTS messages (
        id TEXT PRIMARY KEY,
        chatId TEXT NOT NULL,
        role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system', 'tool')), -- Оновлений CHECK constraint
        content TEXT NOT NULL,
        timestamp TEXT NOT NULL,
        FOREIGN KEY (chatId) REFERENCES chats(id) ON DELETE CASCADE -- Забезпечує видалення повідомлень
    );`;


    const createMessagesChatIdIndex = `
    CREATE INDEX IF NOT EXISTS idx_messages_chatId ON messages (chatId);
    `;
    const createMessagesTimestampIndex = `
    CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages (timestamp ASC);
    `;

    db.serialize(() => {
        console.log("Executing database schema setup...");
        db.exec(createChatsTable, (err) => {
            if (err) console.error("!!! Error creating chats table", err.message);
            else console.log("Chats table checked/created.");
        });
        db.exec(createChatsIndex, (err) => {
            if (err) console.error("!!! Error creating chats index", err.message);
            else console.log("Chats index checked/created.");
        });
        db.exec(createMessagesTable, (err) => {
            if (err) console.error("!!! Error creating messages table", err.message);
            else console.log("Messages table checked/created.");
        });
        db.exec(createMessagesChatIdIndex, (err) => {
            if (err) console.error("!!! Error creating messages chatId index", err.message);
            else console.log("Messages chatId index checked/created.");
        });
        db.exec(createMessagesTimestampIndex, (err) => {
            if (err) console.error("!!! Error creating messages timestamp index", err.message);
            else console.log("Messages timestamp index checked/created.");
        });
        console.log("Database schema setup finished.");
    });
}

module.exports = db;
