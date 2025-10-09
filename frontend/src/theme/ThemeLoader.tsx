
import { useEffect } from 'react';
import { useColorScheme } from '@mui/material/styles';

const basePath = './color-schemes/'; 

type Contrast = 'normal' | 'mc' | 'hc';

const ThemeLoader = () => {
  const { mode } = useColorScheme();
  const contrast: Contrast = 'normal'; 

  useEffect(() => {
    switch (`${mode}-${contrast}`) {
      case 'dark-hc':
        import(`${basePath}dark-hc.css`);
        break;
      case 'dark-mc':
        import(`${basePath}dark-mc.css`);
        break;
      case 'light-hc':
        import(`${basePath}light-hc.css`);
        break;
      case 'light-mc':
        import(`${basePath}light-mc.css`);
        break;
      case 'light-normal':
        import(`${basePath}light.css`);
        break;
      default: // 'dark-normal'
        import(`${basePath}dark.css`);
        break;
    }
  }, [mode, contrast]);

  return null;
};

export default ThemeLoader;