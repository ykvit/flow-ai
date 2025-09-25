
import { useRef, useEffect, useState, useMemo } from 'react';
import styles from './App.module.css';

import Sidebar from './components/Sidebar/Sidebar';
import Header from './components/Header/Header';
import WelcomeMessage from './components/WelcomeMessage/WelcomeMessage';
import MessageList from './components/MessageList/MessageList';
import ChatInput from './components/ChatInput/ChatInput';
// import SettingsModal from './components/SettingsModal/SettingsModal';
// import SettingsIcon from './assets/settings-button.svg?react?react';

import { useChatStore } from './stores/chats';
import { useSettingsStore } from './stores/settings';

function App() {

    const {
        chats,
        currentChat,
        isLoading: isChatLoading, 
        fetchChats,
        fetchChatById,
        deleteChat,
        updateChatTitle,
        createMessage,

    } = useChatStore();

    const { settings, fetchSettings } = useSettingsStore();

    const [isSidebarOpen, setSidebarOpen] = useState(true);
    // const [isSettingsModalOpen, setSettingsModalOpen] = useState(false);

    const inputRef = useRef<HTMLTextAreaElement>(null);

    useEffect(() => {
        fetchChats();
        fetchSettings();
    }, [fetchChats, fetchSettings]);

    useEffect(() => {
        if (!isChatLoading && !currentChat && chats.length > 0) {
            fetchChatById(chats[0].id);
        }
    }, [isChatLoading, chats, currentChat, fetchChatById]);

    const createNewChat = () => {

        useChatStore.setState({ currentChat: null });
        setTimeout(() => inputRef.current?.focus(), 100);
    };

    const handleSendMessage = (messageText: string) => {
        if (!messageText.trim()) return;

        createMessage({

            chat_id: currentChat?.id || '', 
            content: messageText,
            model: settings.main_model, 
            support_model: settings.support_model, 
            system_prompt: settings.system_prompt, 
        });
    };
    
    
    const handleEditChatTitle = (chatId: string) => {
        const currentTitle = chats.find(c => c.id === chatId)?.title || '';
        const newTitle = prompt("Enter new chat title:", currentTitle);
        if (newTitle && newTitle.trim() !== currentTitle) {
            updateChatTitle(chatId, { title: newTitle });
        }
    };
    
    const handleDeleteChat = (chatId: string) => {
        if (window.confirm("Are you sure you want to delete this chat?")) {
            deleteChat(chatId);
        }
    };

    // const currentChatMessages = currentChat?.messages || [];

    const showWelcomeMessage = useMemo(() => {
        return !currentChat && !isChatLoading;
    }, [currentChat, isChatLoading]);

    if (useChatStore.getState().isLoading && chats.length === 0) {
        return <div className={styles.centralLoading}>Loading Application...</div>;
    }

    return (
        <div className={`${styles.chatAppContainer} ${isSidebarOpen ? styles.sidebarOpen : ''}`}>
            <Sidebar
                isOpen={isSidebarOpen}
                savedChats={chats} 
                activeChatId={currentChat?.id || null}
                onSelectChat={fetchChatById} 
                onDeleteChat={handleDeleteChat}
                onEditChat={handleEditChatTitle}
            />

            <Header
                onToggleSidebar={() => setSidebarOpen(!isSidebarOpen)}
                isSidebarOpen={isSidebarOpen}
                onCreateNewChat={createNewChat}
                chatTitle={currentChat?.title || ''}
            />

            <div className={styles.mainContent}>
                { showWelcomeMessage && <WelcomeMessage isSidebarOpen={isSidebarOpen} /> }

                { isChatLoading && !currentChat && <div className={styles.centralLoading}>Loading chat...<div className={styles.spinner}></div></div> }
                
                { currentChat && <MessageList /> }
            </div>

            <ChatInput
                onSendMessage={handleSendMessage}
                isLoading={isChatLoading} 
                inputRef={inputRef}
                isSidebarOpen={isSidebarOpen}
            />

            {/* <button className={styles.bottomRightSettingsButton} onClick={() => setSettingsModalOpen(true)} aria-label="Open Settings">
                <SettingsIcon />
            </button> */}

            {/* <SettingsModal
                isOpen={isSettingsModalOpen}
                onClose={() => setSettingsModalOpen(false)}
            /> */}
        </div>
    );
}

export default App;