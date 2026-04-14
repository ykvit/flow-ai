import React from 'react';
import {
  Box,
  Drawer,
  List,
  Typography,
  Button,
  IconButton,
  Divider,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import SettingsIcon from '@mui/icons-material/Settings';
import ChatListItem from './ChatListItem';
import { useChatStore } from '../stores/chats';
import type { Chat } from '../types/chat';

interface ChatSidebarProps {
  open: boolean;
  onClose: () => void;
  onNewChat: () => void;
  onOpenSettings: () => void;
  onChatSelect: () => void;
  drawerWidth: number;
  isMobile: boolean;
}

export default function ChatSidebar({
  open,
  onClose,
  onNewChat,
  onOpenSettings,
  onChatSelect,
  drawerWidth,
  isMobile,
}: ChatSidebarProps) {
  const { chats, currentChat, fetchChatById, deleteChat, updateChatTitle } = useChatStore();

  const handleChatClick = async (chatId: string) => {
    await fetchChatById(chatId);
    onChatSelect();
  };

  const handleDeleteChat = async (chatId: string) => {
    await deleteChat(chatId);
  };

  const handleRenameChat = async (chatId: string, newTitle: string) => {
    await updateChatTitle(chatId, { title: newTitle });
  };

  const drawerContent = (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        bgcolor: 'background.paper',
      }}
    >
      {/* Header */}
      <Box sx={{ p: 3, pb: 2 }}>
        <Typography
          variant="h5"
          sx={{
            fontWeight: 700,
            mb: 3,
            color: 'primary.main',
            letterSpacing: '-0.03em',
            fontFamily: '"Outfit", sans-serif',
            display: 'flex',
            alignItems: 'center',
            gap: 1.5,
          }}
        >
          <Box
            sx={{
              width: 32,
              height: 32,
              borderRadius: '8px',
              bgcolor: 'primary.main',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'white',
              fontSize: '1.2rem',
              fontWeight: 800,
            }}
          >
            F
          </Box>
          Flow-AI
        </Typography>
        <Button
          id="new-chat-button"
          variant="contained"
          fullWidth
          startIcon={<AddIcon />}
          onClick={onNewChat}
          sx={{
            py: 1.2,
            borderRadius: '12px',
            fontWeight: 600,
            fontSize: '0.875rem',
            background: 'linear-gradient(135deg, #4f46e5 0%, #6366f1 100%)',
            boxShadow: '0 4px 12px rgba(79, 70, 229, 0.2)',
            transition: 'all 0.2s ease',
            '&:hover': {
              transform: 'translateY(-1px)',
              boxShadow: '0 6px 16px rgba(79, 70, 229, 0.3)',
            },
          }}
        >
          New Insight
        </Button>
      </Box>

      <Divider sx={{ mx: 2, opacity: 0.5 }} />

      {/* Chat List */}
      <List
        sx={{
          flex: 1,
          overflow: 'auto',
          px: 1,
          py: 1,
        }}
      >
        {chats.length === 0 ? (
          <Box sx={{ p: 3, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary" sx={{ opacity: 0.6 }}>
              No conversations yet
            </Typography>
          </Box>
        ) : (
          (() => {
            const now = new Date();
            const oneDay = 24 * 60 * 60 * 1000;
            const sevenDays = 7 * oneDay;
            const thirtyDays = 30 * oneDay;

            const groups: { label: string; chats: Chat[] }[] = [
              { label: 'Recent', chats: [] },
              { label: 'This Week', chats: [] },
              { label: 'This Month', chats: [] },
              { label: 'Older', chats: [] },
            ];

            chats.forEach((chat) => {
              const updated = new Date(chat.updated_at);
              const diff = now.getTime() - updated.getTime();

              if (diff < 2 * oneDay) {
                groups[0].chats.push(chat);
              } else if (diff < sevenDays) {
                groups[1].chats.push(chat);
              } else if (diff < thirtyDays) {
                groups[2].chats.push(chat);
              } else {
                groups[3].chats.push(chat);
              }
            });

            return groups.map((group) => {
              if (group.chats.length === 0) return null;

              return (
                <Box key={group.label} sx={{ mb: 1 }}>
                  <Typography
                    variant="caption"
                    sx={{
                      px: 2,
                      py: 1,
                      display: 'block',
                      color: 'text.secondary',
                      fontWeight: 700,
                      textTransform: 'uppercase',
                      letterSpacing: '0.05em',
                      fontSize: '0.65rem',
                      fontFamily: '"Outfit", sans-serif',
                      opacity: 0.8,
                    }}
                  >
                    {group.label}
                  </Typography>
                  {group.chats.map((chat) => (
                    <ChatListItem
                      key={chat.id}
                      chat={chat}
                      isActive={currentChat?.id === chat.id}
                      onClick={() => handleChatClick(chat.id)}
                      onDelete={() => handleDeleteChat(chat.id)}
                      onRename={(newTitle: string) => handleRenameChat(chat.id, newTitle)}
                    />
                  ))}
                </Box>
              );
            });
          })()
        )}
      </List>

      {/* Footer */}
      <Divider sx={{ mx: 2, opacity: 0.3 }} />
      <Box sx={{ p: 2 }}>
        <IconButton
          id="settings-button"
          onClick={onOpenSettings}
          sx={{
            borderRadius: '10px',
            width: '100%',
            justifyContent: 'flex-start',
            gap: 1.5,
            px: 2,
            py: 1.2,
            color: 'text.secondary',
            transition: 'all 0.2s ease',
            '&:hover': {
              bgcolor: 'action.hover',
              color: 'text.primary',
            },
          }}
        >
          <SettingsIcon fontSize="small" />
          <Typography variant="body2" sx={{ fontWeight: 500 }}>Preferences</Typography>
        </IconButton>
      </Box>
    </Box>
  );

  return (
    <Drawer
      variant={isMobile ? 'temporary' : 'persistent'}
      anchor="left"
      open={open}
      onClose={onClose}
      sx={{
        flexShrink: 0,
        '& .MuiDrawer-paper': {
          width: drawerWidth,
          border: 'none',
          bgcolor: 'background.paper',
        },
      }}
      ModalProps={{
        keepMounted: true,
      }}
    >
      {drawerContent}
    </Drawer>
  );
}
