
import { useState, useEffect, useCallback } from 'react';
import type { OllamaTagModel } from '../types';

const OLLAMA_API_BASE_URL = '/ollama-api';

export function useModels() {
    const [availableModels, setAvailableModels] = useState<OllamaTagModel[]>([]);
    const [selectedModel, setSelectedModelInternal] = useState<string>('');
    const [selectedChatNamingModel, setSelectedChatNamingModelInternal] = useState<string>('disabled'); // <<< Новий стан
    const [modelsLoading, setModelsLoading] = useState<boolean>(false);
    const [modelsError, setModelsError] = useState<string | null>(null);

    const fetchModels = useCallback(async () => {
        console.log("Fetching models...");
        setModelsLoading(true);
        setModelsError(null);
        try {
            const response = await fetch(`${OLLAMA_API_BASE_URL}/tags`);
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`Failed to fetch models: ${response.status} ${errorText}`);
            }
            const data = await response.json();
            const fetchedModels: OllamaTagModel[] = data.models || [];
            setAvailableModels(fetchedModels);

            const savedModel = localStorage.getItem('flowai_selectedModel');
            if (savedModel && fetchedModels.some(m => m.name === savedModel)) {
                setSelectedModelInternal(savedModel);
            } else if (fetchedModels.length > 0) {
                const defaultModel = fetchedModels[0].name;
                setSelectedModelInternal(defaultModel);
                localStorage.setItem('flowai_selectedModel', defaultModel);
            } else {
                setSelectedModelInternal('');
            }

            const savedChatNamingModel = localStorage.getItem('flowai_selectedChatNamingModel');
            if (savedChatNamingModel) {
                if (savedChatNamingModel === 'disabled' || fetchedModels.some(m => m.name === savedChatNamingModel)) {
                    setSelectedChatNamingModelInternal(savedChatNamingModel);
                } else {
                    setSelectedChatNamingModelInternal('disabled');
                    localStorage.setItem('flowai_selectedChatNamingModel', 'disabled');
                }
            } else {
                setSelectedChatNamingModelInternal('disabled'); 
            }

        } catch (error) {
            console.error("Error fetching models:", error);
            setModelsError(error instanceof Error ? error.message : "An unknown error occurred");
            setAvailableModels([]);
            setSelectedModelInternal('');
            setSelectedChatNamingModelInternal('disabled'); 
        } finally {
            setModelsLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchModels();
    }, [fetchModels]);

    const setSelectedModel = (modelName: string) => {
        setSelectedModelInternal(modelName);
        localStorage.setItem('flowai_selectedModel', modelName);
    };


    const setSelectedChatNamingModel = (modelName: string) => {
        setSelectedChatNamingModelInternal(modelName);
        localStorage.setItem('flowai_selectedChatNamingModel', modelName);
    };

    return {
        availableModels,
        selectedModel,
        setSelectedModel,
        selectedChatNamingModel,    
        setSelectedChatNamingModel, 
        modelsLoading,
        modelsError,
        fetchModels
    };
}