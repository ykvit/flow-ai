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
            res.json({ chats: rows });
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
    const sqlChat = "SELECT id, title, createdAt, lastModified, model, systemPrompt, ollamaOptions FROM chats WHERE id = ?";
    const sqlMessages = `SELECT id, chatId, role, content, timestamp FROM messages WHERE chatId = ? ORDER BY timestamp ASC`;

    db.get(sqlChat, [chatId], (err, chatRow) => {
        if (err) {
            console.error(`!!! Error fetching chat ${chatId}:`, err.message);
            res.status(500).json({ "error": err.message });
            return;
        }
        if (!chatRow) {
            console.log(`Chat ${chatId} not found`);
            res.status(404).json({ "error": "Chat not found" });
            return;
        }
        console.log(`Successfully fetched chat metadata for ${chatId}`);

        db.all(sqlMessages, [chatId], (err, messageRows) => {
            if (err) {
                console.error(`!!! Error fetching messages for chat ${chatId}:`, err.message);
                res.status(500).json({ "error": err.message });
                return;
            }
            console.log(`Successfully fetched ${messageRows.length} messages for chat ${chatId}`);

            try {
                const chat = {
                    id: chatRow.id,
                    title: chatRow.title,
                    createdAt: chatRow.createdAt,
                    lastModified: chatRow.lastModified,
                    model: chatRow.model,
                    systemPrompt: chatRow.systemPrompt, 
                    ollamaOptions: chatRow.ollamaOptions ? JSON.parse(chatRow.ollamaOptions) : null, 
                    messages: messageRows.map(msg => ({
                        id: msg.id,
                        sender: msg.role,
                        text: msg.content,
                        timestamp: msg.timestamp
                    }))
                };
                res.json({ chat });
            } catch (parseError) {
                 console.error(`!!! Error parsing ollamaOptions for chat ${chatId}:`, parseError.message);
                 res.status(500).json({ "error": "Error processing chat data (parsing options)" });
            }
        });
    });
});

// POST /api/chats 
app.post('/api/chats', (req, res) => {
    console.log("Received POST /api/chats request");
    const {
        title = "New Chat",
        model, 
        systemPrompt = null, 
        ollamaOptions = null 
    } = req.body;

    if (!model || typeof model !== 'string' || model.trim() === '') {
         console.error("!!! Missing or invalid 'model' in request body");
         return res.status(400).json({ "error": "Request body must contain a valid 'model' string" });
    }

    const newChatId = uuidv4();
    const nowISO = new Date().toISOString();

    let optionsString = null;
    if (ollamaOptions && typeof ollamaOptions === 'object') {
        try {
            optionsString = JSON.stringify(ollamaOptions);
        } catch (stringifyError) {
             console.error("!!! Error stringifying ollamaOptions:", stringifyError.message);
             return res.status(400).json({ "error": "Invalid 'ollamaOptions' format" });
        }
    } else if (typeof ollamaOptions === 'string') {
         optionsString = ollamaOptions; 
    }


    const newChatData = {
        id: newChatId,
        title: title.trim(),
        createdAt: nowISO,
        lastModified: nowISO,
        model: model.trim(),
        systemPrompt: systemPrompt ? systemPrompt.trim() : null,
        ollamaOptions: optionsString
    };

    const sql = `INSERT INTO chats (id, title, createdAt, lastModified, model, systemPrompt, ollamaOptions) VALUES (?, ?, ?, ?, ?, ?, ?)`;
    const params = [
        newChatData.id,
        newChatData.title,
        newChatData.createdAt,
        newChatData.lastModified,
        newChatData.model,
        newChatData.systemPrompt,
        newChatData.ollamaOptions
    ];

    db.run(sql, params, function (err) {
        if (err) {
            console.error("!!! Error creating new chat:", err.message);
            res.status(500).json({ "error": "Failed to create chat", "details": err.message });
            return;
        }
        console.log(`Successfully created new chat with ID: ${newChatData.id}, Model: ${newChatData.model}`);
        const createdChatForResponse = {
            ...newChatData,
            ollamaOptions: ollamaOptions,
            messages: [],
        };
        res.status(201).json({ chat: createdChatForResponse });
    });
});

// POST /api/chats/:chatId/messages
app.post('/api/chats/:chatId/messages', (req, res) => {
    const chatId = req.params.chatId;
    const { messages: messagesToAdd } = req.body;
    console.log(`Received POST /api/chats/${chatId}/messages request`);

    if (!Array.isArray(messagesToAdd) || messagesToAdd.length === 0) {
        return res.status(400).json({ "error": "Request body must contain a non-empty 'messages' array" });
    }

    const nowISO = new Date().toISOString();
    const sqlInsertMsg = `INSERT INTO messages (id, chatId, role, content, timestamp) VALUES (?, ?, ?, ?, ?)`;
    const sqlUpdateChat = `UPDATE chats SET lastModified = ? WHERE id = ?`;
    const updateChatParams = [nowISO, chatId];

    db.serialize(() => {
        db.run("BEGIN TRANSACTION;");
        let errorOccurred = null;

        messagesToAdd.forEach(msg => {
            if (errorOccurred) return;
            if (!msg || typeof msg.content !== 'string' || !['user', 'assistant', 'system', 'tool'].includes(msg.role)) {
                errorOccurred = new Error(`Invalid message format or role: ${JSON.stringify(msg)}`);
                console.error(errorOccurred.message);
                return;
            }
            const messageId = uuidv4();
            const params = [messageId, chatId, msg.role, msg.content, nowISO];
            db.run(sqlInsertMsg, params, function (err) {
                if (err) { errorOccurred = err; console.error("!!! Error inserting message:", err.message); }
                else { console.log(`Inserted message ${messageId} (role: ${msg.role}) for chat ${chatId}`); }
            });
        });

        if (!errorOccurred) {
            db.run(sqlUpdateChat, updateChatParams, function (err) {
                if (err) { errorOccurred = err; console.error("!!! Error updating chat lastModified:", err.message); }
                else if (this.changes === 0) { errorOccurred = new Error(`Chat ${chatId} not found during update.`); console.error(errorOccurred.message); }
                else { console.log(`Updated lastModified for chat ${chatId}`); }
            });
        }

        if (errorOccurred) {
            db.run("ROLLBACK;");
            console.error("Transaction rolled back:", errorOccurred.message);
            const statusCode = errorOccurred.message.includes("not found") ? 404 : (errorOccurred.message.includes("Invalid message") ? 400 : 500);
            res.status(statusCode).json({ "error": "Failed to add messages", "details": errorOccurred.message });
        } else {
            db.run("COMMIT;");
            console.log(`Transaction committed successfully for chat ${chatId}`);
             res.status(201).json({ message: "Messages added successfully"});
        }
    });
});

// DELETE /api/chats/:chatId
app.delete('/api/chats/:chatId', (req, res) => {
    const chatId = req.params.chatId;
    console.log(`Received DELETE /api/chats/${chatId} request`);
    const sql = 'DELETE FROM chats WHERE id = ?';
    db.run(sql, [chatId], function (err) {
        if (err) {
            console.error(`!!! Error deleting chat ${chatId}:`, err.message);
            res.status(500).json({ "error": "Failed to delete chat", "details": err.message });
            return;
        }
        if (this.changes === 0) {
            console.log(`Chat ${chatId} not found for deletion`);
            res.status(404).json({ "error": "Chat not found" });
            return;
        }
        console.log(`Successfully deleted chat ${chatId} and associated messages (via CASCADE), changes: ${this.changes}`);
        res.status(204).send(); 
    });
});


// PATCH /api/chats/:chatId
app.patch('/api/chats/:chatId', (req, res) => {
    const chatId = req.params.chatId;
    const { title, systemPrompt, ollamaOptions } = req.body;

    console.log(`Received PATCH /api/chats/${chatId} request with data:`, req.body);

    const fieldsToUpdate = {};
    const params = [];
    const setClauses = [];

    const nowISO = new Date().toISOString();
    setClauses.push("lastModified = ?");
    params.push(nowISO);

    if (title !== undefined) {
        if (typeof title === 'string') {
            fieldsToUpdate.title = title.trim();
            setClauses.push("title = ?");
            params.push(fieldsToUpdate.title);
        } else {
             return res.status(400).json({ "error": "Invalid 'title' format, must be a string" });
        }
    }

    if (systemPrompt !== undefined) {
         if (typeof systemPrompt === 'string' || systemPrompt === null) {
             fieldsToUpdate.systemPrompt = systemPrompt ? systemPrompt.trim() : null;
             setClauses.push("systemPrompt = ?");
             params.push(fieldsToUpdate.systemPrompt);
         } else {
              return res.status(400).json({ "error": "Invalid 'systemPrompt' format, must be a string or null" });
         }
    }

    if (ollamaOptions !== undefined) {
         if (typeof ollamaOptions === 'object' || ollamaOptions === null) {
            try {
                fieldsToUpdate.ollamaOptions = ollamaOptions ? JSON.stringify(ollamaOptions) : null;
                setClauses.push("ollamaOptions = ?");
                params.push(fieldsToUpdate.ollamaOptions);
            } catch (stringifyError) {
                 console.error("!!! Error stringifying ollamaOptions:", stringifyError.message);
                 return res.status(400).json({ "error": "Invalid 'ollamaOptions' format" });
            }
         } else {
             return res.status(400).json({ "error": "Invalid 'ollamaOptions' format, must be an object or null" });
         }
    }


    if (setClauses.length <= 1) {
        console.warn(`PATCH /api/chats/${chatId}: No valid fields provided for update.`);
         return res.status(400).json({ "error": "No valid fields provided for update (e.g., title, systemPrompt, ollamaOptions)" });
    }

    params.push(chatId);

    const sql = `UPDATE chats SET ${setClauses.join(', ')} WHERE id = ?`;

    console.log(`Executing SQL: ${sql} with params:`, params);

    db.run(sql, params, function (err) {
        if (err) {
            console.error(`!!! Error updating chat ${chatId}:`, err.message);
            res.status(500).json({ "error": "Failed to update chat", "details": err.message });
            return;
        }
        if (this.changes === 0) {
            console.log(`Chat ${chatId} not found for update`);
            res.status(404).json({ "error": "Chat not found" });
            return;
        }

        console.log(`Successfully updated chat ${chatId}, changes: ${this.changes}`);
        res.status(200).json({ message: "Chat updated successfully" });
    });
});


app.listen(port, () => {
    console.log(`Backend server running on http://localhost:${port}`);
});

process.on('SIGINT', () => {
    console.log('SIGINT signal received: closing DB connection.');
    db.close((err) => {
        if (err) {
            console.error("!!! Error closing the database connection:", err.message);
            process.exit(1);
        }
        console.log('Closed the database connection successfully.');
        process.exit(0);
    });
});