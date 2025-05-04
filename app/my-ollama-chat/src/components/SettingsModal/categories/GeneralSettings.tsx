import React from 'react';
import styles from '../SettingsContent.module.css';

const GeneralSettings: React.FC = () => {
  return (
    <div>
      {/* Language & Startup Behavior Section */}
      <div className={styles.settingSection}>
        {/* Language */}
        <div className={styles.settingItem}>
          <label className={styles.settingLabel} htmlFor="language-select">
            Language
          </label>
          <div className={styles.settingControl}>
            <select id="language-select" defaultValue="en">
              <option value="en">English</option>
              <option value="uk">Українська</option>
            </select>
          </div>
        </div>
        {/* Startup Behavior */}
        <div className={styles.settingItem}>
           <label className={styles.settingLabel} htmlFor="startup-behavior">
             Startup Behavior
           </label>
           <div className={styles.settingControl}>
             <select id="startup-behavior" defaultValue="last">
               <option value="last">Open last active chat</option>
               <option value="new">Always create new chat</option>
             </select>
           </div>
         </div>
      </div>

      {/* Default System Prompt Section */}
      <div className={styles.settingSection}>
         <label className={styles.settingLabel} htmlFor="system-prompt" style={{ marginBottom: '10px', display: 'block' }}>
             Default System Prompt
         </label>
         <div className={`${styles.settingControl} ${styles.fullWidthTextarea}`}>
             <textarea
                id="system-prompt"
                rows={3}
                placeholder="e.g., You are a helpful and friendly assistant."
             />
         </div>
      </div>

      {/* Notifications & Theme Section */}
      <div className={styles.settingSection}>
         {/* Notifications */}
         <div className={styles.settingItem}>
           <label className={styles.settingLabel} htmlFor="notifications">
             Notifications
           </label>
           <div className={styles.settingControl}>
             <label htmlFor="notifications" style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }}>
               <input type="checkbox" id="notifications" defaultChecked style={{ marginRight: '8px' }}/> Enable
             </label>
           </div>
         </div>
         {/* Theme */}
         <div className={styles.settingItem}>
           <label className={styles.settingLabel} htmlFor="theme-select">
             Theme
           </label>
           <div className={styles.settingControl}>
             <select id="theme-select" defaultValue="dark">
               <option value="dark">Dark</option>
               <option value="light">Light</option>
             </select>
           </div>
         </div>
      </div>

    </div>
  );
};

export default GeneralSettings;