import React, { useState, useRef, useEffect } from 'react';
import styles from './App.module.css';

import { Message as AppMessage } from './types/index';


interface OllamaMessage {
    role: 'user' | 'assistant' | 'system';
    content: string;
    // images?: string[];
}

interface OllamaChatResponse {
    model: string;
    created_at: string;
    message: OllamaMessage;
    done: boolean;
    total_duration?: number;
    load_duration?: number;
    prompt_eval_count?: number;
    prompt_eval_duration?: number;
    eval_count?: number;
    eval_duration?: number;
}

interface OllamaTagModel {
    name: string;
    model: string;
    modified_at: string;
    size: number;
    digest: string;
    details: { /* ... */ };
}

interface OllamaTagsResponse {
    models: OllamaTagModel[];
}

import Sidebar from './components/Sidebar/Sidebar';
import Header from './components/Header/Header';
import WelcomeMessage from './components/WelcomeMessage/WelcomeMessage';
import MessageList from './components/MessageList/MessageList';
import ChatInput from './components/ChatInput/ChatInput';
import SettingsModal from './components/SettingsModal/SettingsModal';
import SettingsIcon from './assets/settings-button.svg?react';

function App() {
    // ---  UI ---
    const [messages, setMessages] = useState<AppMessage[]>([]); 
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [isSidebarOpen, setIsSidebarOpen] = useState<boolean>(false);
    const [isSettingsModalOpen, setIsSettingsModalOpen] = useState<boolean>(false);

    // --- Models ---
    const [availableModels, setAvailableModels] = useState<OllamaTagModel[]>([]);
    const [selectedModel, setSelectedModel] = useState<string>('phi3:mini'); // example
    const [modelsLoading, setModelsLoading] = useState<boolean>(false);
    const [modelsError, setModelsError] = useState<string | null>(null);

    // --- Refs ---
    const inputRef = useRef<HTMLInputElement>(null);

    // --- Model download function ---
    const fetchModels = async () => {
        setModelsLoading(true);
        setModelsError(null);
        try {
            const response = await fetch('/ollama-api/tags'); 
            if (!response.ok) {
                let errorMsg = `HTTP error ${response.status}: ${response.statusText}`;
                try { 
                    const errBody = await response.json();
                    errorMsg += ` - ${errBody.error || JSON.stringify(errBody)}`;
                } catch (e) {/* ... */ }
                throw new Error(errorMsg);
            }
            const data: OllamaTagsResponse = await response.json();
            const models = data.models || [];
            setAvailableModels(models);
            const currentModelExists = models.some(m => m.name === selectedModel);
            if ((!currentModelExists || !selectedModel) && models.length > 0) {
                 console.log(`Selected model "${selectedModel}" not found or empty. Setting default to "${models[0].name}"`);
                 setSelectedModel(models[0].name);
            } else if (models.length === 0) {
                 console.warn("No local Ollama models found.");
                 setSelectedModel(''); 
            }

        } catch (error) {
            console.error("Error fetching Ollama models:", error);
            setModelsError(error instanceof Error ? error.message : "Unknown error fetching models");
            setAvailableModels([]);
            setSelectedModel(''); 
        } finally {
            setModelsLoading(false);
        }
    };


    useEffect(() => {
        inputRef.current?.focus();
    }, []);

    // Loading models at the first render
    useEffect(() => {
        fetchModels();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []); 

    const openSettingsModal = () => setIsSettingsModalOpen(true);
    const closeSettingsModal = () => setIsSettingsModalOpen(false);
    const toggleSidebar = () => setIsSidebarOpen(!isSidebarOpen);

    // --- Message sending function  ---
    const handleSendMessage = async (userMessageText: string) => {
        if (!userMessageText.trim() || isLoading || !selectedModel) {
            if (!selectedModel && !isLoading) { 
                console.error("No model selected!");
                setMessages((prev) => [...prev, { sender: 'ai', text: 'Error: No AI model selected. Please choose one in Settings.' }]);
            }
            return;
        }

        const newUserMessage: AppMessage = { sender: 'user', text: userMessageText.trim() };
        setMessages((prevMessages) => [...prevMessages, newUserMessage]);
        setIsLoading(true);

        const formatMessagesForApi = (currentMessages: AppMessage[]): OllamaMessage[] =>
            currentMessages.map((msg): OllamaMessage => ({
                role: msg.sender === 'user' ? 'user' : 'assistant',
                content: msg.text,
            }));

        const messagesToSend = formatMessagesForApi([...messages, newUserMessage]);

        try {
            const response = await fetch('/ollama-api/chat', { 
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    model: selectedModel, 
                    messages: messagesToSend,
                    stream: false,
                }),
            });

            if (!response.ok) {
                let errorText = `HTTP error! status: ${response.status} ${response.statusText}`;
                try {
                    const errorData = await response.json();
                    errorText += ` - ${errorData.error || JSON.stringify(errorData)}`;
                } catch (e) { /* ... */ }
                throw new Error(errorText);
            }

            const data: OllamaChatResponse = await response.json();

            if (data.message?.content) {
                const aiResponse: AppMessage = {
                    sender: 'ai',
                    text: data.message.content.trim(),
                };
                setMessages((prevMessages) => [...prevMessages, aiResponse]);
            } else {
                console.warn("Received empty or invalid response from AI:", data);
                const errorResponse: AppMessage = {
                    sender: 'ai',
                    text: "Sorry, I couldn't generate a response.",
                };
                setMessages((prevMessages) => [...prevMessages, errorResponse]);
            }

        } catch (error) {
            console.error("Error calling Ollama /api/chat:", error);
            const errorMessageText = error instanceof Error ? error.message : "Unknown error contacting AI";
            const errorResponse: AppMessage = { sender: 'ai', text: `Error: ${errorMessageText}` };
            setMessages((prevMessages) => [...prevMessages, errorResponse]);
        } finally {
            setIsLoading(false);
            inputRef.current?.focus();
        }
    };

    return (
        <div className={`${styles.chatAppContainer} ${isSidebarOpen ? styles.sidebarOpen : ''}`}>
            <Sidebar isOpen={isSidebarOpen} onClose={toggleSidebar} />
            <Header onToggleSidebar={toggleSidebar} isSidebarOpen={isSidebarOpen} />
            <div className={styles.mainContent}>
                {messages.length === 0 && !isLoading && <WelcomeMessage />}
                {messages.length > 0 && <MessageList messages={messages} />}
                {isLoading && messages.length === 0 && (
                    <div className={styles.centralLoading}>
                        AI...
                        <div className={styles.spinner}></div>
                    </div>
                )}
            </div>


            <ChatInput
                onSendMessage={handleSendMessage}
                isLoading={isLoading}
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