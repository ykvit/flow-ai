import { useState } from 'react';
import { Chat, Message, OllamaMessage, OllamaChatResponse } from '../types'; 

export function useChat() {
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [abortController, setAbortController] = useState<AbortController | null>(null);

    const handleSendMessage = async (
        userMessageText: string, 
        activeChatId: string | null,
        createNewChat: (selectedModel: string) => Promise<string | null>,
        selectedModel: string,
        savedChats: Chat[], // Змінено з any[] на Chat[]
        setSavedChats: React.Dispatch<React.SetStateAction<Chat[]>>, 
        inputRef: React.RefObject<HTMLTextAreaElement | null>
    ) => {
        const trimmedMessage = userMessageText.trim();
        if (!trimmedMessage || isLoading) return;

        let currentChatId = activeChatId;
        let isNewChat = false;

        if (!currentChatId) {
            const newChatId = await createNewChat(selectedModel);
            if (!newChatId) return;
            currentChatId = newChatId;
            isNewChat = true;
        }
        
        if (!currentChatId) {
            console.error("Failed to set currentChatId even after attempting to create a new chat.");
            return; 
        }


        if (!selectedModel) {
            const noModelErrorMsg: Message = { sender: 'assistant', text: 'Error: No AI model selected. Please select one in settings.' };
            if (currentChatId) {
                setSavedChats(prev => prev.map(c => c.id === currentChatId ? {...c, messages: [...(c.messages || []), noModelErrorMsg ]} : c).sort((a,b) => b.lastModified.getTime() - a.lastModified.getTime()));
            }
            return;
        }
    
        const newUserMessage: Message = { sender: 'user', text: trimmedMessage };
        const chatBeforeUpdate = savedChats.find(c => c.id === currentChatId);
        const previousMessages: Message[] = chatBeforeUpdate?.messages || []; 

        const messagesForOllamaApi: OllamaMessage[] = [
            ...previousMessages.map((msg: Message) => ({ 
                role: msg.sender === 'assistant' ? 'assistant' : 'user' as ('user' | 'assistant'),
                content: msg.text
            })),
            { role: 'user', content: newUserMessage.text }
        ];
    
        setIsLoading(true);
        if (currentChatId) {
            setSavedChats(prevChats => prevChats
                .map(chat =>
                    chat.id === currentChatId
                        ? {
                            ...chat,
                            title: isNewChat && messagesForOllamaApi.length === 1 ?
                                   trimmedMessage.substring(0, 40) + (trimmedMessage.length > 40 ? '...' : '') : chat.title,
                            messages: [...previousMessages, newUserMessage],
                            lastModified: new Date() 
                          }
                        : chat
                )
                .sort((a, b) => new Date(b.lastModified).getTime() - new Date(a.lastModified).getTime())
            );
        }
    
        let aiResponseText: string | null = null;
        let errorMessageText: string | null = null;
        let wasAborted = false;

        const controller = new AbortController();
        setAbortController(controller);
    
        try {
            const ollamaResponse = await fetch('/ollama-api/chat', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ model: selectedModel, messages: messagesForOllamaApi, stream: false }),
                signal: controller.signal,
            });

            if (!ollamaResponse.ok) {
                let errorText = `Ollama API error! Status: ${ollamaResponse.status}`;
                try { 
                    const errorBody = await ollamaResponse.json(); 
                    errorText += ` - ${errorBody.error || JSON.stringify(errorBody)}`; 
                } catch (e) { 
                    errorText += ` - ${ollamaResponse.statusText}`; 
                }
                throw new Error(errorText);
            }

            const ollamaData: OllamaChatResponse = await ollamaResponse.json();
            if (ollamaData.message?.content) {
                aiResponseText = ollamaData.message.content.trim();
            } else {
                throw new Error("Assistant returned empty or invalid response format.");
            }
        } catch (error) {
            if (error instanceof Error && error.name === 'AbortError') {
                errorMessageText = "Generation stopped by user.";
                wasAborted = true;
            } else {
                errorMessageText = error instanceof Error ? error.message : "Unknown error contacting assistant";
            }
        } finally {
            setIsLoading(false);
            setAbortController(null); 
            inputRef.current?.focus();
        }

        let finalAssistantMessage: Message | null = null;
        if (aiResponseText) {
            finalAssistantMessage = { sender: 'assistant', text: aiResponseText };
        } else if (errorMessageText) {
            finalAssistantMessage = { sender: 'assistant', text: wasAborted ? errorMessageText : `Error: ${errorMessageText}` };
        }

        if (finalAssistantMessage && currentChatId) { 
            setSavedChats(prevChats => prevChats.map(chat => {
                if (chat.id === currentChatId) {
                    return { ...chat, messages: [...(chat.messages || []), finalAssistantMessage!], lastModified: new Date() };
                }
                return chat;
            }).sort((a, b) => new Date(b.lastModified).getTime() - new Date(a.lastModified).getTime()));
        }

        if (aiResponseText && !wasAborted && currentChatId) { 
            try {
                const backendPayload = {
                    messages: [ { role: 'user', content: newUserMessage.text }, { role: 'assistant', content: aiResponseText } ]
                };
                const backendResponse = await fetch(`/backend-api/chats/${currentChatId}/messages`, {
                    method: 'POST', headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(backendPayload)
                });
                if (!backendResponse.ok) {
                    const errText = await backendResponse.text();
                    throw new Error(`Backend error (${backendResponse.status}): ${errText}`);
                }
            } catch (error) {
                console.error("Error saving messages to backend:", error);
                const saveErrorMessageForUi: Message = {
                   sender: 'assistant',
                   text: `⚠️ Error saving chat: ${error instanceof Error ? error.message : 'Unknown error'}`
                };
                if (currentChatId) {
                    setSavedChats(prevChats => prevChats.map(chat =>
                       chat.id === currentChatId ? {...chat, messages: [...(chat.messages || []), saveErrorMessageForUi ]} : chat
                    ).sort((a, b) => new Date(b.lastModified).getTime() - new Date(a.lastModified).getTime()));
                }
            }
        }
    };

    const stopGenerating = () => {
        if (abortController) {
            console.log("Attempting to stop generation via button...");
            abortController.abort();
        }
    };

    return {
        isLoading,
        handleSendMessage,
        stopGenerating
    };
}