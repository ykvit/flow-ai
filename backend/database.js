const sqlite3 = require('sqlite3').verbose();
const path = require('path');


const dbPath = path.resolve(__dirname, 'data', 'database.db');
console.log(`Database path: ${dbPath}`); 


const db = new sqlite3.Database(dbPath, (err) => {
    if (err) {
        console.error("Error opening database", err.message);
    } else {
        console.log("Connected to the SQLite database.");

        createTables();
    }
});


function createTables() {
    const createChatsTable = `
    CREATE TABLE IF NOT EXISTS chats (
        id TEXT PRIMARY KEY,
        title TEXT NOT NULL,
        createdAt TEXT NOT NULL,
        lastModified TEXT NOT NULL,
        model TEXT
    );`;

    const createMessagesTable = `
    CREATE TABLE IF NOT EXISTS messages (
        id TEXT PRIMARY KEY,
        chatId TEXT NOT NULL,
        role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system', 'ai')), 
        content TEXT NOT NULL,
        timestamp TEXT NOT NULL,
        FOREIGN KEY (chatId) REFERENCES chats(id) ON DELETE CASCADE
    );`;

    db.serialize(() => { 
        db.exec(createChatsTable, (err) => {
            if (err) console.error("Error creating chats table", err.message);
            else console.log("Chats table checked/created.");
        });
        db.exec(createMessagesTable, (err) => {
            if (err) console.error("Error creating messages table", err.message);
            else console.log("Messages table checked/created.");
        });
    });
}

module.exports = db; 

    