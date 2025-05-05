const express = require('express');
const cors = require('cors');
const db = require('./database.js');
const { v4: uuidv4 } = require('uuid');

const app = express();
const port = 3001;

app.use(cors());
app.use(express.json());

// --- API Endpoints ---

// GET /api/chats
app.get('/api/chats', (req, res) => {
    console.log("Received GET /api/chats request");
    const sql = "SELECT id, title, createdAt, lastModified, model FROM chats ORDER BY lastModified DESC";
    db.all(sql, [], (err, rows) => {
        if (err) {
            console.error("!!! Error fetching chats list from DB:", err.message);
            res.status(500).json({ "error": err.message });
            return;
        }
        console.log(`Successfully fetched ${rows.length} chats from DB`);
        try {
            const chats = rows.map(chat => ({ ...chat }));
            res.json({ chats });
        } catch (mapError) {
             console.error("!!! Error processing chat rows:", mapError.message);
             res.status(500).json({ "error": "Error processing chat data" });
        }
    });
});

// GET /api/chats/:chatId
app.get('/api/chats/:chatId', (req, res) => {
    const chatId = req.params.chatId;
    console.log(`Received GET /api/chats/${chatId} request`);
    const sqlChat = "SELECT id, title, createdAt, lastModified, model FROM chats WHERE id = ?";
    const sqlMessages = `SELECT id, chatId, role, content, timestamp FROM messages WHERE chatId = ? ORDER BY timestamp ASC`;

    db.get(sqlChat, [chatId], (err, chatRow) => {
        if (err) { res.status(500).json({ "error": err.message }); return; }
        if (!chatRow) { res.status(404).json({ "error": "Chat not found" }); return; }
        console.log(`Successfully fetched chat metadata for ${chatId}`);

        db.all(sqlMessages, [chatId], (err, messageRows) => {
            if (err) { res.status(500).json({ "error": err.message }); return; }
            console.log(`Successfully fetched ${messageRows.length} messages for chat ${chatId}`);

            const chat = {
                id: chatRow.id,
                title: chatRow.title,
                createdAt: chatRow.createdAt,
                lastModified: chatRow.lastModified,
                model: chatRow.model,
                messages: messageRows.map(msg => {
                    const senderRole = msg.role === 'assistant' ? 'ai' : 'user';
                    return { sender: senderRole, text: msg.content };
                })
            };
            res.json({ chat });
        });
    });
});

// POST /api/chats
app.post('/api/chats', (req, res) => {
    console.log("Received POST /api/chats request");
    const { title = "New Chat", model = "unknown" } = req.body;
    const newChatId = uuidv4();
    const nowISO = new Date().toISOString();

    const newChatData = {
        id: newChatId,
        title: title,
        createdAt: nowISO,
        lastModified: nowISO,
        model: model ?? "unknown"
    };

    const sql = `INSERT INTO chats (id, title, createdAt, lastModified, model) VALUES (?, ?, ?, ?, ?)`;
    const params = [newChatData.id, newChatData.title, newChatData.createdAt, newChatData.lastModified, newChatData.model];

    db.run(sql, params, function (err) {
        if (err) {
            console.error("!!! Error creating new chat:", err.message);
            res.status(500).json({ "error": "Failed to create chat", "details": err.message });
            return;
        }
        console.log(`Successfully created new chat with ID: ${newChatData.id}`);
        const createdChatForResponse = {
             ...newChatData,
             messages: [],

        };
        res.status(201).json({ chat: createdChatForResponse });
    });
});

// POST /api/chats/:chatId/messages
app.post('/api/chats/:chatId/messages', (req, res) => {
    const chatId = req.params.chatId;
    const { messages: messagesToAdd, title: newTitle } = req.body;
    console.log(`Received POST /api/chats/${chatId}/messages request`);

    if (!Array.isArray(messagesToAdd) || messagesToAdd.length === 0) {
        return res.status(400).json({ "error": "Request body must contain a non-empty 'messages' array" });
    }

    const nowISO = new Date().toISOString();
    const sqlInsertMsg = `INSERT INTO messages (id, chatId, role, content, timestamp) VALUES (?, ?, ?, ?, ?)`;
    const sqlUpdateChat = `UPDATE chats SET lastModified = ? ${newTitle ? ', title = ?' : ''} WHERE id = ?`;
    const updateChatParams = newTitle ? [nowISO, newTitle, chatId] : [nowISO, chatId];

    db.serialize(() => {
        db.run("BEGIN TRANSACTION;");
        let errorOccurred = null;

        messagesToAdd.forEach(msg => {
            if (errorOccurred) return;
            if (!msg || typeof msg.content !== 'string' || !['user', 'assistant'].includes(msg.role)) {
                 errorOccurred = new Error(`Invalid message format: ${JSON.stringify(msg)}`); 
                 console.error(errorOccurred.message);
                 return;
            }
            const messageId = uuidv4();
            const params = [messageId, chatId, msg.role, msg.content, nowISO];
            db.run(sqlInsertMsg, params, function(err) {
                if (err) { errorOccurred = err; console.error("!!! Error inserting message:", err.message); }
                else { console.log(`Inserted message ${messageId} for chat ${chatId}`); }
            });
        });

        if (!errorOccurred) {
            db.run(sqlUpdateChat, updateChatParams, function(err) {
                if (err) { errorOccurred = err; console.error("!!! Error updating chat metadata:", err.message); }
                else if (this.changes === 0) { errorOccurred = new Error(`Chat ${chatId} not found during update.`); console.error(errorOccurred.message); }
                else { console.log(`Updated metadata for chat ${chatId}`); }
            });
        }

        if (errorOccurred) {
            db.run("ROLLBACK;");
            console.error("Transaction rolled back:", errorOccurred.message);
            const statusCode = errorOccurred.message.includes("not found") ? 404 : 500;
            res.status(statusCode).json({ "error": "Failed to add messages", "details": errorOccurred.message });
        } else {
            db.run("COMMIT;");
            console.log(`Transaction committed successfully for chat ${chatId}`);
            res.status(201).json({ message: "Messages added successfully" });
        }
    });
});

// DELETE /api/chats/:chatId
app.delete('/api/chats/:chatId', (req, res) => {
    const chatId = req.params.chatId;
    console.log(`Received DELETE /api/chats/${chatId} request`);
    const sql = 'DELETE FROM chats WHERE id = ?';
    db.run(sql, [chatId], function (err) {
        if (err) { console.error(`!!! Error deleting chat ${chatId}:`, err.message); res.status(500).json({ "error": "Failed to delete chat", "details": err.message }); return; }
        if (this.changes === 0) { console.log(`Chat ${chatId} not found for deletion`); res.status(404).json({ "error": "Chat not found" }); return; }
        console.log(`Successfully deleted chat ${chatId}, changes: ${this.changes}`);
        res.status(204).send();
    });
});


// PATCH /api/chats/:chatId (to update the chat name)
app.patch('/api/chats/:chatId', (req, res) => {
    const chatId = req.params.chatId;
    const { title: newTitle } = req.body; 

    console.log(`Received PATCH /api/chats/${chatId} request with new title: "${newTitle}"`);

    // Checking whether the correct name was passed
    if (typeof newTitle !== 'string' || newTitle.trim() === '') {
        console.error("!!! Invalid or missing title in request body");
        return res.status(400).json({ "error": "Request body must contain a valid 'title' string" });
    }

    const trimmedNewTitle = newTitle.trim();
    const nowISO = new Date().toISOString(); 

    const sql = `UPDATE chats SET title = ?, lastModified = ? WHERE id = ?`;
    const params = [trimmedNewTitle, nowISO, chatId];

    db.run(sql, params, function (err) {
        if (err) {
            console.error(`!!! Error updating chat title for ${chatId}:`, err.message);
            res.status(500).json({ "error": "Failed to update chat title", "details": err.message });
            return;
        }
        if (this.changes === 0) {
            console.log(`Chat ${chatId} not found for title update`);
            res.status(404).json({ "error": "Chat not found" }); 
            return;
        }


        console.log(`Successfully updated title for chat ${chatId}, changes: ${this.changes}`);
        res.status(200).json({ message: "Chat title updated successfully" });
       
    });
});


app.listen(port, () => {
    console.log(`Backend server running on http://localhost:${port}`);
});

process.on('SIGINT', () => {
    db.close((err) => {
        if (err) { return console.error(err.message); }
        console.log('Closed the database connection.');
        process.exit(0);
    });
});