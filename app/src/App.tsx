import { useRef, useEffect, useMemo, useState } from 'react';
import styles from './App.module.css';

import Sidebar from './components/Sidebar/Sidebar';
import Header from './components/Header/Header';
import WelcomeMessage from './components/WelcomeMessage/WelcomeMessage';
import MessageList from './components/MessageList/MessageList';
import ChatInput from './components/ChatInput/ChatInput';
import SettingsModal from './components/SettingsModal/SettingsModal';

import SettingsIcon from './assets/settings-button.svg?react';

import { useChats } from './hooks/useChats';
import { useModels } from './hooks/useModels';
import { useChat } from './hooks/useChat';
import { useAppUI } from './hooks/useAppUI';

function App() {
    const {
        savedChats,
        setSavedChats,
        activeChatId,
        setActiveChatId,
        isChatLoading: isChatHistoryLoading,
        loadInitialChats,
        createNewChat,
        updateChatTitle,
        deleteChat,
        editChat
    } = useChats();

    const {
        availableModels,
        selectedModel,
        setSelectedModel,
        selectedChatNamingModel,
        setSelectedChatNamingModel,
        modelsLoading,
        modelsError,
        fetchModels
    } = useModels();

    const {
        isLoading: isSendingMessage,
        handleSendMessage: sendMessageFromUseChat,
        stopGenerating
    } = useChat();

    const handleSendMessageWrapper = (userMessageText: string) => {
        return sendMessageFromUseChat(
            userMessageText,
            activeChatId,
            createNewChat,
            updateChatTitle,
            selectedModel,
            selectedChatNamingModel,
            savedChats,
            setSavedChats,
            inputRef
        );
    };

    const [isAppLoading, setIsAppLoading] = useState<boolean>(true);

    const {
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

    useEffect(() => {
        setTimeout(() => inputRef.current?.focus(), 100);
    }, [activeChatId]);

    useEffect(() => {
        let isMounted = true;
        setIsAppLoading(true);

        Promise.all([
            loadInitialChats(),
        ])
            .catch(err => {
                if (isMounted) {
                    console.error("Error during initial data load (chats or models):", err);
                }
            })
            .finally(() => {
                if (isMounted) {
                    setIsAppLoading(false);
                }
            });

        return () => {
            isMounted = false;
        };
    }, [loadInitialChats]);

    useEffect(() => {
        if (!isAppLoading && !activeChatId && savedChats.length > 0) {
            setActiveChatId(savedChats[0].id);
        } else if (!isAppLoading && savedChats.length === 0 && activeChatId) {
            setActiveChatId(null);
        }
    }, [isAppLoading, savedChats, activeChatId, setActiveChatId]);

    const currentChatMessages = useMemo(() => {
        if (!activeChatId) return [];
        const chat = savedChats.find(c => c.id === activeChatId);
        return chat?.messages || [];
    }, [activeChatId, savedChats]);

    const showWelcomeMessage = useMemo(() => {
        return (!activeChatId || (activeChatId && currentChatMessages.length === 0)) &&
               !isChatHistoryLoading && !isSendingMessage && !isAppLoading;
    }, [activeChatId, currentChatMessages, isChatHistoryLoading, isSendingMessage, isAppLoading]);

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
                { showWelcomeMessage && <WelcomeMessage isSidebarOpen={isSidebarOpen} /> }

                {!showWelcomeMessage && isChatHistoryLoading && currentChatMessages.length === 0 && (
                    <div className={styles.centralLoading}>Loading chat history...<div className={styles.spinner}></div></div>
                )}
                {!showWelcomeMessage && isSendingMessage && currentChatMessages.length === 0 && !isChatHistoryLoading && (
                    <div className={styles.centralLoading}>AI is thinking...<div className={styles.spinner}></div></div>
                )}
                {currentChatMessages.length > 0 && <MessageList messages={currentChatMessages} />}
            </div>

            <ChatInput
                onSendMessage={handleSendMessageWrapper}
                isLoading={isSendingMessage || isChatHistoryLoading}
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
                selectedChatNamingModel={selectedChatNamingModel}
                onSelectChatNamingModel={setSelectedChatNamingModel}
                modelsLoading={modelsLoading}
                modelsError={modelsError}
                onRefreshModels={fetchModels}
            />
        </div>
    );
}

export default App;