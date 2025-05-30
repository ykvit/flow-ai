import React from 'react';
import styles from '../SettingsContent.module.css';
import CustomDropdown, { DropdownOption } from '../../CustomDropdown/CustomDropdown';
import type { OllamaTagModel } from '../../../types';

interface ChatNamingSettingsProps {
  availableModels: OllamaTagModel[];
  isLoading: boolean; 
  selectedChatNamingModel: string; 
  onSelectChatNamingModel: (modelName: string) => void;
}

const ChatNamingSettings: React.FC<ChatNamingSettingsProps> = ({
  availableModels,
  isLoading,
  selectedChatNamingModel,
  onSelectChatNamingModel,
}) => {
  const chatNamingModelOptions: DropdownOption[] = [
    { value: 'disabled', label: 'Disabled (Manual Naming)' },
    ...availableModels.map(model => ({
      value: model.name,
      label: `${model.name} (auto-title)`
    }))
  ];

  return (
    <div className={styles.settingItem}>
      <label className={styles.settingLabel} htmlFor="chat-naming-model-dropdown">
        Chat Naming Model
      </label>
      <div className={styles.settingControl}>
        <CustomDropdown
          id="chat-naming-model-dropdown"
          options={chatNamingModelOptions}
          selectedValue={selectedChatNamingModel}
          onChange={onSelectChatNamingModel}
          placeholder="Select model or disable..."
          disabled={isLoading || availableModels.length === 0}
        />
      </div>
    </div>
  );
};

export default ChatNamingSettings;