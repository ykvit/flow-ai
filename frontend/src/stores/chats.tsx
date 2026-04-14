import { create } from 'zustand';
import axios from 'axios';
import type {
    Chat,
    ChatWithMessages,
    CreateMessagePayload,
    UpdateTitlePayload,
    RegenerateMessagePayload,
} from '../types/chat';

interface ChatState {
  chats: Chat[];
  currentChat: ChatWithMessages | null;
  isLoading: boolean;
  isStreaming: boolean;
  streamingContent: string;
  error: string | null;

  fetchChats: () => Promise<void>;
  fetchChatById: (chatId: string) => Promise<void>;
  createMessage: (payload: CreateMessagePayload) => Promise<void>;
  deleteChat: (chatId: string) => Promise<void>;
  updateChatTitle: (
    chatId: string,
    payload: UpdateTitlePayload
  ) => Promise<void>;
  regenerateMessage: (payload: RegenerateMessagePayload) => Promise<void>;
  switchBranch: (messageId: string) => Promise<void>;
  setCurrentChat: (chat: ChatWithMessages | null) => void;
  clearCurrentChat: () => void;
  createNewChat: () => void;
}

const API_BASE_URL = '/api/v1';

export const useChatStore = create<ChatState>((set, get) => ({

  chats: [],
  currentChat: null,
  isLoading: false,
  isStreaming: false,
  streamingContent: '',
  error: null,

  fetchChats: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await axios.get<any>(`${API_BASE_URL}/chats`);
      const chatsArray = Array.isArray(response.data) ? response.data : (Array.isArray(response.data?.chats) ? response.data.chats : []);
      set({ chats: chatsArray, isLoading: false });
    } catch (error) {
      console.error(error);
      set({ error: 'Failed to load chats', isLoading: false, chats: [] });
    }
  },

  fetchChatById: async (chatId: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await axios.get<ChatWithMessages>(`${API_BASE_URL}/chats/${chatId}/tree`);
      set({ currentChat: response.data, isLoading: false });
    } catch (error) {
      set({ error: `Failed to load chat ${chatId}`, isLoading: false });
      console.error(error);
    }
  },

  createMessage: async (payload: CreateMessagePayload) => {
    set({ isStreaming: true, streamingContent: '', error: null });

    // Optimistically add user message to the current chat
    const userMessage = {
      id: `temp-${Date.now()}`,
      content: payload.content,
      role: 'user' as const,
      model: payload.model,
      parent_id: '',
      metadata: {},
      timestamp: new Date().toISOString(),
      is_active: true,
    };

    set((state) => ({
      currentChat: state.currentChat
        ? { ...state.currentChat, messages: [...state.currentChat.messages, userMessage] }
        : {
            id: '',
            title: 'New Chat',
            model: payload.model,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            user_id: '',
            messages: [userMessage]
          },
    }));

    try {
      const response = await fetch(`${API_BASE_URL}/chats/messages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...payload, stream: true }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error ${response.status}`);
      }

      if (!response.body) {
        throw new Error('Response body is empty');
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let fullContent = '';
      let chatId = payload.chat_id;

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line.trim() === '') continue;
          if (line.startsWith('data: ')) {
            const jsonStr = line.slice(6).trim();
            if (jsonStr === '[DONE]') continue; // Common EOF marker in SSE
            try {
              const data = JSON.parse(jsonStr);

              // Capture chat_id from first response (for new chats)
              if (data.chat_id && !chatId) {
                chatId = data.chat_id;
              }

              // Accumulate streaming content
              if (data.content) {
                fullContent += data.content;
                set({ streamingContent: fullContent });
              }
            } catch (e) {
              console.error('Failed to parse SSE line:', jsonStr, e);
            }
          }
        }
      }

      // After streaming is complete, refresh the chat
      set({ isStreaming: false, streamingContent: '' });
      if (chatId) {
        await get().fetchChatById(chatId);
        await get().fetchChats();
      }
    } catch (error) {
      set({ error: 'Failed to send message', isStreaming: false, streamingContent: '' });
      console.error(error);
    }
  },

  regenerateMessage: async (payload: RegenerateMessagePayload) => {
    if (get().isStreaming) return;

    set({ isStreaming: true, streamingContent: '', error: null });

    try {
      let fullContent = '';
      const response = await fetch(`${API_BASE_URL}/chats/${payload.chat_id}/messages/${payload.message_id}/regenerate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (!response.body) throw new Error('ReadableStream not supported');

      const reader = response.body.getReader();
      const decoder = new TextDecoder('utf-8');
      let buffer = '';
      let chatId = payload.chat_id;

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line.trim() === '') continue;
          if (line.startsWith('data: ')) {
            const jsonStr = line.slice(6).trim();
            if (jsonStr === '[DONE]') continue;
            try {
              const data = JSON.parse(jsonStr);

              if (data.error) {
                set({ error: data.error, isStreaming: false });
                break;
              }

              if (data.chat_id && !chatId) {
                chatId = data.chat_id;
              }

              if (data.content) {
                fullContent += data.content;
                set({ streamingContent: fullContent });
              }
            } catch (e) {
              console.error('Failed to parse SSE line:', jsonStr, e);
            }
          }
        }
      }

      set({ isStreaming: false, streamingContent: '' });
      if (chatId) {
        await get().fetchChatById(chatId);
        await get().fetchChats();
      }
    } catch (error) {
      set({ error: 'Failed to regenerate message', isStreaming: false, streamingContent: '' });
      console.error(error);
    }
  },

  deleteChat: async (chatId: string) => {
    set({ isLoading: true, error: null });
    try {
      await axios.delete(`${API_BASE_URL}/chats/${chatId}`);
      set((state) => ({
        chats: state.chats.filter((chat) => chat.id !== chatId),
        currentChat: state.currentChat?.id === chatId ? null : state.currentChat,
        isLoading: false,
      }));
    } catch (error) {
      set({ error: `Failed to delete chat ${chatId}`, isLoading: false });
      console.error(error);
    }
  },

  updateChatTitle: async (chatId: string, payload: UpdateTitlePayload) => {
    set({ isLoading: true, error: null });
    try {
      await axios.put(`${API_BASE_URL}/chats/${chatId}/title`, payload);
      set((state) => ({
        chats: state.chats.map((chat) =>
          chat.id === chatId ? { ...chat, title: payload.title } : chat
        ),
        currentChat:
          state.currentChat?.id === chatId
            ? { ...state.currentChat, title: payload.title }
            : state.currentChat,
        isLoading: false,
      }));
    } catch (error) {
      set({ error: `Failed to update chat title ${chatId}`, isLoading: false });
      console.error(error);
    }
  },
  switchBranch: async (messageId: string) => {
    const { currentChat } = get();
    if (!currentChat) return;

    set({ isLoading: true, error: null });
    try {
      await axios.post(`${API_BASE_URL}/chats/${currentChat.id}/messages/${messageId}/activate`);
      await get().fetchChatById(currentChat.id);
    } catch (error) {
      set({ error: 'Failed to switch branch', isLoading: false });
      console.error(error);
    }
  },
  setCurrentChat: (chat: ChatWithMessages | null) => {
    set({ currentChat: chat, error: null });
  },

  clearCurrentChat: () => {
    set({ currentChat: null, streamingContent: '', error: null });
  },

  createNewChat: () => {
    set({ currentChat: null, streamingContent: '', error: null });
  },
}));