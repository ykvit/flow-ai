import React from 'react';
import styles from './Header.module.css';
import BoxIcon from '../../assets/side-field-button.svg?react';
import DatabaseIcon from '../../assets/database-icon.svg?react'; 
import SearchIcon from '../../assets/search-chat-button.svg?react'; 

interface HeaderProps {
  onToggleSidebar: () => void;
  isSidebarOpen: boolean;
}

const Header: React.FC<HeaderProps> = ({ onToggleSidebar, isSidebarOpen }) => {
  return (
    <>
      {/*left part*/}
      <div className={`${styles.topIcons} ${styles.topLeftIcons}`}>
        <button onClick={onToggleSidebar} className={styles.sidebarToggleButton} aria-label="Toggle Sidebar">
          <BoxIcon />
        </button>

        {/* icons sidebar */}
        {isSidebarOpen && (
          <>
            <button className={styles.dataCollectionButton}>
              <span className={styles.buttonIcon}><DatabaseIcon /> </span>
              <span>Data collection</span>
            </button>

            <button className={styles.searchButton} aria-label="Search">
              <span className={styles.buttonIcon}>
              <SearchIcon />
              </span>
            </button>
          </>
        )}
      </div>

      {/* right part */}
      <div className={`${styles.topIcons} ${styles.topRightIcons}`}>
        <span>⏳ Temporary</span>
        <span>A</span>
      </div>
    </>
  );
};

export default Header;