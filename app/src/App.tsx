import { useState, useRef, useEffect, useMemo } from 'react';
import styles from './App.module.css';

import {
    Message as AppMessage, 
    OllamaTagModel, 
    Chat,            
    BackendChatResponse,  
    BackendChatsResponse, 
    OllamaMessage, 
    OllamaChatResponse,
    OllamaTagsResponse
} from './types';

import Sidebar from './components/Sidebar/Sidebar';
import Header from './components/Header/Header';
import WelcomeMessage from './components/WelcomeMessage/WelcomeMessage';
import MessageList from './components/MessageList/MessageList';
import ChatInput from './components/ChatInput/ChatInput';
import SettingsModal from './components/SettingsModal/SettingsModal';

import SettingsIcon from './assets/settings-button.svg?react';


const DEFAULT_FAVICON = '/logo.svg'; 
const ACTIVE_CHAT_FAVICON = '/logo.svg'; 
const DEFAULT_TITLE = 'Ollama Chat'; 

function App() {

    const [savedChats, setSavedChats] = useState<Chat[]>([]);
    const [activeChatId, setActiveChatId] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [isChatLoading, setIsChatLoading] = useState<boolean>(false); 
    const [isAppLoading, setIsAppLoading] = useState<boolean>(true); 
    const [isSidebarOpen, setIsSidebarOpen] = useState<boolean>(false);
    const [isSettingsModalOpen, setIsSettingsModalOpen] = useState<boolean>(false);
    
    const [availableModels, setAvailableModels] = useState<OllamaTagModel[]>([]);
    const [selectedModel, setSelectedModel] = useState<string>('phi3:mini'); 
    const [modelsLoading, setModelsLoading] = useState<boolean>(false);
    const [modelsError, setModelsError] = useState<string | null>(null);
    
    const inputRef = useRef<HTMLInputElement>(null);
    const loadInitialChats = async () => {
        console.log("Attempting to load initial chats...");
        try {
            const response = await fetch('/backend-api/chats');
            if (!response.ok) throw new Error(`Failed to fetch chats list: ${response.statusText}`);
            const data: BackendChatsResponse = await response.json();

            const chatsFromBackend: Chat[] = (data.chats || []).map(c => ({
                id: c.id,            
                title: c.title,      
                messages: [], 
                createdAt: new Date(c.createdAt), 
                lastModified: new Date(c.lastModified),
                model: c.model 
            }));


            const sortedChats = chatsFromBackend.sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime());
            setSavedChats(sortedChats);
            console.log("Loaded chats:", sortedChats.length);
            setActiveChatId(sortedChats.length > 0 ? sortedChats[0].id : null);

        } catch (error) {
            console.error("Error loading initial chats:", error);
            setActiveChatId(null);
            setSavedChats([]);
        }
    };
    
    const fetchModels = async () => {
        console.log("Attempting to fetch Ollama models...");
        setModelsLoading(true);
        setModelsError(null);
        try {
            const response = await fetch('/ollama-api/tags'); 
            if (!response.ok) {
                 let errorMsg = `HTTP error ${response.status}: ${response.statusText}`;
                 try { const errBody = await response.json(); errorMsg += ` - ${errBody.error || JSON.stringify(errBody)}`;} catch (e) {/* ignore */}
                 throw new Error(errorMsg);
            }
            const data: OllamaTagsResponse = await response.json();
            const models = data.models || [];
            setAvailableModels(models);
            console.log("Fetched Ollama models:", models.length);
    
            const currentModelExists = models.some(m => m.name === selectedModel);
            if ((!currentModelExists || !selectedModel) && models.length > 0) {
                console.log(`Selected model "${selectedModel}" not found/empty. Setting default: "${models[0].name}"`);
                setSelectedModel(models[0].name);
            } else if (models.length === 0) {
                console.warn("No local Ollama models found.");
                setSelectedModel('');
            }
        } catch (error) {
            console.error("Error fetching Ollama models:", error);
            setModelsError(error instanceof Error ? error.message : "Unknown error");
            setAvailableModels([]);
            setSelectedModel('');
        } finally {
            setModelsLoading(false);
        }
    };
    
    
    const loadMessagesForChat = async (chatId: string | null) => {
        if (!chatId) return;
        const chatState = savedChats.find(c => c.id === chatId);
        if ((chatState?.messages ?? []).length > 0) {
            console.log("Messages for chat", chatId, "already loaded.");
            return;
        }
    
        console.log("Loading messages for chat:", chatId);
        setIsChatLoading(true); 
        try {
            const response = await fetch(`/backend-api/chats/${chatId}`);
            if (!response.ok) {
                 if (response.status === 404) {
                      console.warn(`Chat ${chatId} not found on backend. Removing from list.`);
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
                    ...chat,
                    messages: data.chat.messages || [],
                    title: data.chat.title,
                    createdAt: new Date(data.chat.createdAt),
                    lastModified: new Date(data.chat.lastModified),
                    model: data.chat.model
                 } : chat
            ));
        } catch (error) {
            console.error("Error loading messages:", error);
        } finally {
            setIsChatLoading(false);
        }
    };
    
    useEffect(() => {
        setIsAppLoading(true);
        Promise.all([loadInitialChats(), fetchModels()])
            .catch(err => console.error("Error during initial data load:", err))
            .finally(() => setIsAppLoading(false));
    }, []); 
    
    useEffect(() => {
        loadMessagesForChat(activeChatId);
    
    }, [activeChatId]);
    
    useEffect(() => {
    
        setTimeout(() => inputRef.current?.focus(), 100);
    }, [activeChatId]); 

    useEffect(() => {
        const faviconElement = document.getElementById('favicon') as HTMLLinkElement | null;

        const activeChat = savedChats.find(chat => chat.id === activeChatId);
    
        if (activeChat) {

            document.title = activeChat.title || 'Chat'; 
            if (faviconElement) {
                faviconElement.href = ACTIVE_CHAT_FAVICON;
            }
        } else {
            document.title = DEFAULT_TITLE;
            if (faviconElement) {
                faviconElement.href = DEFAULT_FAVICON;
            }
        }
    
    }, [activeChatId, savedChats]);
    
    const openSettingsModal = () => setIsSettingsModalOpen(true);
    const closeSettingsModal = () => setIsSettingsModalOpen(false);
    const toggleSidebar = () => setIsSidebarOpen(!isSidebarOpen);
    
    const handleCreateNewChat = async (): Promise<string | null> => {
        console.log("handleCreateNewChat called");
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
    
    const handleDeleteChat = async (chatIdToDelete: string) => {
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
    
    const handleEditChat = async (chatIdToEdit: string) => {
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
    
    const handleSendMessage = async (userMessageText: string) => {
        const trimmedMessage = userMessageText.trim();
        if (!trimmedMessage || isLoading) return; 
    
        let currentChatId = activeChatId;
        let isNewChat = false;
    
        if (!currentChatId) {
            console.log("No active chat, creating new one...");
            setIsChatLoading(true);
            const newChatId = await handleCreateNewChat();
            setIsChatLoading(false);
            if (!newChatId) return; 
            currentChatId = newChatId;
            isNewChat = true;
        }
    
        if (!selectedModel) {
            console.error("No model selected!");
            const noModelErrorMsg: AppMessage = { sender: 'assistant', text: 'Error: No AI model selected. Please select one in settings.' };
            setSavedChats(prev => prev.map(c => c.id === currentChatId ? {...c, messages: [...(c.messages || []), noModelErrorMsg ]} : c)
                .sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime()));
            return;
        }
    
        const newUserMessage: AppMessage = { sender: 'user', text: trimmedMessage };
        const chatBeforeUpdate = savedChats.find(c => c.id === currentChatId);
        const previousMessages = chatBeforeUpdate?.messages || [];
    
        const messagesForOllamaApi: OllamaMessage[] = [
            ...previousMessages.map(msg => ({
                role: msg.sender === 'assistant' ? 'assistant' : 'user' as ('user' | 'assistant'),
                content: msg.text
            })),
            { role: 'user', content: newUserMessage.text } 
        ];
    
        if (messagesForOllamaApi.length === 0) {
             console.error("Logic Error: History for Ollama is unexpectedly empty even after adding user message.");
             return;
        }
    
        setIsLoading(true);
        setSavedChats(prevChats => prevChats
            .map(chat =>
                chat.id === currentChatId
                    ? {
                        ...chat,
                        title: isNewChat ? trimmedMessage.substring(0, 30) + (trimmedMessage.length > 30 ? '...' : '') : chat.title,
                        messages: [...previousMessages, newUserMessage], 
                        lastModified: new Date() 
                      }
                    : chat
            )
             .sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime()) 
        );
    
        let aiResponseText: string | null = null;
        let errorMessageText: string | null = null;
    
        try {
             console.log(`Sending to Ollama (Model: ${selectedModel}):`, messagesForOllamaApi.length, "messages"); 
             const ollamaResponse = await fetch('/ollama-api/chat', {
                 method: 'POST',
                 headers: { 'Content-Type': 'application/json' },
                 body: JSON.stringify({ model: selectedModel, messages: messagesForOllamaApi, stream: false }), 
             });
    
             if (!ollamaResponse.ok) {
                 let errorText = `Ollama API error! Status: ${ollamaResponse.status}`;
                 try { const errorBody = await ollamaResponse.json(); errorText += ` - ${errorBody.error || JSON.stringify(errorBody)}`; } catch (e) { errorText += ` - ${ollamaResponse.statusText}`; }
                 throw new Error(errorText);
              }
    
             const ollamaData: OllamaChatResponse = await ollamaResponse.json();
             if (ollamaData.message?.content) {
                 aiResponseText = ollamaData.message.content.trim();
                 if (aiResponseText) {
                     console.log("Ollama Response OK:", aiResponseText.substring(0, 100) + "...");
                 } else {
                     console.warn("Ollama Response is null or empty.");
                 }
             } else {
                 console.warn("Ollama response format unexpected:", ollamaData);
                 throw new Error("assistant returned empty or invalid response format.");
             }
        } catch (error) {
             console.error("Error calling Ollama /api/chat:", error);
             errorMessageText = error instanceof Error ? error.message : "Unknown error contacting assistant";
        } finally {
             setIsLoading(false);
             inputRef.current?.focus();
        }
    
        const aiResponseMessage: AppMessage | null = aiResponseText ? { sender: 'assistant', text: aiResponseText } : null;
        const errorResponseMessageForUI: AppMessage | null = errorMessageText ? { sender: 'assistant', text: `Error: ${errorMessageText}` } : null;
        const messageToAddToUi = aiResponseMessage || errorResponseMessageForUI;
    
        if (messageToAddToUi) {
            setSavedChats(prevChats => prevChats.map(chat => {
                if (chat.id === currentChatId) {
                     const currentMessages = chat.messages || [];
                     return { ...chat, messages: [...currentMessages, messageToAddToUi], lastModified: new Date() };
                }
                return chat;
            }).sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime())); 
        }
        if (aiResponseMessage) {
            try {
                 const backendPayload = {
                     messages: [
                         { role: 'user', content: newUserMessage.text },
                         { role: 'assistant', content: aiResponseMessage.text }
                     ]
                 };
    
                 console.log("Attempting to save messages to backend for chat:", currentChatId);
                 const backendResponse = await fetch(`/backend-api/chats/${currentChatId}/messages`, {
                     method: 'POST',
                     headers: { 'Content-Type': 'application/json' },
                     body: JSON.stringify(backendPayload)
                 });
    
                 if (!backendResponse.ok) {
                     const errText = await backendResponse.text();
                     throw new Error(`Backend error (${backendResponse.status}): ${errText}`);
                 }
                 console.log("Messages successfully saved to backend for chat:", currentChatId);
    
             } catch (error) {
                console.error("Error saving messages to backend:", error);
                const saveErrorMessageForUi: AppMessage = {
                   sender: 'assistant', 
                   text: `⚠️ Error saving chat: ${error instanceof Error ? error.message : 'Unknown error'}`
                };
                setSavedChats(prevChats => prevChats.map(chat =>
                   chat.id === currentChatId ? {...chat, messages: [...(chat.messages || []), saveErrorMessageForUi ]} : chat
                ).sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime()));
            }
        }
    }; 
    
    const currentChatMessages = useMemo(() => {
        if (!activeChatId) return [];
        return savedChats.find(chat => chat.id === activeChatId)?.messages || [];
    }, [activeChatId, savedChats]);
    
    if (isAppLoading) {
        return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', fontSize: '1.5em', color: '#ccc', background: '#1a1a1a' }}>Loading Application...</div>;
    }
    
    

    
    
    
    return (
        <div className={`${styles.chatAppContainer} ${isSidebarOpen ? styles.sidebarOpen : ''}`}>
    
            <Sidebar
                isOpen={isSidebarOpen}
                savedChats={savedChats.map(c => ({ id: c.id, title: c.title }))} 
                activeChatId={activeChatId}
                onSelectChat={setActiveChatId}
                onDeleteChat={handleDeleteChat}
                onEditChat={handleEditChat} 
            />
    
            <Header
                onToggleSidebar={toggleSidebar}
                isSidebarOpen={isSidebarOpen}
                onCreateNewChat={handleCreateNewChat}
            />
    
            <div className={styles.mainContent}>
                {(!activeChatId || (activeChatId && currentChatMessages.length === 0)) && !isChatLoading && !isLoading && <WelcomeMessage />}
                {isChatLoading && currentChatMessages.length === 0 && (
                    <div className={styles.centralLoading}>Loading chat history...<div className={styles.spinner}></div></div>
                )}
                 {isLoading && currentChatMessages.length === 0 && !isChatLoading && (
                      <div className={styles.centralLoading}>AI is thinking...<div className={styles.spinner}></div></div>
                 )}
    
                {currentChatMessages.length > 0 && <MessageList messages={currentChatMessages} />}
    
            </div>
            <ChatInput
                onSendMessage={handleSendMessage}
                isLoading={isLoading || isChatLoading} 
                inputRef={inputRef}
            />
            <button className={styles.bottomRightSettingsButton} onClick={openSettingsModal} aria-label="Open Settings">
                <SettingsIcon />
            </button>
            <SettingsModal
                isOpen={isSettingsModalOpen}
                onClose={closeSettingsModal}
                availableModels={availableModels}
                selectedModel={selectedModel}
                onSelectModel={setSelectedModel}
                modelsLoading={modelsLoading}
                modelsError={modelsError}
                onRefreshModels={fetchModels}
            />
        </div>
    );
}

export default App;