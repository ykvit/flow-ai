import React, { useState } from 'react';
import Modal from '../Modal/Modal';
import SettingsNav from './SettingsNav';
import SettingsContent from './SettingsContent';
import layoutStyles from './SettingsModal.module.css';

import GeneralIcon from '../../assets/settings/settings-general.svg?react';
import ModelsIcon from '../../assets/settings/settings-models.svg?react';
import ConnectionsIcon from '../../assets/settings/settings-connections.svg?react';
import DataIcon from '../../assets/settings/settings-data.svg?react';
import AboutIcon from '../../assets/settings/settings-about.svg?react';

export type SettingsCategory = 'general' | 'appearance' | 'models' | 'connections' | 'data' | 'about';

import { OllamaTagModel } from '../../types'; 

const categories: { id: SettingsCategory; label: string; icon: React.FC<React.SVGProps<SVGSVGElement>> }[] = [
  { id: 'general', label: 'General', icon: GeneralIcon },
  { id: 'models', label: 'Models', icon: ModelsIcon },
  { id: 'connections', label: 'Connections', icon: ConnectionsIcon },
  { id: 'data', label: 'Data Management', icon: DataIcon },
  { id: 'about', label: 'About', icon: AboutIcon },
];

interface SettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
  availableModels: OllamaTagModel[];
  selectedModel: string;
  onSelectModel: (modelName: string) => void;
  modelsLoading: boolean;
  modelsError: string | null;
  onRefreshModels?: () => void; 
}

const SettingsModal: React.FC<SettingsModalProps> = ({
  isOpen,
  onClose,
  availableModels,
  selectedModel,
  onSelectModel,
  modelsLoading,
  modelsError,
  onRefreshModels,
}) => {
  const [activeCategory, setActiveCategory] = useState<SettingsCategory>('general');

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Settings">
      <div className={layoutStyles.settingsLayout}>
        <SettingsNav
          categories={categories}
          activeCategory={activeCategory}
          onSelectCategory={setActiveCategory}
        />
        <SettingsContent
          activeCategory={activeCategory}
          availableModels={availableModels}
          selectedModel={selectedModel}
          onSelectModel={onSelectModel}
          modelsLoading={modelsLoading}
          modelsError={modelsError}
          onRefreshModels={onRefreshModels}
        />
      </div>
    </Modal>
  );
};

export default SettingsModal;