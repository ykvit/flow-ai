import { useState, useCallback, useRef, useEffect } from 'react';
import { Box, IconButton, TextField, Typography } from '@mui/material';
import SendIcon from '@mui/icons-material/Send';
import StopCircleIcon from '@mui/icons-material/StopCircle';

interface MessageInputProps {
  onSend: (content: string) => void;
  disabled: boolean;
  isStreaming: boolean;
}

export default function MessageInput({ onSend, disabled, isStreaming }: MessageInputProps) {
  const [value, setValue] = useState('');
  const inputRef = useRef<HTMLTextAreaElement>(null);

  // Auto-focus on mount and after sending
  useEffect(() => {
    if (!isStreaming && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isStreaming]);

  const handleSend = useCallback(() => {
    if (!value.trim() || disabled) return;
    onSend(value);
    setValue('');
  }, [value, disabled, onSend]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    },
    [handleSend]
  );

  return (
    <Box
      sx={{
        px: { xs: 2, sm: 4, md: 6 },
        py: 2.5,
        borderTop: '1px solid',
        borderColor: 'divider',
        bgcolor: 'background.default',
        flexShrink: 0,
      }}
    >
      <Box
        sx={{
          maxWidth: 840,
          mx: 'auto',
          display: 'flex',
          flexDirection: 'column',
          gap: 1.5,
        }}
      >
        <Box
          sx={{
            display: 'flex',
            alignItems: 'flex-end',
            gap: 1.5,
            p: 0.75,
            pr: 1.25,
            borderRadius: '16px',
            bgcolor: 'background.paper',
            border: '1px solid',
            borderColor: 'divider',
            boxShadow: '0 2px 10px rgba(0,0,0,0.03)',
            transition: 'all 0.2s ease',
            '&:focus-within': {
              borderColor: 'primary.main',
              boxShadow: '0 4px 20px rgba(79, 70, 229, 0.08)',
            }
          }}
        >
          <TextField
            id="message-input"
            inputRef={inputRef}
            fullWidth
            multiline
            maxRows={8}
            placeholder="Formulate your request..."
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={isStreaming}
            variant="standard"
            sx={{
              '& .MuiInputBase-root:before, & .MuiInputBase-root:after': {
                display: 'none',
              },
              '& .MuiInputBase-input': {
                py: 1.5,
                px: 2,
                fontSize: '0.925rem',
                lineHeight: 1.6,
              },
            }}
          />
          <IconButton
            id="send-button"
            onClick={handleSend}
            disabled={(!value.trim() && !isStreaming) || (disabled && !isStreaming)}
            sx={{
              bgcolor: value.trim() ? 'primary.main' : 'transparent',
              color: value.trim() ? 'primary.contrastText' : 'text.disabled',
              width: 40,
              height: 40,
              borderRadius: '10px',
              flexShrink: 0,
              mb: 0.5,
              transition: 'all 0.2s ease',
              '&:hover': {
                bgcolor: value.trim() ? 'primary.dark' : 'action.hover',
                transform: value.trim() ? 'scale(1.05)' : 'none',
              },
              '&.Mui-disabled': {
                bgcolor: 'transparent',
                color: 'text.disabled',
                opacity: 0.5,
              },
            }}
          >
            {isStreaming ? (
              <StopCircleIcon sx={{ fontSize: 24, color: 'error.main' }} />
            ) : (
              <SendIcon sx={{ fontSize: 18 }} />
            )}
          </IconButton>
        </Box>
        
        <Box sx={{ display: 'flex', justifyContent: 'center' }}>
          <Typography
            variant="caption"
            sx={{
              fontSize: '0.65rem',
              color: 'text.secondary',
              opacity: 0.5,
              fontFamily: '"Outfit", sans-serif',
              fontWeight: 500,
              letterSpacing: '0.02em',
            }}
          >
            Shift + Enter for new line • Enter to synthesize
          </Typography>
        </Box>
      </Box>
    </Box>
  );
}
