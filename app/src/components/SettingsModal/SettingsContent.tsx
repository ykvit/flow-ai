// src/components/SettingsModal/SettingsContent.tsx
import React from 'react';
import styles from './SettingsContent.module.css';
import { SettingsCategory } from './SettingsModal';


import GeneralSettings from './categories/GeneralSettings';
import ModelsSettings from './categories/ModelsSettings';

// Remove the local definition and import the shared OllamaTagModel interface
import { OllamaTagModel } from '../../types'; 


interface SettingsContentProps {
  activeCategory: SettingsCategory;
  availableModels: OllamaTagModel[];
  selectedModel: string;
  onSelectModel: (modelName: string) => void;
  modelsLoading: boolean;
  modelsError: string | null;
  onRefreshModels?: () => void;
}

const SettingsContent: React.FC<SettingsContentProps> = ({
  activeCategory,
  availableModels,
  selectedModel,
  onSelectModel,
  modelsLoading,
  modelsError,
  onRefreshModels,
}) => {
  const renderContent = () => {
    switch (activeCategory) {
      case 'general':
        return <GeneralSettings />;
      case 'models':
        return <ModelsSettings
                  availableModels={availableModels}
                  currentModel={selectedModel}
                  onSelectModel={onSelectModel}
                  isLoading={modelsLoading}
                  error={modelsError}
                  onRefresh={onRefreshModels}
               />;
       default:
         return <div>Select a category</div>;
     }
  };

  return (
      <div className={styles.content}>
          {renderContent()}
        
      </div>
  );
};

export default SettingsContent;