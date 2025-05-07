import React from 'react';
import styles from '../SettingsContent.module.css';

const AboutSettings: React.FC = () => {
  const appVersion = "0.0.1"; 

  return (
    <div>
       <div className={styles.settingSection}>
            <div className={styles.settingItem}>
                 <label className={styles.settingLabel}>
                     Application Version
                 </label>
                 <div className={styles.settingControl}>
                     <span style={{ color: 'rgba(255,255,255,0.8)' }}>{appVersion}</span>
                 </div>
            </div>
       </div>

       <div className={styles.settingSection}>
            <div className={styles.settingItem}>
                 <label className={styles.settingLabel}>
                     Repository / Website
                 </label>
                 <div className={styles.settingControl}>
                    <a href="YOUR_REPO_LINK_HERE" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--gradient-middle)', textDecoration: 'underline' }}>
                       Visit GitHub
                    </a>
                 </div>
            </div>
       </div>

        <div className={styles.settingSection}>
            <div className={styles.settingItem}>
                 <label className={styles.settingLabel}>
                     Acknowledgements
                 </label>
            </div>
            <p style={{ fontSize: '0.85em', color: 'rgba(255,255,255,0.7)', marginTop: '10px', lineHeight: '1.5' }}>
                This application uses Ollama for local AI model execution. <br />
                Built with React, Vite, and TypeScript. <br />
            </p>
       </div>

       <div className={styles.settingSection}>
            <div className={styles.settingItem}>
                 <label className={styles.settingLabel}>
                     Licenses
                 </label>
            </div>
            <p style={{ fontSize: '0.85em', color: 'rgba(255,255,255,0.7)', marginTop: '10px', lineHeight: '1.5' }}>
                MIT License (Application Code) <br />
                Refer to individual component licenses for more details.
            </p>
       </div>
    </div>
  );
};

export default AboutSettings;