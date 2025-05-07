import { useState } from 'react';
import { OllamaTagModel, OllamaTagsResponse } from '../types';

export function useModels() {
    const [availableModels, setAvailableModels] = useState<OllamaTagModel[]>([]);
    const [selectedModel, setSelectedModel] = useState<string>('phi3:mini'); 
    const [modelsLoading, setModelsLoading] = useState<boolean>(false);
    const [modelsError, setModelsError] = useState<string | null>(null);

    const fetchModels = async () => {
        console.log("Attempting to fetch Ollama models...");
        setModelsLoading(true); 
        setModelsError(null);
        
        try {
            const response = await fetch('/ollama-api/tags');
            if (!response.ok) {
                let errorMsg = `HTTP error ${response.status}: ${response.statusText}`;
                try { 
                    const errBody = await response.json(); 
                    errorMsg += ` - ${errBody.error || JSON.stringify(errBody)}`; 
                } catch (e) { 
                    /* ignore */ 
                }
                throw new Error(errorMsg);
            }
            
            const data: OllamaTagsResponse = await response.json();
            const models = data.models || [];
            setAvailableModels(models);
            
            const currentModelExists = models.some(m => m.name === selectedModel);
            if ((!currentModelExists || !selectedModel) && models.length > 0) {
                setSelectedModel(models[0].name);
            } else if (models.length === 0) {
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

    return {
        availableModels,
        selectedModel,
        setSelectedModel,
        modelsLoading,
        modelsError,
        fetchModels
    };
}