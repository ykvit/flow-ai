import React from 'react';
import styles from './Sidebar.module.css';
import TrashIcon from '../../assets/delete-icon.svg?react';
import DatabaseIcon from '../../assets/database-icon.svg?react'; 
import SearchIcon from '../../assets/search-chat-button.svg?react';
import Archive from '../../assets/archive-icon.svg?react';


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
}

const Sidebar: React.FC<SidebarProps> = ({
    isOpen,
    savedChats,
    activeChatId,
    onSelectChat,
    onDeleteChat,
}) => {

    const handleDelete = (e: React.MouseEvent, chatId: string) => {
        e.stopPropagation();
        onDeleteChat(chatId);
    };

  return (
    <div className={`${styles.sidebarContainer} ${isOpen ? styles.visible : ''}`}>

        <div className={styles.sidebarActions}>
             <button className={styles.sidebaricons}>
             <DatabaseIcon />
             </button>
             <button className={styles.sidebaricons} aria-label="archive">
               <Archive />
             </button>  
             <button className={styles.sidebaricons} aria-label="Search">
               <SearchIcon />
             </button>
        </div>

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
                        <button
                           className={styles.chatDeleteButton}
                           onClick={(e) => handleDelete(e, chat.id)}
                           aria-label={`Delete chat ${chat.title || 'Untitled Chat'}`}
                           title={`Delete chat ${chat.title || 'Untitled Chat'}`}
                        >
                           <TrashIcon />
                        </button>
                    </li>
                ))}
            </ul>
        </div>

    </div> 
  );
};

export default Sidebar;