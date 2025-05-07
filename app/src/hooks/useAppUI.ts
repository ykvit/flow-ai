import { useState, useEffect } from 'react';

const DEFAULT_FAVICON = '/logo.svg'; 
const ACTIVE_CHAT_FAVICON = '/logo.svg'; 
const DEFAULT_TITLE = 'Flow-ai'; 

export function useAppUI(activeChatId: string | null, savedChats: any[]) {
    const [isAppLoading, setIsAppLoading] = useState<boolean>(true);
    const [isSidebarOpen, setIsSidebarOpen] = useState<boolean>(false);
    const [isSettingsModalOpen, setIsSettingsModalOpen] = useState<boolean>(false);

    useEffect(() => {
        const faviconElement = document.getElementById('favicon') as HTMLLinkElement | null;
        const activeChat = savedChats.find(chat => chat.id === activeChatId);
        
        if (activeChat) {
            document.title = activeChat.title || 'Chat';
            if (faviconElement) faviconElement.href = ACTIVE_CHAT_FAVICON;
        } else {
            document.title = DEFAULT_TITLE;
            if (faviconElement) faviconElement.href = DEFAULT_FAVICON;
        }
    }, [activeChatId, savedChats]);

    const openSettingsModal = () => setIsSettingsModalOpen(true);
    const closeSettingsModal = () => setIsSettingsModalOpen(false);
    const toggleSidebar = () => setIsSidebarOpen(!isSidebarOpen);

    return {
        isAppLoading,
        setIsAppLoading,
        isSidebarOpen,
        isSettingsModalOpen,
        openSettingsModal,
        closeSettingsModal,
        toggleSidebar
    };
}