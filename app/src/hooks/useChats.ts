
import { useState, useEffect, useCallback } from 'react';
import { Chat, BackendChatResponse, BackendChatsResponse } from '../types';

export function useChats() {
    const [savedChats, setSavedChats] = useState<Chat[]>([]);
    const [activeChatId, setActiveChatId] = useState<string | null>(null);
    const [isChatLoading, setIsChatLoading] = useState<boolean>(false); 

    const updateChatTitle = useCallback(async (chatIdToUpdate: string, newTitle: string) => {
        if (!newTitle || newTitle.trim() === "") {
            console.warn("[updateChatTitle] Attempted to update with empty title. Skipping.");
            return;
        }
        const trimmedNewTitle = newTitle.trim();
        console.log(`[updateChatTitle] Attempting to update title for chat ${chatIdToUpdate} to: "${trimmedNewTitle}"`);


        setSavedChats(prevChats =>
            prevChats.map(chat =>
                chat.id === chatIdToUpdate
                    ? { ...chat, title: trimmedNewTitle, lastModified: new Date() } 
                    : chat
            ).sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime())
        );

        try {
            const response = await fetch(`/backend-api/chats/${chatIdToUpdate}`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ title: trimmedNewTitle })
            });

            if (!response.ok) {
                let errorText = `[updateChatTitle] Failed to update chat title on backend (status: ${response.status})`;
                try { const errBody = await response.json(); errorText += ` - ${errBody.error || JSON.stringify(errBody)}`; } catch (e) { /* noop */ }
                throw new Error(errorText);
            }
            console.log(`[updateChatTitle] Chat ${chatIdToUpdate} title updated successfully on backend to: "${trimmedNewTitle}"`);
        } catch (error) {
            console.error("[updateChatTitle] Error updating chat title:", error);
        }
    }, [setSavedChats]);


    const loadInitialChatsEffectHook = useCallback(async () => {
        console.log("Attempting to load initial chats for App.tsx effect hook...");
        try {
            const response = await fetch('/backend-api/chats');
            if (!response.ok) throw new Error(`Failed to fetch chats list: ${response.statusText}`);
            const data: BackendChatsResponse = await response.json();
            const chatsFromBackend: Chat[] = (data.chats || []).map(c => ({
                id: c.id, title: c.title, messages: [], 
                createdAt: new Date(c.createdAt), lastModified: new Date(c.lastModified), model: c.model
            }));
            const sortedChats = chatsFromBackend.sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime());
            setSavedChats(sortedChats);
        } catch (error) {
            console.error("Error loading initial chats:", error);
            setSavedChats([]);
        }
    }, [setSavedChats]);

    const loadMessagesForChat = useCallback(async (chatIdToLoad: string | null) => {
        if (!chatIdToLoad) return;

        console.log("Loading messages for chat (useCallback, stable deps):", chatIdToLoad);
        setIsChatLoading(true);
        try {
            const response = await fetch(`/backend-api/chats/${chatIdToLoad}`);
            if (!response.ok) {
                if (response.status === 404) {
                    setSavedChats(prev => prev.filter(c => c.id !== chatIdToLoad));
                    if (activeChatId === chatIdToLoad) {
                        setActiveChatId(null);
                    }
                } else {
                    throw new Error(`Failed to fetch messages for chat ${chatIdToLoad}: ${response.statusText}`);
                }
                return;
            }
            const data: BackendChatResponse = await response.json();
            setSavedChats(prevChats => prevChats.map(chat =>
                chat.id === chatIdToLoad ? {
                    ...chat,
                    messages: data.chat.messages || [],
                    title: data.chat.title,
                    createdAt: new Date(data.chat.createdAt),
                    lastModified: new Date(data.chat.lastModified),
                    model: data.chat.model
                } : chat
            ));
        } catch (error) {
            console.error(`Error loading messages for chat ${chatIdToLoad}:`, error);
        } finally {
            setIsChatLoading(false);
        }
    }, [activeChatId, setActiveChatId, setIsChatLoading, setSavedChats]);

    const createNewChat = useCallback(async (selectedModel: string, initialTitle?: string): Promise<string | null> => {
        console.log("createNewChat called with model:", selectedModel, "and initial title:", initialTitle);
        const titleForBackend = initialTitle || "New Chat";
        try {
            const response = await fetch('/backend-api/chats', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ title: titleForBackend, model: selectedModel || "unknown" })
            });
            if (!response.ok) {
                const errText = await response.text();
                throw new Error(`Failed to create new chat on backend: ${response.status} ${errText}`);
            }
            const data: BackendChatResponse = await response.json();
            const newChat: Chat = {
                id: data.chat.id,
                title: data.chat.title,
                messages: [], 
                createdAt: new Date(data.chat.createdAt),
                lastModified: new Date(data.chat.lastModified),
                model: data.chat.model
            };
            setSavedChats(prevChats => [newChat, ...prevChats].sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime()));
            setActiveChatId(newChat.id);
            console.log("New chat created and set active:", newChat.id, "with title:", newChat.title);
            return newChat.id;
        } catch (error) {
            console.error("Error creating new chat:", error);
            alert(`Error: Could not create a new chat. ${error instanceof Error ? error.message : ''}`);
            return null;
        }
    }, [setSavedChats, setActiveChatId]);

    const deleteChat = useCallback(async (chatIdToDelete: string) => {
        if (!window.confirm("Are you sure you want to delete this chat history? This action cannot be undone.")) return;


        const originalActiveId = activeChatId;
        setSavedChats(prevSavedChats => {
            const remainingChats = prevSavedChats.filter(chat => chat.id !== chatIdToDelete);
            const sortedRemaining = remainingChats.sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime());
            if (originalActiveId === chatIdToDelete) {
                setActiveChatId(sortedRemaining.length > 0 ? sortedRemaining[0].id : null);
            }
            return sortedRemaining;
        });

        try {
            const response = await fetch(`/backend-api/chats/${chatIdToDelete}`, { method: 'DELETE' });
            if (!response.ok && response.status !== 404) {
                 throw new Error(`Failed to delete chat (status: ${response.status})`);
            }
            console.log(`Chat ${chatIdToDelete} deleted successfully (or was already gone).`);
        } catch (error) {
            console.error("Error deleting chat:", error);
            alert(`Error deleting chat: ${error instanceof Error ? error.message : ''}`);
            loadInitialChatsEffectHook();
        }
    }, [activeChatId, setActiveChatId, setSavedChats, loadInitialChatsEffectHook]);

    const editChat = useCallback(async (chatIdToEdit: string) => {
        const currentChat = savedChats.find(chat => chat.id === chatIdToEdit);
        const currentTitle = currentChat ? currentChat.title : "this chat";
        const newTitle = prompt(`Enter new title for "${currentTitle}":`, currentChat?.title || "");

        if (newTitle === null || newTitle.trim() === "") return;

        const trimmedNewTitle = newTitle.trim();

        setSavedChats(prevChats =>
            prevChats.map(chat =>
                chat.id === chatIdToEdit
                    ? { ...chat, title: trimmedNewTitle, lastModified: new Date() }
                    : chat
            ).sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime())
        );

        try {
            const response = await fetch(`/backend-api/chats/${chatIdToEdit}`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ title: trimmedNewTitle })
            });
             if (!response.ok) {
                let errorText = `Failed to update chat title (status: ${response.status})`;
                try { const errBody = await response.json(); errorText += ` - ${errBody.error || JSON.stringify(errBody)}`; } catch (e) { /* noop */ }
                throw new Error(errorText);
            }
        } catch (error) {
            console.error("Error updating chat title:", error);
            alert(`Error updating chat title: ${error instanceof Error ? error.message : 'Unknown error'}`);
            loadInitialChatsEffectHook();
        }
    }, [savedChats, setSavedChats, loadInitialChatsEffectHook]);

    useEffect(() => {
        if (activeChatId) {
            console.log(`useEffect for activeChatId triggered: ${activeChatId}. Calling loadMessagesForChat.`);
            loadMessagesForChat(activeChatId);
        }
    }, [activeChatId, loadMessagesForChat]);

    return {
        savedChats,
        setSavedChats,
        activeChatId,
        setActiveChatId,
        isChatLoading,
        loadInitialChats: loadInitialChatsEffectHook,
        loadMessagesForChat,
        createNewChat,
        updateChatTitle, 
        deleteChat,
        editChat
    };
}