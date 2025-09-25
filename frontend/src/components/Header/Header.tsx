import React from 'react';
import styles from './Header.module.css';
// import BoxIcon from '../../assets/BoxIcon.svg?react'; 
// import PlusIcon from '../../assets/PlusIcon.svg?react'; 

// import ModelIcon from '../../assets/model-a-icon.svg?react?react';

interface HeaderProps {
  onToggleSidebar: () => void;
  isSidebarOpen: boolean;
  onCreateNewChat: () => void;
  chatTitle: string | null;
}

const Header: React.FC<HeaderProps> = ({
  onToggleSidebar,
  isSidebarOpen,
  onCreateNewChat,
  chatTitle,
}) => {
  const displayTitle = chatTitle || "New Chat";

  return (
    <>
      <div className={styles.sidebarToggleContainerFixed}>
        <button onClick={onToggleSidebar} className={styles.headerButton} aria-label="Toggle Sidebar">
          {/* <BoxIcon /> */}
        </button>
      </div>

      <header className={`${styles.mainHeaderContent} ${isSidebarOpen ? styles.sidebarOpenEffect : ''}`}>
        <div className={styles.leftControlsInHeader}>
          <button onClick={onCreateNewChat} className={styles.headerButton} aria-label="New Chat" title="New Chat">
            {/* <PlusIcon /> */}
          </button>
        </div>
        
        <div className={styles.chatTitleContainer}>
          <h1 className={styles.chatTitleText}>{displayTitle}</h1>
        </div>

        {/* <div className={styles.rightControls}> ... </div> */}
      </header>
    </>
  );
};

export default Header;