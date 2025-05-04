import React from 'react';
import styles from './Sidebar.module.css';

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({ isOpen }) => {
  return (
    <div className={`${styles.sidebarContainer} ${isOpen ? styles.visible : ''}`}>
      <div className={styles.content}>
      </div>
    </div>
  );
};

export default Sidebar;