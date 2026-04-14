import { useEffect } from 'react';
import { useColorScheme } from '@mui/material/styles';

type Contrast = 'normal' | 'mc' | 'hc';

const ThemeLoader = () => {
  const { mode } = useColorScheme();
  const contrast: Contrast = 'normal';

  useEffect(() => {
    // Remove previous theme classes
    document.documentElement.classList.remove('dark', 'light', 'dark-hc', 'dark-mc', 'light-hc', 'light-mc');

    const themeKey = `${mode}-${contrast}`;
    let className: string;

    switch (themeKey) {
      case 'dark-hc':
        import('./flow-ai-color-sheme/css/dark-hc.css');
        className = 'dark-hc';
        break;
      case 'dark-mc':
        import('./flow-ai-color-sheme/css/dark-mc.css');
        className = 'dark-mc';
        break;
      case 'light-hc':
        import('./flow-ai-color-sheme/css/light-hc.css');
        className = 'light-hc';
        break;
      case 'light-mc':
        import('./flow-ai-color-sheme/css/light-mc.css');
        className = 'light-mc';
        break;
      case 'light-normal':
        import('./flow-ai-color-sheme/css/light.css');
        className = 'light';
        break;
      default: // 'dark-normal'
        import('./flow-ai-color-sheme/css/dark.css');
        className = 'dark';
        break;
    }

    document.documentElement.classList.add(className);
  }, [mode, contrast]);

  return null;
};

export default ThemeLoader;