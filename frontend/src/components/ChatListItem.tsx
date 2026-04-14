import * as React from 'react';
import { useState, useRef, useEffect } from 'react';
import {
  ListItemButton,
  ListItemText,
  IconButton,
  Typography,
  Box,
  TextField,
  Tooltip,
} from '@mui/material';
import DeleteOutlinedIcon from '@mui/icons-material/DeleteOutlined';
import EditOutlinedIcon from '@mui/icons-material/EditOutlined';
import CheckIcon from '@mui/icons-material/Check';
import CloseIcon from '@mui/icons-material/Close';
import ChatBubbleOutlinedIcon from '@mui/icons-material/ChatBubbleOutlined';
import type { Chat } from '../types/chat';

interface ChatListItemProps {
  chat: Chat;
  isActive: boolean;
  onClick: () => void;
  onDelete: () => void;
  onRename: (newTitle: string) => void;
}

export default function ChatListItem({ chat, isActive, onClick, onDelete, onRename }: ChatListItemProps) {
  const [isHovered, setIsHovered] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editTitle, setEditTitle] = useState(chat.title);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing]);

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  const handleRename = (e?: React.FormEvent | React.MouseEvent) => {
    e?.stopPropagation();
    if (editTitle.trim() && editTitle !== chat.title) {
      onRename(editTitle.trim());
    }
    setIsEditing(false);
  };

  const cancelRename = (e: React.MouseEvent) => {
    e.stopPropagation();
    setEditTitle(chat.title);
    setIsEditing(false);
  };

  const startEditing = (e: React.MouseEvent) => {
    e.stopPropagation();
    setEditTitle(chat.title);
    setIsEditing(true);
  };

  return (
    <ListItemButton
      selected={isActive}
      onClick={isEditing ? undefined : onClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      sx={{
        borderRadius: '10px',
        mb: 0.5,
        px: 1.5,
        py: 1,
        transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
        '&.Mui-selected': {
          bgcolor: 'action.selected',
          color: 'primary.main',
          '&:hover': {
            bgcolor: 'action.selected',
          },
          '& .MuiTypography-root': {
            fontWeight: 600,
            color: 'primary.main',
          },
        },
        '&:hover': {
          bgcolor: 'action.hover',
          '& .item-actions': {
            opacity: 1,
          },
        },
      }}
    >
      <ChatBubbleOutlinedIcon
        sx={{
          fontSize: 16,
          mr: 1.5,
          color: isActive ? 'primary.main' : 'text.secondary',
          opacity: isActive ? 1 : 0.4,
          flexShrink: 0,
        }}
      />
      
      {isEditing ? (
        <Box sx={{ flex: 1, mr: 1, display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <TextField
            fullWidth
            size="small"
            value={editTitle}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEditTitle(e.target.value)}
            onKeyDown={(e: React.KeyboardEvent) => {
              if (e.key === 'Enter') handleRename();
              if (e.key === 'Escape') setIsEditing(false);
            }}
            inputRef={inputRef}
            onClick={(e: React.MouseEvent) => e.stopPropagation()}
            sx={{
              '& .MuiInputBase-input': {
                fontSize: '0.875rem',
                py: 0.5,
                px: 1,
              },
            }}
          />
          <IconButton size="small" onClick={handleRename} sx={{ color: 'success.main' }}>
            <CheckIcon sx={{ fontSize: 16 }} />
          </IconButton>
          <IconButton size="small" onClick={cancelRename} sx={{ color: 'text.secondary' }}>
            <CloseIcon sx={{ fontSize: 16 }} />
          </IconButton>
        </Box>
      ) : (
        <>
          <ListItemText
            primary={
              <Typography
                variant="body2"
                sx={{
                  fontWeight: isActive ? 500 : 400,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                  color: isActive ? 'text.primary' : 'text.secondary',
                }}
              >
                {chat.title || 'Untitled Chat'}
              </Typography>
            }
          />
          <Box
            sx={{
              opacity: isHovered ? 1 : 0,
              transition: 'opacity 0.15s ease',
              flexShrink: 0,
              display: 'flex',
              gap: 0,
            }}
          >
            <Tooltip title="Rename" placement="top">
              <IconButton
                size="small"
                onClick={startEditing}
                sx={{
                  color: 'text.secondary',
                  '&:hover': {
                    color: 'primary.main',
                    bgcolor: 'action.hover',
                  },
                }}
              >
                <EditOutlinedIcon sx={{ fontSize: 18 }} />
              </IconButton>
            </Tooltip>
            <Tooltip title="Delete" placement="top">
              <IconButton
                size="small"
                onClick={(e: React.MouseEvent) => {
                  e.stopPropagation();
                  onDelete();
                }}
                sx={{
                  color: 'text.secondary',
                  '&:hover': {
                    color: 'error.main',
                    backgroundColor: 'rgba(211, 47, 47, 0.08)',
                  },
                }}
              >
                <DeleteOutlinedIcon sx={{ fontSize: 18 }} />
              </IconButton>
            </Tooltip>
          </Box>
        </>
      )}
    </ListItemButton>
  );
}
