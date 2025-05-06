import { useState, useEffect } from 'react';
import { Chat, BackendChatResponse, BackendChatsResponse } from '../types';

export function useChats() {
    const [savedChats, setSavedChats] = useState<Chat[]>([]);
    const [activeChatId, setActiveChatId] = useState<string | null>(null);
    const [isChatLoading, setIsChatLoading] = useState<boolean>(false);
    
    const loadInitialChats = async () => {
        console.log("Attempting to load initial chats...");
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
            setActiveChatId(sortedChats.length > 0 ? sortedChats[0].id : null);
        } catch (error) {
            console.error("Error loading initial chats:", error);
            setActiveChatId(null); setSavedChats([]);
        }
    };
    
    const loadMessagesForChat = async (chatId: string | null) => {
        if (!chatId) return;
        const chatState = savedChats.find(c => c.id === chatId);
        if (((chatState?.messages ?? []).length > 0) || (isChatLoading && activeChatId === chatId)) {
            return;
        }
        console.log("Loading messages for chat:", chatId);
        setIsChatLoading(true);
        try {
            const response = await fetch(`/backend-api/chats/${chatId}`);
            if (!response.ok) {
                if (response.status === 404) {
                    setSavedChats(prev => prev.filter(c => c.id !== chatId));
                    if (activeChatId === chatId) setActiveChatId(null);
                } else {
                    throw new Error(`Failed to fetch messages for chat ${chatId}: ${response.statusText}`);
                }
                return;
            }
            const data: BackendChatResponse = await response.json();
            setSavedChats(prevChats => prevChats.map(chat =>
                chat.id === chatId ? {
                    ...chat, messages: data.chat.messages || [], title: data.chat.title,
                    createdAt: new Date(data.chat.createdAt), lastModified: new Date(data.chat.lastModified),
                    model: data.chat.model
                } : chat
            ));
        } catch (error) {
            console.error("Error loading messages:", error);
        } finally {
            setIsChatLoading(false);
        }
    };
    
    const createNewChat = async (selectedModel: string): Promise<string | null> => {
        console.log("createNewChat called");
        try {
            const response = await fetch('/backend-api/chats', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ title: "New Chat", model: selectedModel || "unknown" }) 
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
            console.log("New chat created and set active:", newChat.id);
            return newChat.id;
        } catch (error) {
            console.error("Error creating new chat:", error);
            alert(`Error: Could not create a new chat. ${error instanceof Error ? error.message : ''}`);
            return null;
        }
    };
    
    const deleteChat = async (chatIdToDelete: string) => {
        if (!window.confirm("Are you sure you want to delete this chat history? This action cannot be undone.")) return;
    
        const originalChats = [...savedChats];
        const originalActiveId = activeChatId;
    
        const remainingChats = savedChats.filter(chat => chat.id !== chatIdToDelete);
        const sortedRemaining = remainingChats.sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime());
        setSavedChats(sortedRemaining);
    
        if (activeChatId === chatIdToDelete) {
            setActiveChatId(sortedRemaining.length > 0 ? sortedRemaining[0].id : null);
        }
    
        try {
            const response = await fetch(`/backend-api/chats/${chatIdToDelete}`, { method: 'DELETE' });
    
            if (!response.ok && response.status !== 404) {
                 throw new Error(`Failed to delete chat (status: ${response.status})`);
            }
            console.log(`Chat ${chatIdToDelete} deleted successfully (or was already gone).`);
        } catch (error) {
            console.error("Error deleting chat:", error);
            alert(`Error deleting chat: ${error instanceof Error ? error.message : ''}`);
    
            setSavedChats(originalChats);
            setActiveChatId(originalActiveId);
        }
    };
    
    const editChat = async (chatIdToEdit: string) => {
        console.log("Attempting to edit chat:", chatIdToEdit);
        const currentChat = savedChats.find(chat => chat.id === chatIdToEdit);
        const currentTitle = currentChat ? currentChat.title : "this chat";
    
        const newTitle = prompt(`Enter new title for "${currentTitle}":`, currentChat?.title || "");
    
        if (newTitle === null || newTitle.trim() === "") {
            console.log("Edit cancelled or new title empty.");
            return;
        }
        const trimmedNewTitle = newTitle.trim();
    
        const originalChats = [...savedChats];
    
        setSavedChats(prevChats =>
            prevChats.map(chat =>
                chat.id === chatIdToEdit
                    ? { ...chat, title: trimmedNewTitle, lastModified: new Date() } 
                    : chat
            ).sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime()) 
        );
    
        try {
            console.log(`Sending PUT request to /backend-api/chats/${chatIdToEdit} with title: ${trimmedNewTitle}`);
            const response = await fetch(`/backend-api/chats/${chatIdToEdit}`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ title: trimmedNewTitle })
            });
    
            if (!response.ok) {
                let errorText = `Failed to update chat title (status: ${response.status})`;
                try { const errBody = await response.json(); errorText += ` - ${errBody.error || JSON.stringify(errBody)}`; } catch (e) { errorText += ` - ${response.statusText}`; }
                throw new Error(errorText);
            }
            console.log(`Chat ${chatIdToEdit} title updated successfully on backend.`);
        } catch (error) {
            console.error("Error updating chat title:", error);
            alert(`Error updating chat title: ${error instanceof Error ? error.message : 'Unknown error'}`);
            setSavedChats(originalChats);
        }
    };

    useEffect(() => {
        if (activeChatId) {
            loadMessagesForChat(activeChatId);
        }
    }, [activeChatId]);

    return {
        savedChats,
        setSavedChats,
        activeChatId, 
        setActiveChatId,
        isChatLoading,
        loadInitialChats,
        loadMessagesForChat,
        createNewChat,
        deleteChat,
        editChat
    };
}