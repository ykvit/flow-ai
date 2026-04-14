import { extendTheme } from '@mui/material/styles';

const theme = extendTheme({
  colorSchemes: {
    light: {
      palette: {
        primary: {
          main: '#4f46e5', // Deep Indigo
          contrastText: '#ffffff',
        },
        secondary: {
          main: '#64748b', // Slate
          contrastText: '#ffffff',
        },
        background: {
          default: '#f8fafc', // Very subtle gray
          paper: '#ffffff',
        },
        text: {
          primary: '#0f172a', // Deep Slate
          secondary: '#475569',
        },
        action: {
          hover: 'rgba(79, 70, 229, 0.04)',
          selected: 'rgba(79, 70, 229, 0.08)',
        },
        divider: 'rgba(226, 232, 240, 0.8)',
      },
    },
    dark: {
      palette: {
        primary: {
          main: '#818cf8', // Indigo 400
          contrastText: '#ffffff',
        },
        secondary: {
          main: '#94a3b8',
          contrastText: '#ffffff',
        },
        background: {
          default: '#09090b', // Zinc 950
          paper: '#18181b', // Zinc 900
        },
        text: {
          primary: '#fafafa',
          secondary: '#a1a1aa',
        },
        action: {
          hover: 'rgba(129, 140, 248, 0.08)',
          selected: 'rgba(129, 140, 248, 0.12)',
        },
        divider: 'rgba(39, 39, 42, 0.8)',
      },
    },
  },
  typography: {
    fontFamily: '"Inter", system-ui, sans-serif',
    h1: { fontFamily: '"Outfit", sans-serif', fontWeight: 700 },
    h2: { fontFamily: '"Outfit", sans-serif', fontWeight: 700 },
    h3: { fontFamily: '"Outfit", sans-serif', fontWeight: 600 },
    h4: { fontFamily: '"Outfit", sans-serif', fontWeight: 600, letterSpacing: '-0.02em' },
    h5: { fontFamily: '"Outfit", sans-serif', fontWeight: 600 },
    h6: { fontFamily: '"Outfit", sans-serif', fontWeight: 600, letterSpacing: '0.01em' },
    subtitle1: { fontWeight: 500, letterSpacing: '0.01em' },
    subtitle2: { fontWeight: 600, fontSize: '0.875rem' },
    body1: { lineHeight: 1.6, fontSize: '0.925rem' },
    body2: { lineHeight: 1.5, fontSize: '0.875rem' },
    button: { textTransform: 'none', fontWeight: 600, fontFamily: '"Outfit", sans-serif' },
  },
  shape: {
    borderRadius: 12,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: '12px',
          padding: '8px 20px',
          boxShadow: 'none',
          '&:hover': {
            boxShadow: '0 4px 12px rgba(0,0,0,0.08)',
          },
          '&.MuiButton-containedPrimary:hover': {
            backgroundColor: '#4338ca',
          },
        },
      },
    },
    MuiIconButton: {
      styleOverrides: {
        root: {
          borderRadius: '10px',
          transition: 'all 0.2s ease-in-out',
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: '12px',
            backgroundColor: 'transparent',
            transition: 'all 0.2s ease-in-out',
            '&:hover': {
              backgroundColor: 'rgba(0,0,0,0.01)',
            },
            '&.Mui-focused': {
              backgroundColor: 'transparent',
            },
          },
        },
      },
    },
    MuiDialog: {
      styleOverrides: {
        paper: {
          borderRadius: '24px',
          backgroundImage: 'none',
          boxShadow: '0 20px 40px rgba(0,0,0,0.1)',
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          borderRight: '1px solid',
          borderColor: 'divider',
          backgroundImage: 'none',
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
        },
      },
    },
  },
});

export default theme;