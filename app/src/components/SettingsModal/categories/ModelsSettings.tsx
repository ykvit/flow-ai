import React from 'react';
import styles from '../SettingsContent.module.css';
import modelStyles from './ModelsSettings.module.css';
import RefreshIcon from '../../../assets/settings/refresh-con.svg?react';
import TrashIcon from '../../../assets/settings/delete-icon.svg?react';
import CustomDropdown, { DropdownOption } from '../../CustomDropdown/CustomDropdown';
import ChatNamingSettings from './ChatNamingSettings'; 

import type { OllamaTagModel } from '../../../types';


interface ModelsSettingsProps {
  availableModels: OllamaTagModel[];
  currentModel: string;
  onSelectModel: (modelName: string) => void;
  selectedChatNamingModel: string; 
  onSelectChatNamingModel: (modelName: string) => void;
  isLoading: boolean;
  error: string | null;
  onRefresh?: () => void;
}

const ModelsSettings: React.FC<ModelsSettingsProps> = ({
  availableModels,
  currentModel,
  onSelectModel,
  selectedChatNamingModel, 
  onSelectChatNamingModel,
  isLoading,
  error,
  onRefresh,
}) => {

    const modelOptions: DropdownOption[] = availableModels.map(model => ({
        value: model.name,
        label: model.name
    }));

  const handleDeleteClick = (modelName: string) => {
      console.log("Attempting to delete model:", modelName);
       if (window.confirm(`Are you sure you want to delete model "${modelName}"? This requires interacting with the Ollama API.`)) {
           alert(`Deletion for "${modelName}" not implemented yet.`);
       }
  };

  const formatBytes = (bytes: number, decimals = 2) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
  }

  return (
      <div className={modelStyles.modelsSettingsContainer}>
          <div className={styles.section}>
              <h3 className={styles.sectionTitle}>Models for use</h3>
              <div className={styles.settingsList}>
                  <div className={styles.settingItem}>
                    <label className={styles.settingLabel} htmlFor="default-model-dropdown">
                      Default Model
                    </label>
                      <div className={styles.settingControl}>
                          {isLoading && modelOptions.length === 0 && <span className={modelStyles.loadingText}>Loading models...</span>}
                          {error && modelOptions.length === 0 && <span className={modelStyles.errorText}>Error loading</span>}
                          {(!isLoading || modelOptions.length > 0) && !error && (
                              <CustomDropdown
                                id="default-model-dropdown"
                                options={modelOptions}
                                selectedValue={currentModel}
                                onChange={onSelectModel}
                                placeholder="Select model..."
                                disabled={modelOptions.length === 0 || isLoading}
                              />
                          )}
                        </div>
                    </div>

                    
                    <ChatNamingSettings
                        availableModels={availableModels}
                        isLoading={isLoading}
                        selectedChatNamingModel={selectedChatNamingModel}
                        onSelectChatNamingModel={onSelectChatNamingModel}
                    />

                  <div className={styles.settingItem}>
                      <label className={styles.settingLabel}> Embedder Model</label>
                      <div className={styles.settingControl}>
                           <select className={styles.styledSelect} disabled>
                              <option>Not configured</option>
                           </select>
                      </div>
                  </div>
              </div>
          </div>

          <div className={styles.section}>
              <div className={modelStyles.manageHeader}>
                  <h3 className={modelStyles.sectionTitleInHeader}>Manage Local Models</h3>
                  <div className={modelStyles.manageIcons}>
                      <button
                          className={modelStyles.iconButton}
                          aria-label="Refresh model list"
                          onClick={onRefresh}
                          disabled={isLoading}
                          title="Refresh model list"
                       >
                          <RefreshIcon />
                      </button>
                  </div>
              </div>

               {isLoading && availableModels.length === 0 && <div className={modelStyles.loadingText}>Loading model list...</div>}
               {error && availableModels.length === 0 && <div className={`${modelStyles.errorText} ${modelStyles.centeredStatusText}`}>Failed to load models: {error}</div>}
               {!isLoading && !error && availableModels.length === 0 && (
                    <div className={`${modelStyles.noModelsText} ${modelStyles.centeredStatusText}`}>
                        No local models found. <br/> You can pull models using Ollama CLI.
                    </div>
                )}

              {availableModels.length > 0 && (
                <div className={`${styles.settingsList} ${modelStyles.modelManageList}`}>
                    {!isLoading && !error && availableModels.map(model => (
                        <div key={model.name} className={modelStyles.modelItem}>
                            <div className={modelStyles.modelInfo}>
                                <span className={modelStyles.modelName}>{model.name}</span>
                                <span className={modelStyles.modelDetails}>
                                    {formatBytes(model.size)} - Modified: {new Date(model.modified_at).toLocaleDateString()}
                                </span>
                            </div>
                            <button
                                className={`${modelStyles.iconButton} ${modelStyles.deleteButton}`}
                                aria-label={`Delete model ${model.name}`}
                                title={`Delete model ${model.name}`}
                                onClick={() => handleDeleteClick(model.name)}
                            >
                                <TrashIcon />
                            </button>
                        </div>
                    ))}
                </div>
              )}
          </div>
      </div>
  );
};

export default ModelsSettings;