import { create } from 'zustand';
import axios from 'axios';

import type {
    Model,
    ModelDetails,
    PullStatus,
    DeleteModelPayload,
    PullModelPayload,
    ShowModelPayload,
} from '../types/models';

interface ModelsState {
    models: Model[];
    currentModelDetails: ModelDetails | null;
    pullStatus: PullStatus | null;
    isLoading: boolean;
    error: string | null;

    fetchModels: () => Promise<void>;
    deleteModel: (payload: DeleteModelPayload) => Promise<void>;
    showModelInfo: (payload: ShowModelPayload) => Promise<void>;
    pullModel: (payload: PullModelPayload) => Promise<void>;
}

const API_BASE_URL = '/api';

export const useModelsStore = create<ModelsState>((set, get) => ({
    models: [],
    currentModelDetails: null,
    pullStatus: null,
    isLoading: false,
    error: null,
    fetchModels: async () => {
        set({ isLoading: true, error: null });
        try {
            const response = await axios.get<{ models: Model[] }>(`${API_BASE_URL}/models`);
            set({ models: response.data.models, isLoading: false });
        } catch (error) {
            set({ error: 'Не вдалося завантажити моделі', isLoading: false });
            console.error(error);
        }
    },

    deleteModel: async (payload: DeleteModelPayload) => {
        set({ isLoading: true, error: null });
        try {
            await axios.delete(`${API_BASE_URL}/models`, { data: payload });
            set((state) => ({
                models: state.models.filter((model) => model.name !== payload.name),
                isLoading: false,
            }));
        } catch (error) {
            set({ error: `Не вдалося видалити модель ${payload.name}`, isLoading: false });
            console.error(error);
        }
    },

    showModelInfo: async (payload: ShowModelPayload) => {
        set({ isLoading: true, error: null, currentModelDetails: null });
        try {
            const response = await axios.post<ModelDetails>(`${API_BASE_URL}/models/show`, payload);
            set({ currentModelDetails: response.data, isLoading: false });
        } catch (error) {
            set({ error: `Не вдалося отримати інформацію про модель ${payload.name}`, isLoading: false });
            console.error(error);
        }
    },

    pullModel: async (payload: PullModelPayload) => {
        set({ pullStatus: { name: payload.name, status: 'Starting...' }, error: null });

        try {
            const response = await fetch(`${API_BASE_URL}/models/pull`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ...payload, stream: true }),
            });

            if (!response.body) throw new Error('Response body is empty');

            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let buffer = '';

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                buffer += decoder.decode(value, { stream: true });
                const lines = buffer.split('\n');
                buffer = lines.pop() || '';

                for (const line of lines) {
                    if (line.trim() === '') continue;
                    const statusUpdate: PullStatus = JSON.parse(line);
                    
                    set({ pullStatus: statusUpdate });
                }
            }
        } catch (error) {
            const errorMessage = `Не вдалося завантажити модель ${payload.name}`;
            set({ error: errorMessage });
            console.error(error);
        } finally {
            set({ pullStatus: null });
            await get().fetchModels();
        }
    },
}));