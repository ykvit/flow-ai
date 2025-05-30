import { useState, useEffect, useCallback } from 'react';

const NOTIFICATIONS_ENABLED_KEY = 'flowai_notificationsEnabled';
const BROWSER_PERMISSION_KEY = 'flowai_browserNotificationPermission'; 

export function useNotifications() {
  const [browserPermission, setBrowserPermission] = useState<NotificationPermission>(() => {
    if ('Notification' in window) {
      return Notification.permission;
    }
    return 'default'; 
  });

  const [appSettingEnabled, setAppSettingEnabled] = useState<boolean>(() => {
    const saved = localStorage.getItem(NOTIFICATIONS_ENABLED_KEY);
    return saved !== null ? JSON.parse(saved) : false; 
  });

  useEffect(() => {
    if (!('Notification' in window) || !('permissions' in navigator)) {
      return;
    }


    navigator.permissions.query({ name: 'notifications' }).then((permissionStatus) => {
      setBrowserPermission(permissionStatus.state as NotificationPermission); 
      permissionStatus.onchange = () => { 
        setBrowserPermission(permissionStatus.state as NotificationPermission);

      };
    }).catch(error => {
        console.warn("Could not subscribe to notification permission changes:", error);

    });

  }, []);


  const requestPermission = useCallback(async (): Promise<boolean> => {
    if (!('Notification' in window)) {
      console.error('This browser does not support desktop notification');
      return false;
    }

    if (browserPermission === 'granted') {
      console.log('Notification permission already granted.');
      return true;
    }

    if (browserPermission === 'denied') {
      console.log('Notification permission was previously denied by user.');
      alert('Notifications are blocked. Please check your browser settings for this site to enable them.');
      return false;
    }

    try {
      const currentPermission = await Notification.requestPermission();
      setBrowserPermission(currentPermission); 
      
      if (currentPermission === 'granted') {
        console.log('Notification permission granted by user.');
        return true;
      } else {
        console.log('Notification permission denied by user.');
        setAppSettingEnabled(false);
        localStorage.setItem(NOTIFICATIONS_ENABLED_KEY, JSON.stringify(false));
        return false;
      }
    } catch (error) {
      console.error('Error requesting notification permission:', error);
      return false;
    }
  }, [browserPermission]);

  const toggleNotificationsSetting = useCallback(async (enable: boolean) => {
    setAppSettingEnabled(enable);
    localStorage.setItem(NOTIFICATIONS_ENABLED_KEY, JSON.stringify(enable));

    if (enable && browserPermission !== 'granted') {
      const permissionGranted = await requestPermission();
      if (!permissionGranted) {
        setAppSettingEnabled(false);
        localStorage.setItem(NOTIFICATIONS_ENABLED_KEY, JSON.stringify(false));
      }
    }
  }, [browserPermission, requestPermission]);

  const isEffectivelyEnabled = appSettingEnabled && browserPermission === 'granted';

  const showNotification = useCallback((title: string, options?: NotificationOptions & { chatId?: string }) => {
    if (!isEffectivelyEnabled || !document.hidden) {
      if (document.hidden === false) console.log("Notification not shown because tab is active.");
      if (!isEffectivelyEnabled) console.log(`Notification not shown. AppSetting: ${appSettingEnabled}, BrowserPerm: ${browserPermission}`);
      return;
    }
    if (!('Notification' in window)) return; 

    console.log('Attempting to show notification:', title, options);
    
    const tag = options?.chatId ? `flow-ai-chat-${options.chatId}` : `flow-ai-notification-${Date.now()}`; // Унікальний тег, якщо chatId немає
    const notificationOptions = { 
      body: options?.body, 
      icon: options?.icon, 
      badge: options?.badge, 
      renotify: !!tag, 
      tag: tag,
      ...options 
    };

    try {
      const notification = new Notification(title, notificationOptions);
      notification.onclick = () => {
        window.focus();
        notification.close();
      };
      notification.onerror = (err) => {
        console.error("Notification API error: ", err);
      };
    } catch (e) {
      console.error("Failed to create notification:", e);
    }
  }, [isEffectivelyEnabled, appSettingEnabled, browserPermission]); 

  useEffect(() => {
    const handleStorageChange = (event: StorageEvent) => {
      if (event.key === NOTIFICATIONS_ENABLED_KEY) {
        try {
          const newValue = event.newValue ? JSON.parse(event.newValue) : false;
          if (typeof newValue === 'boolean') {
            setAppSettingEnabled(newValue);
          }
        } catch {
          setAppSettingEnabled(false);
        }
      }
    };
    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, []);

  return {
    isNotificationEnabled: isEffectivelyEnabled, 
    notificationPermission: browserPermission,
    appSettingEnabled: appSettingEnabled,
    requestPermission,
    showNotification,
    toggleNotificationsSetting
  };
}