import { useRef, useEffect, useMemo } from 'react';
import styles from './App.module.css';

import Sidebar from './components/Sidebar/Sidebar';
import Header from './components/Header/Header';
import WelcomeMessage from './components/WelcomeMessage/WelcomeMessage';
import MessageList from './components/MessageList/MessageList';
import ChatInput from './components/ChatInput/ChatInput';
import SettingsModal from './components/SettingsModal/SettingsModal';

import SettingsIcon from './assets/settings-button.svg?react';

// Import custom hooks
import { useChats } from './hooks/useChats';
import { useModels } from './hooks/useModels';
import { useChat } from './hooks/useChat';
import { useAppUI } from './hooks/useAppUI';

function App() {
    // Use our custom hooks
    const { 
        savedChats, 
        setSavedChats,
        activeChatId, 
        setActiveChatId, 
        isChatLoading, 
        loadInitialChats, 
        createNewChat, 
        deleteChat, 
        editChat
    } = useChats();

    const { 
        availableModels, 
        selectedModel, 
        setSelectedModel, 
        modelsLoading, 
        modelsError, 
        fetchModels 
    } = useModels();

    const { 
        isLoading, 
        handleSendMessage: sendMessage, 
        stopGenerating 
    } = useChat();

    const { 
        isAppLoading, 
        setIsAppLoading, 
        isSidebarOpen, 
        isSettingsModalOpen, 
        openSettingsModal, 
        closeSettingsModal, 
        toggleSidebar 
    } = useAppUI(activeChatId, savedChats);

    const inputRef = useRef<HTMLTextAreaElement>(null);
    const activeChat = useMemo(() => {
        if (!activeChatId) return null;
        return savedChats.find(chat => chat.id === activeChatId);
    }, [activeChatId, savedChats]);

    const activeChatTitle = activeChat ? activeChat.title : null; 

    // Focus the input field when active chat changes
    useEffect(() => {
        setTimeout(() => inputRef.current?.focus(), 100);
    }, [activeChatId]);

    // Initial data loading
    useEffect(() => {
        setIsAppLoading(true);
        Promise.all([loadInitialChats(), fetchModels()])
            .catch(err => console.error("Error during initial data load:", err))
            .finally(() => setIsAppLoading(false));
    }, []);

    // Wrapper for sending a message (simplifies the parameter list)
    const handleSendMessage = (userMessageText: string) => {
        return sendMessage(
            userMessageText,
            activeChatId,
            createNewChat,
            selectedModel,
            savedChats,
            setSavedChats,
            inputRef
        );
    };

    const currentChatMessages = useMemo(() => {
        if (!activeChatId) return [];
        return savedChats.find(chat => chat.id === activeChatId)?.messages || [];
    }, [activeChatId, savedChats]);

    if (isAppLoading) {
        return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', fontSize: '1.5em', color: '#ccc', background: '#1a1a1a' }}>Loading Application...</div>;
    }

    return (
        <div className={`${styles.chatAppContainer} ${isSidebarOpen ? styles.sidebarOpen : ''}`}>
            <Sidebar
                isOpen={isSidebarOpen}
                savedChats={savedChats.map(c => ({ id: c.id, title: c.title }))} 
                activeChatId={activeChatId}
                onSelectChat={setActiveChatId}
                onDeleteChat={deleteChat}
                onEditChat={editChat} 
            />

            <Header
                onToggleSidebar={toggleSidebar}
                isSidebarOpen={isSidebarOpen}
                onCreateNewChat={() => createNewChat(selectedModel)}
                chatTitle={activeChatTitle}
            />

            <div className={styles.mainContent}>
                {(!activeChatId || (activeChatId && currentChatMessages.length === 0)) && !isChatLoading && !isLoading && <WelcomeMessage />}
                {isChatLoading && currentChatMessages.length === 0 && (
                    <div className={styles.centralLoading}>Loading chat history...<div className={styles.spinner}></div></div>
                )}
                {isLoading && currentChatMessages.length === 0 && !isChatLoading && (
                    <div className={styles.centralLoading}>AI is thinking...<div className={styles.spinner}></div></div>
                )}

                {currentChatMessages.length > 0 && <MessageList messages={currentChatMessages} />}
            </div>

            <ChatInput
                onSendMessage={handleSendMessage}
                isLoading={isLoading || isChatLoading} 
                inputRef={inputRef}
                onStopGenerating={stopGenerating}
                isSidebarOpen={isSidebarOpen}
            />

            <button className={styles.bottomRightSettingsButton} onClick={openSettingsModal} aria-label="Open Settings">
                <SettingsIcon />
            </button>

            <SettingsModal
                isOpen={isSettingsModalOpen}
                onClose={closeSettingsModal}
                availableModels={availableModels}
                selectedModel={selectedModel}
                onSelectModel={setSelectedModel}
                modelsLoading={modelsLoading}
                modelsError={modelsError}
                onRefreshModels={fetchModels}
            />
        </div>
    );
}

export default App;