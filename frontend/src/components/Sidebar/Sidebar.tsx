import React from 'react';
import styles from './Sidebar.module.css';
// import TrashIcon from '../../assets/delete-icon.svg?react?react';
// import EditNameChat from '../../assets/edit-icon.svg?react?react';

interface ChatListItem {
    id: string;
    title: string;
}

interface SidebarProps {
    isOpen: boolean;
    savedChats: ChatListItem[];
    activeChatId: string | null;
    onSelectChat: (chatId: string) => void;
    onDeleteChat: (chatId: string) => void;
    onEditChat: (chatId: string) => void;
}

const Sidebar: React.FC<SidebarProps> = ({
    isOpen,
    savedChats,
    activeChatId,
    onSelectChat,
    // onDeleteChat,
    // onEditChat,
}) => {

    // const handleDelete = (e: React.MouseEvent, chatId: string) => {
    //     e.stopPropagation();
    //     onDeleteChat(chatId);
    // };

    // const handleEdit = (e: React.MouseEvent, chatId: string) => {
    //     e.stopPropagation();
    //     onEditChat(chatId);
    // };

    return (
        <div className={`${styles.sidebarContainer} ${isOpen ? styles.visible : ''}`}>
            <div className={styles.content}>
                {savedChats.length === 0 && (
                    <p className={styles.noChats}>No chats yet.</p>
                )}
                <ul className={styles.chatList}>
                    {savedChats.map((chat) => (
                        <li
                            key={chat.id}
                            className={`${styles.chatItem} ${chat.id === activeChatId ? styles.active : ''}`}
                            onClick={() => onSelectChat(chat.id)}
                            role="button"
                            tabIndex={0}
                            onKeyDown={(e) => (e.key === 'Enter' || e.key === ' ') && onSelectChat(chat.id)}
                        >
                            <span className={styles.chatTitle}>
                                {chat.title || "Untitled Chat"}
                            </span>
                            {/* <div className={styles.chatItemActions}>
                                <button
                                    className={styles.chatEditButton}
                                    onClick={(e) => handleEdit(e, chat.id)}
                                    aria-label={`Edit chat ${chat.title || 'Untitled Chat'}`}
                                    title={`Edit chat ${chat.title || 'Untitled Chat'}`}
                                >
                                    <EditNameChat />
                                </button>
                                <button
                                    className={styles.chatDeleteButton}
                                    onClick={(e) => handleDelete(e, chat.id)}
                                    aria-label={`Delete chat ${chat.title || 'Untitled Chat'}`}
                                    title={`Delete chat ${chat.title || 'Untitled Chat'}`}
                                >
                                    <TrashIcon /> */}
                                {/* </button>
                            </div> */}
                        </li>
                    ))}
                </ul>
            </div>
        </div>
    );
};

export default Sidebar;