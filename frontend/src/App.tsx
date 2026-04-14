import { useEffect, useState } from 'react';
import ThemeLoader from './theme/ThemeLoader';
import ChatLayout from './components/ChatLayout';
import SettingsDialog from './components/SettingsDialog';
import { useChatStore } from './stores/chats';
import { useSettingsStore } from './stores/settings';
import { useModelsStore } from './stores/models';

function App() {
  const [settingsOpen, setSettingsOpen] = useState(false);
  const fetchChats = useChatStore((s) => s.fetchChats);
  const fetchSettings = useSettingsStore((s) => s.fetchSettings);
  const fetchModels = useModelsStore((s) => s.fetchModels);

  // Initialize data on mount
  useEffect(() => {
    fetchChats();
    fetchSettings();
    fetchModels();
  }, [fetchChats, fetchSettings, fetchModels]);

  return (
    <>
      <ThemeLoader />
      <ChatLayout onOpenSettings={() => setSettingsOpen(true)} />
      <SettingsDialog open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </>
  );
}

export default App;