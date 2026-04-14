import { useState } from 'react';
import { Box, Typography, Avatar, IconButton, Tooltip, Paper } from '@mui/material';
import PersonIcon from '@mui/icons-material/Person';
import SmartToyIcon from '@mui/icons-material/SmartToy';
import ReplayIcon from '@mui/icons-material/Replay';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import type { Message } from '../types/chat';

interface MessageBubbleProps {
  message: Message;
  isStreaming?: boolean;
  isLast?: boolean;
  onRegenerate?: (messageId: string) => void;
  onShowInfo?: (message: Message, event: React.MouseEvent<HTMLElement>) => void;
  branchInfo?: {
    total: number;
    current: number;
    prevId?: string;
    nextId?: string;
  };
  onSwitchBranch?: (messageId: string) => void;
}

export default function MessageBubble({ 
  message, 
  isStreaming, 
  isLast, 
  onRegenerate, 
  onShowInfo,
  branchInfo,
  onSwitchBranch 
}: MessageBubbleProps) {
  const isUser = message.role === 'user';
  const [isHovered, setIsHovered] = useState(false);

  return (
    <Box
      className={isLast ? 'fade-in-up' : undefined}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      sx={{
        display: 'flex',
        gap: 2,
        mb: 3,
        flexDirection: isUser ? 'row-reverse' : 'row',
        alignItems: 'flex-start',
        px: { xs: 1, md: 2 },
      }}
    >
      {/* Avatar with soft design */}
      <Avatar
        sx={{
          width: 34,
          height: 34,
          bgcolor: isUser ? 'primary.main' : 'background.paper',
          color: isUser ? 'primary.contrastText' : 'primary.main',
          border: '1px solid',
          borderColor: isUser ? 'primary.main' : 'divider',
          flexShrink: 0,
          mt: 0.5,
          boxShadow: '0 2px 4px rgba(0,0,0,0.05)',
        }}
      >
        {isUser ? (
          <PersonIcon sx={{ fontSize: 18 }} />
        ) : (
          <SmartToyIcon sx={{ fontSize: 18 }} />
        )}
      </Avatar>

      {/* Message structure */}
      <Box
        sx={{
          maxWidth: { xs: '88%', md: '75%' },
          minWidth: 60,
          display: 'flex',
          flexDirection: 'column',
          alignItems: isUser ? 'flex-end' : 'flex-start',
        }}
      >
        {/* Role label & Meta */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.75, px: 0.5 }}>
          <Typography
            variant="caption"
            sx={{
              fontWeight: 700,
              color: 'text.secondary',
              fontSize: '0.65rem',
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              fontFamily: '"Outfit", sans-serif',
            }}
          >
            {isUser ? 'Researcher' : (message.model || 'Flow Intelligence')}
          </Typography>
          
          <Typography variant="caption" sx={{ color: 'text.secondary', opacity: 0.4, fontSize: '0.6rem' }}>
            {new Date(message.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </Typography>
        </Box>

        {/* The Bubble */}
        <Paper
          elevation={0}
          sx={{
            px: 2.5,
            py: 1.75,
            borderRadius: isUser
              ? '20px 4px 20px 20px'
              : '4px 20px 20px 20px',
            bgcolor: isUser ? 'primary.main' : 'background.paper',
            color: isUser ? 'primary.contrastText' : 'text.primary',
            border: '1px solid',
            borderColor: isUser ? 'primary.main' : 'divider',
            position: 'relative',
            boxShadow: isUser 
              ? '0 4px 12px rgba(79, 70, 229, 0.15)' 
              : '0 2px 8px rgba(0,0,0,0.02)',
            transition: 'all 0.2s ease',
            '&:hover': {
              boxShadow: isUser 
                ? '0 6px 16px rgba(79, 70, 229, 0.2)' 
                : '0 4px 12px rgba(0,0,0,0.05)',
            }
          }}
        >
          <Typography 
            variant="body1" 
            className={isStreaming && message.content ? 'streaming-cursor' : undefined}
            sx={{ 
              whiteSpace: 'pre-wrap', 
              wordBreak: 'break-word',
              fontFamily: '"Inter", sans-serif',
              fontWeight: 400,
              lineHeight: 1.6,
              display: 'flex',
              alignItems: 'center',
              gap: 1,
            }}
          >
            {message.content ? (
              message.content
            ) : isStreaming ? (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                <Typography variant="body2" sx={{ fontStyle: 'italic', opacity: 0.7 }}>Thinking</Typography>
                <Box className="thinking-dots">
                  <span></span>
                  <span></span>
                  <span></span>
                </Box>
              </Box>
            ) : null}
          </Typography>
        </Paper>

        {/* Interaction Toolbar */}
        <Box 
          sx={{ 
            display: 'flex', 
            alignItems: 'center', 
            gap: 1, 
            mt: 1, 
            minHeight: 28,
            px: 0.5,
          }}
        >
          {/* Branch Nav */}
          {!isUser && !isStreaming && branchInfo && branchInfo.total > 1 && (
            <Box 
              sx={{ 
                display: 'flex', 
                alignItems: 'center', 
                bgcolor: 'action.hover',
                borderRadius: '8px',
                p: 0.25,
                border: '1px solid',
                borderColor: 'divider',
              }}
            >
              <IconButton 
                size="small" 
                disabled={!branchInfo.prevId} 
                onClick={() => branchInfo.prevId && onSwitchBranch && onSwitchBranch(branchInfo.prevId)}
                sx={{ width: 22, height: 22, p: 0 }}
              >
                <ChevronLeftIcon sx={{ fontSize: 14 }} />
              </IconButton>
              
              <Typography 
                variant="caption" 
                sx={{ 
                  fontSize: '0.65rem', 
                  fontWeight: 700, 
                  px: 0.75, 
                  minWidth: 32, 
                  textAlign: 'center',
                  fontFamily: '"Outfit", sans-serif'
                }}
              >
                {branchInfo.current} / {branchInfo.total}
              </Typography>

              <IconButton 
                size="small" 
                disabled={!branchInfo.nextId} 
                onClick={() => branchInfo.nextId && onSwitchBranch && onSwitchBranch(branchInfo.nextId)}
                sx={{ width: 22, height: 22, p: 0 }}
              >
                <ChevronRightIcon sx={{ fontSize: 14 }} />
              </IconButton>
            </Box>
          )}

          {/* Action Buttons */}
          {!isUser && !isStreaming && (
            <Box
              sx={{
                display: 'flex',
                gap: 0.5,
                opacity: isHovered ? 1 : 0,
                transition: 'opacity 0.2s',
              }}
            >
              {onRegenerate && (
                <Tooltip title="Explore Alternative">
                  <IconButton
                    size="small"
                    onClick={() => onRegenerate(message.id)}
                    sx={{ width: 26, height: 26, p: 0.5, border: '1px solid', borderColor: 'divider' }}
                  >
                    <ReplayIcon sx={{ fontSize: 14, color: 'text.secondary' }} />
                  </IconButton>
                </Tooltip>
              )}
              {onShowInfo && Object.keys(message.metadata || {}).length > 0 && (
                <Tooltip title="Inference Details">
                  <IconButton
                    size="small"
                    onClick={(e) => onShowInfo(message, e)}
                    sx={{ width: 26, height: 26, p: 0.5, border: '1px solid', borderColor: 'divider' }}
                  >
                    <InfoOutlinedIcon sx={{ fontSize: 14, color: 'text.secondary' }} />
                  </IconButton>
                </Tooltip>
              )}
            </Box>
          )}
        </Box>
      </Box>
    </Box>
  );
}
