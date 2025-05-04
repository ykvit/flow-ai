import React, { useEffect } from 'react';
import styles from './Modal.module.css';
import CloseIcon from '../../assets/settings/close-icon.svg?react';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
  className?: string;
}

const Modal: React.FC<ModalProps> = ({ isOpen, onClose, title, children, className = '' }) => {
  useEffect(() => {
    const handleEsc = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };
    if (isOpen) {
      document.addEventListener('keydown', handleEsc);
    }
    return () => {
      document.removeEventListener('keydown', handleEsc);
    };
  }, [isOpen, onClose]); 

  if (!isOpen) {
    return null;
  }


  const handleOverlayClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  return (
      <div className={styles.modalOverlay} onClick={handleOverlayClick} role="dialog" aria-modal="true" aria-labelledby={title ? 'modal-title' : undefined}>
        <div className={`${styles.modalContent} ${className}`}>
          <div className={styles.modalHeader}>
            {title && <h2 className={styles.modalTitle} id="modal-title">{title}</h2>}
            <button className={styles.closeButton} onClick={onClose} aria-label="Close modal">
              <CloseIcon />
            </button>
          </div>
          <hr className={styles.headerLine} />
          <div className={styles.modalBody}>
            {children}
          </div>
        </div>
      </div>
    );
  };
  
  export default Modal; 