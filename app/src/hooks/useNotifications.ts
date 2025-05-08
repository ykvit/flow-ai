import { useState, useEffect, useCallback } from 'react';

const NOTIFICATIONS_ENABLED_KEY = 'flowai_notificationsEnabled';

export function useNotifications() {
  const [permission, setPermission] = useState<NotificationPermission>(Notification.permission);
  const [isEnabledInSettings, setIsEnabledInSettings] = useState<boolean>(() => {
    const saved = localStorage.getItem(NOTIFICATIONS_ENABLED_KEY);
    try {
      return saved !== null ? JSON.parse(saved) && Notification.permission === 'granted' : false; 
    } catch {
      return false; 
    }
  });


  const requestPermission = useCallback(async (): Promise<boolean> => {
    if (!('Notification' in window)) {
        console.error('This browser does not support desktop notification');
        alert('Your browser does not support notifications.');
        return false;
    }

    if (permission === 'granted') {
        console.log('Notification permission already granted.');
        return true;
    }

    if (permission === 'denied') {
        console.log('Notification permission was previously denied.');
        alert('Notifications are blocked in your browser settings. Please enable them there.');
        return false;
    }

    try {
        const currentPermission = await Notification.requestPermission();
        setPermission(currentPermission);
        if (currentPermission === 'granted') {
            console.log('Notification permission granted.');
            setIsEnabledInSettings(true); 
            localStorage.setItem(NOTIFICATIONS_ENABLED_KEY, JSON.stringify(true));
            return true;
        } else {
            console.log('Notification permission denied by user.');
            setIsEnabledInSettings(false);
            localStorage.setItem(NOTIFICATIONS_ENABLED_KEY, JSON.stringify(false));
            return false;
        }
    } catch (error) {
        console.error('Error requesting notification permission:', error);
        return false;
    }
  }, [permission]);

  const showNotification = useCallback((title: string, options?: NotificationOptions & { chatId?: string }) => {

    if (!isEnabledInSettings || permission !== 'granted' || !document.hidden) {

        return; 
    }

    console.log('Showing notification:', title, options);
    
    const tag = options?.chatId ? `flow-ai-chat-${options.chatId}` : undefined;
    const notificationOptions = { ...options, tag };

    const notification = new Notification(title, notificationOptions);

    notification.onclick = () => {
        window.focus();
        notification.close();
    };
  }, [isEnabledInSettings, permission]); 

  const toggleNotificationsSetting = useCallback((enable: boolean) => {
     if (enable && permission !== 'granted') {
         console.warn('Cannot enable notifications setting because browser permission is not granted.');
         requestPermission().then(granted => {
             if (granted) {
                 setIsEnabledInSettings(true);
                 localStorage.setItem(NOTIFICATIONS_ENABLED_KEY, JSON.stringify(true));
             } else {
                  setIsEnabledInSettings(false);
                  localStorage.setItem(NOTIFICATIONS_ENABLED_KEY, JSON.stringify(false));
             }
         });
     } else {
        setIsEnabledInSettings(enable);
        localStorage.setItem(NOTIFICATIONS_ENABLED_KEY, JSON.stringify(enable));
     }
  }, [permission, requestPermission]);


  useEffect(() => {
      const handleStorageChange = (event: StorageEvent) => {
          if (event.key === NOTIFICATIONS_ENABLED_KEY) {
               try {
                 const newValue = event.newValue ? JSON.parse(event.newValue) : false;
                 if (typeof newValue === 'boolean') {
                     setIsEnabledInSettings(newValue && Notification.permission === 'granted');
                 }
               } catch {
                   setIsEnabledInSettings(false);
               }
          }
      };
      window.addEventListener('storage', handleStorageChange);
      return () => window.removeEventListener('storage', handleStorageChange);
  }, []);

  return { 
    isNotificationEnabled: isEnabledInSettings,
    notificationPermission: permission,
    requestPermission,     
    showNotification,      
    toggleNotificationsSetting   
   };
}