import { 
  Button, 
  Box, 
  Typography, 
  AppBar, 
  Toolbar,
  IconButton 
} from '@mui/material';
import { useColorScheme } from '@mui/material/styles';
import Brightness4Icon from '@mui/icons-material/Brightness4';
import Brightness7Icon from '@mui/icons-material/Brightness7';
import ThemeLoader from './theme/ThemeLoader'; 

function ModeSwitcher() {
  const { mode, setMode } = useColorScheme();
  return (
    <IconButton onClick={() => setMode(mode === 'light' ? 'dark' : 'light')} color="inherit">
      {mode === 'dark' ? <Brightness7Icon /> : <Brightness4Icon />}
    </IconButton>
  );
}


function App() {
  return (
    <>
      <ThemeLoader /> 
      
      <AppBar position="static" enableColorOnDark>
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            Flow-AI
          </Typography>
          <ModeSwitcher />
        </Toolbar>
      </AppBar>
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 2,
          p: 4,
          bgcolor: 'var(--md-sys-color-background)',
          color: 'var(--md-sys-color-on-background)',
        }}
      >
        <Typography variant="h4" component="h1">
          Вітаємо в Material Design 3!
        </Typography>
        <Typography>
          our cheme
        </Typography>
        <Box sx={{ display: 'flex', gap: 2 }}>
          <Button variant="contained">main btn</Button>
          <Button variant="outlined">second btn</Button>
          <Button variant="text">text btn</Button>
        </Box>
      </Box>
    </>
  );
}

export default App;