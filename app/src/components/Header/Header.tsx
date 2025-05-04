import React from 'react';
import styles from './Header.module.css';
import BoxIcon from '../../assets/side-field-button.svg?react'; 
import PlusIcon from '../../assets/new-chat-icon.svg?react'; 

interface HeaderProps {
  onToggleSidebar: () => void;
  isSidebarOpen: boolean;
  onCreateNewChat: () => void;
}

const Header: React.FC<HeaderProps> = ({
  onToggleSidebar,
  isSidebarOpen,
  onCreateNewChat,
}) => {
  return (
    <>
      <div className={styles.sidebarToggleContainer}>
        <button onClick={onToggleSidebar} className={styles.sidebarToggleButton} aria-label="Toggle Sidebar">
          <BoxIcon />
        </button>
      </div>
      <div className={`${styles.mainHeaderControls} ${isSidebarOpen ? styles.openState : ''}`}>
         <button onClick={onCreateNewChat} className={styles.newChatHeaderButton} aria-label="New Chat" title="New Chat">
           <PlusIcon />
         </button>
      </div>

      <div className={`${styles.topIcons} ${styles.topRightIcons}`}>
        <span>⏳ Temporary</span>
        <span>A</span>
      </div>
    </>
  );
};

export default Header;