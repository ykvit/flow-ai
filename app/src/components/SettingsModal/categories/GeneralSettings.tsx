import React, { useState, useEffect } from 'react';
import styles from '../SettingsContent.module.css'; 
import { DEFAULT_SYSTEM_PROMPT } from '../../../config/constants'; 
import { useNotifications } from '../../../hooks/useNotifications'; 

const GeneralSettings: React.FC = () => {
  const { 
    isNotificationEnabled,  
    notificationPermission,  
    // requestPermission,
    toggleNotificationsSetting 
  } = useNotifications();

  const [systemPrompt, setSystemPrompt] = useState(''); 
  const [answerLanguage, setAnswerLanguage] = useState('auto'); 


  useEffect(() => {
    //  System Prompt
    const savedPrompt = localStorage.getItem('flowai_systemPrompt'); 
    if (savedPrompt !== null) {
      setSystemPrompt(savedPrompt); 
    }

    //  Answer Language
    const savedAnswerLang = localStorage.getItem('flowai_answerLanguage');
    if (savedAnswerLang) { 
      setAnswerLanguage(savedAnswerLang);
    }
    

  }, []); 

  const handlePromptChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newPrompt = event.target.value;
    setSystemPrompt(newPrompt);
    localStorage.setItem('flowai_systemPrompt', newPrompt); 
  };

  const handleAnswerLanguageChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const newLang = event.target.value;
    setAnswerLanguage(newLang);
    localStorage.setItem('flowai_answerLanguage', newLang);
  };


  const handleNotificationToggleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const enable = event.target.checked;
    toggleNotificationsSetting(enable); 
  };


  // --- JSX  ---
  return (
    <div>
      {/* ===  System === */}
      <h3 className={styles.sectionTitle}>System</h3>
      <div className={styles.settingGroup}>
        {/* Language */}
        <div className={styles.settingItem}>
          <label className={styles.settingLabel} htmlFor="language-select">Language</label>
          <div className={styles.settingControl}>
            <select id="language-select" className={styles.styledSelect} defaultValue="en">
              <option value="en">English</option>
            </select>
          </div>
        </div>
        
        {/* Answer Language */}
        <div className={styles.settingItem}>
          <label className={styles.settingLabel} htmlFor="answer-language-select">Answer language</label>
          <div className={styles.settingControl}>
            <select 
              id="answer-language-select" 
              className={styles.styledSelect} 
              value={answerLanguage} 
              onChange={handleAnswerLanguageChange} 
            >
              <option value="auto">Auto</option>
              <option value="en">English</option>
              <option value="fr">French</option>
              {/* ... */}
            </select>
          </div>
        </div>

        {/* Theme */}
        <div className={styles.settingItem}>
          <label className={styles.settingLabel} htmlFor="theme-select">Theme</label>
          <div className={styles.settingControl}>
            <select id="theme-select" className={styles.styledSelect} defaultValue="dark" disabled>
              <option value="dark">Dark</option>
            </select>
          </div>
        </div>

        {/* Notifications */}
        <div className={styles.settingItem}>
          <label className={styles.settingLabel} htmlFor="notifications-toggle">Notifications</label>
          <div className={styles.settingControl}>
            <label className={styles.toggleSwitch} htmlFor="notifications-toggle">
              <input 
                type="checkbox" 
                id="notifications-toggle" 
                checked={isNotificationEnabled}
                onChange={handleNotificationToggleChange} 
                disabled={notificationPermission === 'denied'}
              />
              <span className={styles.toggleSlider}></span>
            </label>
          </div>
        </div>
      </div>

      {/* ===  Default System Prompt === */}
      <h3 className={styles.sectionTitle} style={{marginTop: '30px'}}>Default System Prompt</h3>
      <div className={styles.settingGroup}>
         <textarea
            id="system-prompt"
            className={styles.styledTextarea} 
            rows={4} 
            placeholder={DEFAULT_SYSTEM_PROMPT} 
            value={systemPrompt} 
            onChange={handlePromptChange} 
         />
      </div>
    </div>
  );
};

export default GeneralSettings;