
import { useState } from 'react'; 
import { Chat, Message, OllamaMessage, OllamaChatResponse } from '../types'; 
import { useNotifications } from './useNotifications';

const DEFAULT_SYSTEM_PROMPT = "You are Flow-ai. You are here to help, you have to be polite. If you don't know the answer, say you don't know.";

const TITLE_GENERATION_PROMPT_TEMPLATE = (userMessage: string) =>
    `Generate a concise and relevant title (maximum 5-7 words, preferably in the same language as the user's message if identifiable, otherwise English) for a chat that starts with the following user message: "${userMessage}". Output ONLY the title itself, with no prefixes like "Title:" or quotation marks around it. Be brief.`;

async function generateChatTitle(
    userMessage: string,
    namingModel: string
): Promise<string | null> {
    if (!namingModel || namingModel === 'disabled') {
        console.log('[generateChatTitle] Naming model is disabled or not provided.');
        return null;
    }
    console.log(`[generateChatTitle] Attempting to generate title with model: ${namingModel} for message starting with: "${userMessage.substring(0,50)}..."`);
    try {
        const prompt = TITLE_GENERATION_PROMPT_TEMPLATE(userMessage);
        console.log('[generateChatTitle] Sending request to Ollama for title generation...');
        const response = await fetch('/ollama-api/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                model: namingModel,
                messages: [{ role: 'user', content: prompt }],
                stream: false,
                options: { temperature: 0.3 }
            }),
        });

        if (!response.ok) {
            const errorText = await response.text();
            console.error(`[generateChatTitle] Title generation API error (${response.status}): ${errorText}`);
            return null;
        }

        const data: OllamaChatResponse = await response.json();
        console.log('[generateChatTitle] Ollama response data:', data);
        if (data.message && data.message.content) {
            let title = data.message.content.trim();
            console.log('[generateChatTitle] Raw title from Ollama:', title);
            title = title.replace(/^["']|["']$/g, '');
            title = title.replace(/^(Title:|"Title:"|Title is:|"Title is:")\s*/i, '');
            console.log("[generateChatTitle] Cleaned title:", title);
            return title.substring(0, 100);
        }
        console.error("[generateChatTitle] Title generation: Empty or invalid response format from model.");
        return null;
    } catch (error) {
        console.error("[generateChatTitle] Error during title generation:", error);
        return null;
    }
}

// Це має бути експорт функції useChat
export function useChat() { // <--- ПРАВИЛЬНЕ ІМ'Я ФУНКЦІЇ
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [abortController, setAbortController] = useState<AbortController | null>(null);
    const { showNotification } = useNotifications();

    const handleSendMessage = async (
        userMessageText: string,
        activeChatIdFromApp: string | null,
        createNewChatAndUpdateState: (selectedModel: string, initialTitle?: string) => Promise<string | null>,
        updateChatTitleForState: (chatId: string, newTitle: string) => Promise<void>,
        selectedModel: string,
        selectedChatNamingModel: string,
        currentSavedChats: Chat[], // Chat тип потрібен тут
        setSavedChatsCallback: React.Dispatch<React.SetStateAction<Chat[]>>,
        inputRef: React.RefObject<HTMLTextAreaElement | null>
    ) => {
        const trimmedMessage = userMessageText.trim();
        if (!trimmedMessage || isLoading) return;

        let currentChatId = activeChatIdFromApp;
        let isNewChatInstance = false;

        console.log('[handleSendMessage] Start. activeChatIdFromApp:', currentChatId);
        console.log('[handleSendMessage] selectedModel:', selectedModel);
        console.log('[handleSendMessage] selectedChatNamingModel:', selectedChatNamingModel);

         if (!currentChatId) {
            console.log('[handleSendMessage] No activeChatId, creating new chat with default title.');
            const newChatId = await createNewChatAndUpdateState(selectedModel, "New Chat");
            if (!newChatId) {
                console.error("[handleSendMessage] Failed to create new chat ID.");
                return;
            }
            currentChatId = newChatId;
            isNewChatInstance = true;
            console.log('[handleSendMessage] New chat created with ID:', currentChatId);
        }

        if (!currentChatId) {
            console.error("Critical: currentChatId is still null after creation attempt.");
            return;
        }

        if (!selectedModel) {
            const noModelErrorMsg: Message = { sender: 'assistant', text: 'Error: No AI model selected. Please select one in settings.' };
            setSavedChatsCallback(prev => prev.map(c => c.id === currentChatId ? {...c, messages: [...(c.messages || []), noModelErrorMsg ]} : c).sort((a,b) => b.lastModified.getTime() - a.lastModified.getTime()));
            return;
        }

        const newUserMessage: Message = { sender: 'user', text: trimmedMessage };

        const chatBeforeUpdate = currentSavedChats.find(c => c.id === currentChatId);
        const previousMessages: Message[] = chatBeforeUpdate?.messages || [];

        const isFirstUserMessageInChat = previousMessages.filter(msg => msg.sender === 'user').length === 0;

        console.log(`[handleSendMessage] Chat ID: ${currentChatId}, isNewChatInstance: ${isNewChatInstance}, previousMessages count: ${previousMessages.length}, isFirstUserMessageInChat: ${isFirstUserMessageInChat}`);

        setIsLoading(true);
        setSavedChatsCallback(prevChats => prevChats
            .map(chat =>
                chat.id === currentChatId
                    ? {
                        ...chat,
                        messages: [...previousMessages, newUserMessage],
                        lastModified: new Date()
                      }
                    : chat
            )
            .sort((a, b) => new Date(b.lastModified).getTime() - new Date(a.lastModified).getTime())
        );

        if (isFirstUserMessageInChat && selectedChatNamingModel && selectedChatNamingModel !== 'disabled') {
            console.log(`[handleSendMessage] This is the first user message in chat ${currentChatId}. Attempting to generate and update title.`);
            try {
                const generatedTitle = await generateChatTitle(trimmedMessage, selectedChatNamingModel);
                console.log('[handleSendMessage] generateChatTitle for update returned:', generatedTitle);
                if (generatedTitle && generatedTitle.trim() !== "" && currentChatId) {
                    await updateChatTitleForState(currentChatId, generatedTitle.trim());
                    console.log(`[handleSendMessage] Called updateChatTitleForState for chat ${currentChatId} with title "${generatedTitle.trim()}"`);
                } else {
                    console.log('[handleSendMessage] Title generation for update failed or returned empty.');
                }
            } catch (error) {
                console.error('[handleSendMessage] Error during generateChatTitle call for title update:', error);
            }
        }

        let aiResponseText: string | null = null;
        let errorMessageText: string | null = null;
        let wasAborted = false;
        const controller = new AbortController();
        setAbortController(controller);

        try {
            const userSystemPrompt = localStorage.getItem('flowai_systemPrompt');
            const answerLanguage = localStorage.getItem('flowai_answerLanguage') || 'auto';
            const effectiveSystemPrompt = userSystemPrompt?.trim() ? userSystemPrompt.trim() : DEFAULT_SYSTEM_PROMPT;

            const baseMessagesForApi: OllamaMessage[] = [
                ...previousMessages.map((msg: Message) => ({
                    role: msg.sender === 'assistant' ? 'assistant' : 'user' as ('user' | 'assistant'),
                    content: msg.text
                })),
                { role: 'user', content: trimmedMessage }
            ];

            let messagesToSend: OllamaMessage[] = [...baseMessagesForApi];
            messagesToSend.unshift({ role: 'system', content: effectiveSystemPrompt });

            if (answerLanguage !== 'auto') {
                const languageInstruction = `\n\n(System note: Respond ONLY in ${answerLanguage}.)`;
                const lastMessageIndex = messagesToSend.length - 1;
                if (lastMessageIndex >= 0 && messagesToSend[lastMessageIndex].role === 'user') {
                    messagesToSend[lastMessageIndex].content += languageInstruction;
                }
            }
            console.log('[handleSendMessage] Messages being sent to Ollama for chat response:', JSON.stringify(messagesToSend, null, 2));

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
            setSavedChatsCallback(prevChats => prevChats.map(chat => {
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
                 setSavedChatsCallback(prevChats => prevChats.map(chat =>
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