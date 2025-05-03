import React from 'react';
import styles from '../SettingsContent.module.css';
import modelStyles from './ModelsSettings.module.css';
import RefreshIcon from '../../../assets/settings/refresh-con.svg?react'; 
import TrashIcon from '../../../assets/settings/delete-icon.svg?react'; 
import CustomDropdown, { DropdownOption } from '../../CustomDropdown/CustomDropdown';
import CloudDownloadIcon from '../../../assets/settings/download-icon.svg?react'; 

interface OllamaTagModel {
  name: string;
  size: number;
  modified_at: string; 
}

interface ModelsSettingsProps {
  availableModels: OllamaTagModel[];
  currentModel: string;
  onSelectModel: (modelName: string) => void;
  isLoading: boolean;
  error: string | null;
  onRefresh?: () => void;
  // onDeleteModel?: (modelName: string) => void;
}

const ModelsSettings: React.FC<ModelsSettingsProps> = ({
  availableModels,
  currentModel,
  onSelectModel,
  isLoading,
  error,
  onRefresh,
  // onDeleteModel,
}) => {

    const modelOptions: DropdownOption[] = availableModels.map(model => ({
        value: model.name,
        label: model.name 
    }));

  const handleDeleteClick = (modelName: string) => {
      console.log("Attempting to delete model:", modelName);
       if (window.confirm(`Are you sure you want to delete model "${modelName}"? This requires interacting with the Ollama API.`)) {
           // TODO: Implement actual deletion via API call
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

          {/* === Models for use === */}
          <div className={modelStyles.section}>
              <h3 className={modelStyles.sectionTitle}>Models for use</h3>
              <div className={modelStyles.settingsList}> {/* ... */}
                  {/* Default Model */}
                  <div className={styles.settingItem}>
                  <label className={styles.settingLabel} htmlFor="default-model-dropdown"> {/*  */}
                  Default Model
                  </label>
                      <div className={styles.settingControl}>
                          {isLoading && <span className={modelStyles.loadingText}>Loading...</span>}
                          {error && <span className={modelStyles.errorText}>Error</span>}
                          {!isLoading && !error && (
                              <CustomDropdown
                              id="default-model-dropdown" 
                              options={modelOptions} 
                              selectedValue={currentModel} 
                              onChange={onSelectModel} 
                              placeholder="Select model..."
                              disabled={modelOptions.length === 0}
                          />
                        )}
                        </div>
                    </div>

                  <div className={styles.settingItem}>
                      <label className={styles.settingLabel}>Default Embedder Model</label>
                      <div className={styles.settingControl}>
                           <select className={modelStyles.styledSelect} disabled>
                              <option>Model 1 (example)</option>
                           </select>
                      </div>
                  </div>
                  <div className={styles.settingItem}>
                      <label className={styles.settingLabel}>Support Model</label>
                       <div className={styles.settingControl}>
                           <select className={modelStyles.styledSelect} disabled>
                              <option>Model 1 (example)</option>
                           </select>
                      </div>
                  </div>
              </div>
          </div>

          {/* === Manage Local Models === */}
          <div className={modelStyles.section}>
              <div className={modelStyles.manageHeader}>
                  <h3 className={modelStyles.sectionTitle}>Manage Local Models</h3>
                  <div className={modelStyles.manageIcons}>
                      {/* <button className={modelStyles.iconButton} aria-label="Download models">
                          <CloudDownloadIcon />
                      </button> */}
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

               {isLoading && <div className={modelStyles.loadingText}>Loading model list...</div>}
               {error && <div className={modelStyles.errorText}>Failed to load models: {error}</div>}

              <div className={`${modelStyles.settingsList} ${modelStyles.modelManageList}`}> 
                  {!isLoading && !error && availableModels.length === 0 && (
                       <p className={modelStyles.noModelsText}>No local models found.</p>
                  )}
                  {!isLoading && !error && availableModels.map(model => (
                      <div key={model.name} className={modelStyles.modelItem}>
                          <div className={modelStyles.modelInfo}>
                              <span className={modelStyles.modelName}>{model.name}</span>
                              <span className={modelStyles.modelDetails}>
                                  {formatBytes(model.size)} - Modified: {new Date(model.modified_at).toLocaleDateString()}
                              </span>
                          </div>
                          <button
                              className={modelStyles.iconButton}
                              aria-label={`Delete model ${model.name}`}
                              title={`Delete model ${model.name}`}
                              onClick={() => handleDeleteClick(model.name)}
                          >
                              <TrashIcon />
                          </button>
                      </div>
                  ))}
              </div>
               {/* <div className={modelStyles.pullModelSection}> ... </div> */}
          </div>
      </div>
  );
};

export default ModelsSettings;