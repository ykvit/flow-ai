import { create } from 'zustand';
import axios from 'axios';
import type {
    Chat,
    ChatWithMessages,
    CreateMessagePayload,
    UpdateTitlePayload,
} from '../types/chat';

interface ChatState {
  chats: Chat[];
  currentChat: ChatWithMessages | null; 
  isLoading: boolean;
  error: string | null;

  fetchChats: () => Promise<void>;
  fetchChatById: (chatId: string) => Promise<void>;
  createMessage: (payload: CreateMessagePayload) => Promise<void>;
  deleteChat: (chatId: string) => Promise<void>;
  updateChatTitle: (
    chatId: string,
    payload: UpdateTitlePayload
  ) => Promise<void>;
}

const API_BASE_URL = '/api';

export const useChatStore = create<ChatState>((set, get) => ({

  chats: [],
  currentChat: null,
  isLoading: false,
  error: null,


fetchChats: async () => {
    set({ isLoading: true, error: null });
    try {
        const response = await axios.get<any>(`${API_BASE_URL}/chats`);
        const chatsArray = Array.isArray(response.data.chats) ? response.data.chats : [];
        
        set({ chats: chatsArray, isLoading: false });

    } catch (error) {
        console.error(error); 
        set({ error: 'Не вдалося завантажити чати', isLoading: false, chats: [] }); 
    }
},

  fetchChatById: async (chatId: string) => {
    set({ isLoading: true, error: null, currentChat: null });
    try {
      const response = await axios.get<ChatWithMessages>(`${API_BASE_URL}/chats/${chatId}`);
      set({ currentChat: response.data, isLoading: false });
    } catch (error) {
      set({ error: `'cant upload chats' ${chatId}`, isLoading: false });
      console.error(error);
    }
  },

  createMessage: async (payload: CreateMessagePayload) => {
    set({ isLoading: true, error: null });
    try {
      await axios.post(`${API_BASE_URL}/chats/messages`, payload);
      await get().fetchChatById(payload.chat_id);

    } catch (error) {
      set({ error: 'cant cend message', isLoading: false });
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
      set({ error: `cant delete chats ${chatId}`, isLoading: false });
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
      set({ error: `cant update header chat ${chatId}`, isLoading: false });
      console.error(error);
    }
  },
}));