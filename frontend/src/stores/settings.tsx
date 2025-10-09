import { create } from 'zustand';
import axios from 'axios';
import type { Settings, UpdateSettingsPayload } from '../types/settings';

interface SettingsState {
    settings: Settings;
    isLoading: boolean;
    error: string | null;
    isSuccess: boolean; 

    fetchSettings: () => Promise<void>;
    updateSettings: (payload: UpdateSettingsPayload) => Promise<void>;
    resetSuccess: () => void; 
}

const API_BASE_URL = '/api';

export const useSettingsStore = create<SettingsState>((set) => ({
    settings: {
        main_model: '',
        support_model: '',
        system_prompt: '',
    },
    isLoading: false,
    error: null,
    isSuccess: false,

    fetchSettings: async () => {
        set({ isLoading: true, error: null, isSuccess: false });
        try {
            const response = await axios.get<Settings>(`${API_BASE_URL}/settings`);
            set({ settings: response.data, isLoading: false });
        } catch (error) {
            set({ error: 'Не вдалося завантажити налаштування', isLoading: false });
            console.error(error);
        }
    },

    updateSettings: async (payload: UpdateSettingsPayload) => {
        set({ isLoading: true, error: null, isSuccess: false });
        try {
            const response = await axios.post<Settings>(`${API_BASE_URL}/settings`, payload);
            set({ settings: response.data, isLoading: false, isSuccess: true });
        } catch (error) {
            set({ error: 'Не вдалося зберегти налаштування', isLoading: false });
            console.error(error);
        }
    },

    resetSuccess: () => {
        set({ isSuccess: false });
    },
}));