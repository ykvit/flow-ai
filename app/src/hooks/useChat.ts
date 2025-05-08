import { useState } from 'react';
import { Chat, Message, OllamaMessage, OllamaChatResponse } from '../types'; 
import { useNotifications } from './useNotifications'; 

const DEFAULT_SYSTEM_PROMPT = "You are Flow-ai. You are here to help, you have to be polite. If you don't know the answer, say you don't know.";

export function useChat() {
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [abortController, setAbortController] = useState<AbortController | null>(null);
    const { showNotification } = useNotifications();

    const handleSendMessage = async (
        userMessageText: string, 
        activeChatId: string | null,
        createNewChat: (selectedModel: string) => Promise<string | null>,
        selectedModel: string,
        savedChats: Chat[], 
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
            setSavedChats(prev => prev.map(c => c.id === currentChatId ? {...c, messages: [...(c.messages || []), noModelErrorMsg ]} : c).sort((a,b) => b.lastModified.getTime() - a.lastModified.getTime()));
            return;
        }
    
        const newUserMessage: Message = { sender: 'user', text: trimmedMessage };
        const chatBeforeUpdate = savedChats.find(c => c.id === currentChatId);
        const previousMessages: Message[] = chatBeforeUpdate?.messages || []; 
        const baseMessagesForApi: OllamaMessage[] = [
            ...previousMessages.map((msg: Message) => ({ 
                role: msg.sender === 'assistant' ? 'assistant' : 'user' as ('user' | 'assistant'),
                content: msg.text
            })),
            { role: 'user', content: trimmedMessage } 
        ];
    
        setIsLoading(true);
        setSavedChats(prevChats => prevChats
            .map(chat =>
                chat.id === currentChatId
                    ? {
                        ...chat,
                        title: isNewChat && previousMessages.length === 0 ?
                               trimmedMessage.substring(0, 40) + (trimmedMessage.length > 40 ? '...' : '') : chat.title,
                        messages: [...previousMessages, newUserMessage],
                        lastModified: new Date() 
                      }
                    : chat
            )
            .sort((a, b) => new Date(b.lastModified).getTime() - new Date(a.lastModified).getTime())
        );
            
        // ---  API ---
        let aiResponseText: string | null = null;
        let errorMessageText: string | null = null;
        let wasAborted = false;
        const controller = new AbortController();
        setAbortController(controller);
    
        try {
            const userSystemPrompt = localStorage.getItem('flowai_systemPrompt');
            const answerLanguage = localStorage.getItem('flowai_answerLanguage') || 'auto';
            const effectiveSystemPrompt = userSystemPrompt?.trim() ? userSystemPrompt.trim() : DEFAULT_SYSTEM_PROMPT;
            
            let messagesToSend: OllamaMessage[] = [...baseMessagesForApi]; 

            messagesToSend.unshift({ role: 'system', content: effectiveSystemPrompt });

            if (answerLanguage !== 'auto') {
                const languageInstruction = `\n\n(System note: Respond ONLY in ${answerLanguage}.)`;
                const lastMessageIndex = messagesToSend.length - 1;
                if (lastMessageIndex >= 0 && messagesToSend[lastMessageIndex].role === 'user') {
                    messagesToSend[lastMessageIndex].content += languageInstruction;
                }
            }
            
            console.log('Messages being sent to Ollama:', JSON.stringify(messagesToSend, null, 2));

            const ollamaResponse = await fetch('/ollama-api/chat', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ model: selectedModel, messages: messagesToSend, stream: false }),
                signal: controller.signal,
            });

            if (!ollamaResponse.ok) {
                let errorText = `Ollama API error! Status: ${ollamaResponse.status}`;
                try { 
                    const errorBody = await ollamaResponse.json(); 
                    errorText += ` - ${errorBody.error || JSON.stringify(errorBody)}`; 
                } catch (e) { errorText += ` - ${ollamaResponse.statusText}`; }
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
                console.error("Error during Ollama fetch:", error);
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
             showNotification(
                 "Flow-AI: Response Ready!",
                 {
                     body: `"${trimmedMessage.substring(0, 30)}..." -> ${aiResponseText.substring(0, 60)}${aiResponseText.length > 60 ? '...' : ''}`,
                     icon: '/logo.svg', 
                     chatId: currentChatId 
                 }
             );
        }

 
        if (aiResponseText && !wasAborted && currentChatId) { 
            try {
                const backendPayload = {
                    messages: [ 
                        { role: 'user', content: newUserMessage.text }, 
                        { role: 'assistant', content: aiResponseText } 
                    ]
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
                 setSavedChats(prevChats => prevChats.map(chat =>
                   chat.id === currentChatId ? {...chat, messages: [...(chat.messages || []), saveErrorMessageForUi ]} : chat
                ).sort((a, b) => new Date(b.lastModified).getTime() - new Date(a.lastModified).getTime()));
                
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