import React from 'react';
import styles from '../SettingsContent.module.css';

const ConnectionsSettings: React.FC = () => {
  return (
    <div>
       {/* Ollama API Endpoint */}
       <div className={styles.settingSection}>
            <div className={styles.settingItem}>
                 <label className={styles.settingLabel} htmlFor="ollama-endpoint">
                     Ollama API Endpoint
                 </label>
            </div>
           <div className={styles.settingControl} style={{ display: 'flex', gap: '10px', marginTop: '10px' }}>
                <input
                    type="url"
                    id="ollama-endpoint"
                    defaultValue="http://localhost:11434"
                    style={{ flexGrow: 1 }} 
                    placeholder="http://127.0.0.1:11434"
                />
                <button>Test Connection</button>
           </div>
       </div>

    </div>
  );
};

export default ConnectionsSettings;