// src/components/SettingsModal/categories/GeneralSettings.tsx
import React, { useState, useEffect } from 'react';
import styles from '../SettingsContent.module.css'; 
import { DEFAULT_SYSTEM_PROMPT } from '../../../config/constants'; 
import { useNotifications } from '../../../hooks/useNotifications'; 
import CustomDropdown from '../../CustomDropdown/CustomDropdown';
import type { DropdownOption } from '../../CustomDropdown/CustomDropdown';

const GeneralSettings: React.FC = () => {
  const { 
    isNotificationEnabled,  
    notificationPermission,  
    toggleNotificationsSetting 
  } = useNotifications();


  const [systemPrompt, setSystemPrompt] = useState(''); 
  const [answerLanguage, setAnswerLanguage] = useState('auto'); 
  const [interfaceLanguage, setInterfaceLanguage] = useState('en');
  const [theme, setTheme] = useState('dark');

  const interfaceLanguageOptions: DropdownOption[] = [
    { value: 'en', label: 'English' },
  ];

  const answerLanguageOptions: DropdownOption[] = [
    { value: 'auto', label: 'Auto' },
    { value: 'en', label: 'English' },
    { value: 'fr', label: 'French' },
    { value: 'ua', label: 'Ukrainian' },
    { value: 'es', label: 'Spanish' }
  ];

  const themeOptions: DropdownOption[] = [
    { value: 'dark', label: 'Dark' },
  ];

  useEffect(() => {
    const savedPrompt = localStorage.getItem('flowai_systemPrompt'); 
    if (savedPrompt !== null) {
      setSystemPrompt(savedPrompt); 
    }


    const savedAnswerLang = localStorage.getItem('flowai_answerLanguage');
    if (savedAnswerLang) { 
      setAnswerLanguage(savedAnswerLang);
    }
    
    const savedInterfaceLang = localStorage.getItem('flowai_interfaceLanguage');
    if (savedInterfaceLang) {
      setInterfaceLanguage(savedInterfaceLang);
    }

    const savedTheme = localStorage.getItem('flowai_theme');
    if (savedTheme) {
      setTheme(savedTheme);
    }
  }, []); 

  const handlePromptChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newPrompt = event.target.value;
    setSystemPrompt(newPrompt);
    localStorage.setItem('flowai_systemPrompt', newPrompt); 
  };

  const handleAnswerLanguageChange = (value: string) => {
    setAnswerLanguage(value);
    localStorage.setItem('flowai_answerLanguage', value);
  };

  const handleInterfaceLanguageChange = (value: string) => {
    setInterfaceLanguage(value);
    localStorage.setItem('flowai_interfaceLanguage', value);
  };

  const handleThemeChange = (value: string) => {
    setTheme(value);
    localStorage.setItem('flowai_theme', value);
  };

  const handleNotificationToggleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const enable = event.target.checked;
    toggleNotificationsSetting(enable); 
  };

  return (
    <> 
      <div className={styles.section}>
        <h3 className={styles.sectionTitle}>System</h3>
        <div className={styles.settingsList}>
          <div className={styles.settingItem}>
            <label className={styles.settingLabel} htmlFor="interface-language">Language</label>
            <div className={styles.settingControl}>
              <CustomDropdown
                id="interface-language"
                options={interfaceLanguageOptions}
                selectedValue={interfaceLanguage}
                onChange={handleInterfaceLanguageChange}
              />
            </div>
          </div>
          
          <div className={styles.settingItem}>
            <label className={styles.settingLabel} htmlFor="answer-language">Answer language</label>
            <div className={styles.settingControl}>
              <CustomDropdown
                id="answer-language"
                options={answerLanguageOptions}
                selectedValue={answerLanguage}
                onChange={handleAnswerLanguageChange}
              />
            </div>
          </div>

          <div className={styles.settingItem}>
            <label className={styles.settingLabel} htmlFor="theme-select">Theme</label>
            <div className={styles.settingControl}>
              <CustomDropdown
                id="theme-select"
                options={themeOptions}
                selectedValue={theme}
                onChange={handleThemeChange}
                disabled={true} 
              />
            </div>
          </div>

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
      </div>

      <div className={styles.section}>
        <h3 className={styles.sectionTitle}>Default System Prompt</h3>
        <div className={styles.settingsList}> 
             <textarea
                id="system-prompt"
                className={styles.styledTextarea} 
                rows={5} 
                placeholder={DEFAULT_SYSTEM_PROMPT} 
                value={systemPrompt} 
                onChange={handlePromptChange} 
             />
        </div>
      </div>
    </>
  );
};

export default GeneralSettings;