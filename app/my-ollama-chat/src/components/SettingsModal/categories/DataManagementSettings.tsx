import React from 'react';
import styles from '../SettingsContent.module.css';

const DataManagementSettings: React.FC = () => {

  const handleClearChats = () => {
    //  need to add confirmation logic!!
    if (window.confirm("Are you sure you want to delete ALL chat history? This action cannot be undone.")) {
        console.log("Clearing all chats...");
        // Here will be a call to the cleanup functionQ
    }
  };

  return (
    <div>
       <div className={styles.settingSection}>
            <div className={styles.settingItem}>
                 <label className={styles.settingLabel}>
                     Clear All Chats
                 </label>
                 <div className={styles.settingControl}>
                    <button
                        onClick={handleClearChats}
                        style={{ background: 'rgba(255,50,50,0.2)', border: '1px solid rgba(255,50,50,0.4)', color: '#ffcccc' }}
                    >
                        Clear History...
                    </button>
                 </div>
            </div>
            <p style={{ fontSize: '0.8em', color: 'rgba(255,255,255,0.5)', marginTop: '5px' }}>
                Permanently removes all conversation history from this application.
            </p>
       </div>

        <div className={styles.settingSection}>
            <div className={styles.settingItem}>
                 <label className={styles.settingLabel}>
                     Export Chats
                 </label>
                 <div className={styles.settingControl}>
                    <button>Export All (JSON)</button> {/* add! */}
                 </div>
            </div>
             <p style={{ fontSize: '0.8em', color: 'rgba(255,255,255,0.5)', marginTop: '5px' }}>
                Save your chat history to a file.
            </p>
       </div>

        <div className={styles.settingSection}>
            <div className={styles.settingItem}>
                 <label className={styles.settingLabel}>
                     Import Chats
                 </label>
                 <div className={styles.settingControl}>
                    <button>Import from File...</button> {/* add!*/}
                 </div>
            </div>
             <p style={{ fontSize: '0.8em', color: 'rgba(255,255,255,0.5)', marginTop: '5px' }}>
                Load chat history from a previously exported file.
            </p>
       </div>

        {/* <div className={styles.settingSection}>
            <div className={styles.settingItem}>
                 <label className={styles.settingLabel}>
                     Clear Cache
                 </label>
                 <div className={styles.settingControl}>
                    <button>Clear Cache</button> }
                 {/* </div>
            </div>
        </div> */}

    </div>
  );
};

export default DataManagementSettings;