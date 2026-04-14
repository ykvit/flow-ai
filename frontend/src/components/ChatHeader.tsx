import {
  Box,
  IconButton,
  Typography,
  Select,
  MenuItem,
  FormControl,
  Tooltip,
} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import MenuOpenIcon from '@mui/icons-material/MenuOpen';
import Brightness4Icon from '@mui/icons-material/Brightness4';
import Brightness7Icon from '@mui/icons-material/Brightness7';
import { useColorScheme } from '@mui/material/styles';
import type { Model } from '../types/models';

interface ChatHeaderProps {
  title: string;
  onToggleSidebar: () => void;
  sidebarOpen: boolean;
}

export default function ChatHeader({
  title,
  onToggleSidebar,
  sidebarOpen,
}: ChatHeaderProps) {
  const { mode, setMode } = useColorScheme();

  return (
    <Box
      component="header"
      className={`glass ${mode === 'dark' ? 'glass-dark' : 'glass-light'}`}
      sx={{
        display: 'flex',
        alignItems: 'center',
        gap: 2,
        px: 3,
        py: 1,
        borderBottom: '1px solid',
        borderColor: 'divider',
        minHeight: 64,
        flexShrink: 0,
        zIndex: 10,
        position: 'sticky',
        top: 0,
      }}
    >
      <Tooltip title={sidebarOpen ? "Close Sidebar" : "Open Sidebar"}>
        <IconButton
          id="toggle-sidebar-button"
          onClick={onToggleSidebar}
          sx={{ color: 'text.secondary', p: 1 }}
        >
          {sidebarOpen ? <MenuOpenIcon fontSize="small" /> : <MenuIcon fontSize="small" />}
        </IconButton>
      </Tooltip>

      <Typography
        variant="h6"
        sx={{
          fontWeight: 600,
          flex: 1,
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
          color: 'text.primary',
          fontSize: '1rem',
        }}
      >
        {title}
      </Typography>

      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
        <Tooltip title={mode === 'dark' ? "Switch to Light Mode" : "Switch to Dark Mode"}>
          <IconButton
            id="theme-toggle-button"
            onClick={() => setMode(mode === 'light' ? 'dark' : 'light')}
            sx={{ color: 'text.secondary', p: 1 }}
          >
            {mode === 'dark' ? <Brightness7Icon fontSize="small" /> : <Brightness4Icon fontSize="small" />}
          </IconButton>
        </Tooltip>
      </Box>
    </Box>
  );
}
