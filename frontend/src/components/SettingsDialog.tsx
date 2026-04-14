import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  IconButton,
  Tabs,
  Tab,
  Divider,
  List,
  ListItem,
  ListItemText,
  LinearProgress,
  InputAdornment,
  CircularProgress,
  Typography,
  Box,
  Snackbar,
  Alert,
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import DeleteIcon from '@mui/icons-material/Delete';
import DownloadIcon from '@mui/icons-material/Download';
import SearchIcon from '@mui/icons-material/Search';
import { useSettingsStore } from '../stores/settings';
import { useModelsStore } from '../stores/models';

interface SettingsDialogProps {
  open: boolean;
  onClose: () => void;
}

export default function SettingsDialog({ open, onClose }: SettingsDialogProps) {
  const { settings, updateSettings, isSuccess, resetSuccess } = useSettingsStore();
  const { models, pullStatus, pullModel, deleteModel, isLoading: isModelsLoading } = useModelsStore();

  const [tabValue, setTabValue] = useState(0);
  const [mainModel, setMainModel] = useState(settings.main_model);
  const [supportModel, setSupportModel] = useState(settings.support_model);
  const [systemPrompt, setSystemPrompt] = useState(settings.system_prompt);
  const [newModelName, setNewModelName] = useState('');

  // Sync form with store when dialog opens
  useEffect(() => {
    if (open) {
      setMainModel(settings.main_model);
      setSupportModel(settings.support_model);
      setSystemPrompt(settings.system_prompt);
    }
  }, [open, settings]);

  const handleSave = async () => {
    await updateSettings({
      main_model: mainModel,
      support_model: supportModel,
      system_prompt: systemPrompt,
    });
  };

  const handlePullModel = async () => {
    if (!newModelName.trim()) return;
    await pullModel({ name: newModelName.trim() });
    setNewModelName('');
  };

  const handleDeleteModel = async (name: string) => {
    if (window.confirm(`Are you sure you want to delete model ${name}?`)) {
      await deleteModel({ name });
    }
  };

  const handleClose = () => {
    resetSuccess();
    onClose();
  };

  const handleTabChange = (_: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  return (
    <>
      <Dialog
        id="settings-dialog"
        open={open}
        onClose={handleClose}
        maxWidth="sm"
        fullWidth
        slotProps={{
          paper: {
            sx: {
              borderRadius: '28px',
              p: 1,
            },
          }
        }}
      >
        <DialogTitle
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            pb: 1,
          }}
        >
          <Typography variant="h6" sx={{ fontWeight: 600 }}>
            Settings
          </Typography>
          <IconButton onClick={handleClose} sx={{ color: 'text.secondary' }}>
            <CloseIcon />
          </IconButton>
        </DialogTitle>

        <DialogContent sx={{ pt: 1, px: 0 }}>
          <Tabs 
            value={tabValue} 
            onChange={handleTabChange} 
            variant="fullWidth"
            sx={{ 
              borderBottom: 1, 
              borderColor: 'divider',
              mb: 2,
              px: 2,
              '& .MuiTab-root': { fontWeight: 600, textTransform: 'none' } 
            }}
          >
            <Tab label="General" />
            <Tab label="Models" />
          </Tabs>

          {tabValue === 0 && (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3, p: 3, pt: 1 }}>
              {/* Main Model */}
              <FormControl fullWidth>
                <InputLabel id="main-model-label">Main Model</InputLabel>
                <Select
                  id="main-model-select"
                  labelId="main-model-label"
                  value={mainModel}
                  label="Main Model"
                  onChange={(e) => setMainModel(e.target.value)}
                  sx={{ borderRadius: '12px' }}
                >
                  {models.map((model) => (
                    <MenuItem key={model.name} value={model.name}>
                      {model.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              {/* Support Model */}
              <FormControl fullWidth>
                <InputLabel id="support-model-label">Support Model</InputLabel>
                <Select
                  id="support-model-select"
                  labelId="support-model-label"
                  value={supportModel}
                  label="Support Model"
                  onChange={(e) => setSupportModel(e.target.value)}
                  sx={{ borderRadius: '12px' }}
                >
                  {models.map((model) => (
                    <MenuItem key={model.name} value={model.name}>
                      {model.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              {/* System Prompt */}
              <TextField
                id="system-prompt-input"
                label="System Prompt"
                multiline
                rows={4}
                value={systemPrompt}
                onChange={(e) => setSystemPrompt(e.target.value)}
                fullWidth
                placeholder="You are a helpful assistant..."
                sx={{
                  '& .MuiOutlinedInput-root': {
                    borderRadius: '16px',
                  },
                }}
              />
            </Box>
          )}

          {tabValue === 1 && (
            <Box sx={{ p: 3, pt: 1 }}>
              <Typography variant="subtitle2" sx={{ mb: 2, fontWeight: 600 }}>Pull New Model</Typography>
              <Box sx={{ display: 'flex', gap: 1, mb: 3 }}>
                <TextField
                  fullWidth
                  size="small"
                  placeholder="model-name (e.g. llama3)"
                  value={newModelName}
                  disabled={!!pullStatus}
                  onChange={(e) => setNewModelName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') handlePullModel();
                  }}
                  slotProps={{
                    input: {
                      startAdornment: (
                        <InputAdornment position="start">
                          <SearchIcon fontSize="small" />
                        </InputAdornment>
                      ),
                      sx: { borderRadius: '12px' }
                    }
                  }}
                />
                <Button 
                  variant="contained" 
                  onClick={handlePullModel}
                  disabled={!newModelName.trim() || !!pullStatus}
                  sx={{ borderRadius: '12px', minWidth: '100px' }}
                >
                  {pullStatus ? <CircularProgress size={20} color="inherit" /> : 'Pull'}
                </Button>
              </Box>

              {pullStatus && (
                <Box sx={{ mb: 3, p: 2, bgcolor: 'action.hover', borderRadius: '12px' }}>
                  <Typography variant="caption" sx={{ display: 'block', mb: 1, fontWeight: 600 }}>
                    {pullStatus.status} {pullStatus.completed != null && pullStatus.total != null && pullStatus.total > 0 ? `${Math.round((pullStatus.completed / pullStatus.total) * 100)}%` : ''}
                  </Typography>
                  {pullStatus.total != null && pullStatus.total > 0 && (
                    <LinearProgress 
                      variant="determinate" 
                      value={pullStatus.completed != null ? (pullStatus.completed / pullStatus.total) * 100 : 0} 
                      sx={{ height: 8, borderRadius: 4 }}
                    />
                  )}
                </Box>
              )}

              <Typography variant="subtitle2" sx={{ mb: 1, fontWeight: 600 }}>Installed Models</Typography>
              <List sx={{ bgcolor: 'action.hover', borderRadius: '12px', overflow: 'hidden' }}>
                {models.length === 0 ? (
                  <ListItem><ListItemText primary="No models found" /></ListItem>
                ) : (
                  models.map((model, index) => (
                    <Box key={model.name}>
                      <ListItem
                        secondaryAction={
                          <IconButton edge="end" onClick={() => handleDeleteModel(model.name)} disabled={!!pullStatus}>
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        }
                      >
                        <ListItemText 
                          primary={
                            <Typography variant="body2" sx={{ fontWeight: 500 }}>
                              {model.name}
                            </Typography>
                          } 
                          secondary={model.size ? `${(model.size / (1024 * 1024 * 1024)).toFixed(2)} GB` : ''} 
                        />
                      </ListItem>
                      {index < models.length - 1 && <Divider />}
                    </Box>
                  ))
                )}
              </List>
            </Box>
          )}
        </DialogContent>

        <DialogActions sx={{ px: 3, pb: 2, gap: 1 }}>
          <Button
            onClick={handleClose}
            variant="text"
            sx={{ borderRadius: '20px', px: 3 }}
          >
            Cancel
          </Button>
          <Button
            id="save-settings-button"
            onClick={handleSave}
            variant="contained"
            sx={{
              borderRadius: '20px',
              px: 3,
              boxShadow: 'none',
              '&:hover': {
                boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
              },
            }}
          >
            Save
          </Button>
        </DialogActions>
      </Dialog>

      {/* Success Snackbar */}
      <Snackbar
        open={isSuccess}
        autoHideDuration={3000}
        onClose={resetSuccess}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          onClose={resetSuccess}
          severity="success"
          variant="filled"
          sx={{ borderRadius: '12px' }}
        >
          Settings saved successfully
        </Alert>
      </Snackbar>
    </>
  );
}
