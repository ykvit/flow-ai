import { useState, useCallback, MouseEvent } from 'react';
import { Box, useMediaQuery, useTheme, Popover, Typography, Divider } from '@mui/material';
import ChatSidebar from './ChatSidebar';
import ChatHeader from './ChatHeader';
import MessageList from './MessageList';
import MessageInput from './MessageInput';
import { useChatStore } from '../stores/chats';
import { useSettingsStore } from '../stores/settings';
import { useModelsStore } from '../stores/models';
import type { Message } from '../types/chat';

const DRAWER_WIDTH = 300;

interface ChatLayoutProps {
  onOpenSettings: () => void;
}

export default function ChatLayout({ onOpenSettings }: ChatLayoutProps) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const [sidebarOpen, setSidebarOpen] = useState(!isMobile);

  const {
    currentChat,
    isStreaming,
    streamingContent,
    createMessage,
    createNewChat,
    regenerateMessage,
    switchBranch,
  } = useChatStore();

  const { settings } = useSettingsStore();
  const { models } = useModelsStore();
  const [selectedModel, setSelectedModel] = useState('');

  // Use the model from settings as default
  const activeModel = selectedModel || settings.main_model || (models.length > 0 ? models[0].name : '');

  const handleSendMessage = useCallback(async (content: string) => {
    if (!content.trim() || isStreaming) return;

    await createMessage({
      chat_id: currentChat?.id || '',
      content: content.trim(),
      model: activeModel,
      system_prompt: settings.system_prompt || undefined,
      support_model: settings.support_model || undefined,
    });
  }, [currentChat, activeModel, settings, isStreaming, createMessage]);

  const handleNewChat = useCallback(() => {
    createNewChat();
    if (isMobile) setSidebarOpen(false);
  }, [createNewChat, isMobile]);

  const handleToggleSidebar = useCallback(() => {
    setSidebarOpen(prev => !prev);
  }, []);

  const handleChatSelect = useCallback(() => {
    if (isMobile) setSidebarOpen(false);
  }, [isMobile]);

  // Metadata Popover State
  const [metaAnchorEl, setMetaAnchorEl] = useState<HTMLElement | null>(null);
  const [metaMessage, setMetaMessage] = useState<Message | null>(null);

  const handleShowInfo = useCallback((message: Message, event: MouseEvent<HTMLElement>) => {
    setMetaMessage(message);
    setMetaAnchorEl(event.currentTarget);
  }, []);

  const handleCloseMeta = useCallback(() => {
    setMetaAnchorEl(null);
    setMetaMessage(null);
  }, []);

  const handleRegenerate = useCallback((messageId: string) => {
    if (currentChat) {
      regenerateMessage({
        chat_id: currentChat.id,
        message_id: messageId,
        model: activeModel,
      });
    }
  }, [currentChat, activeModel, regenerateMessage]);

  return (
    <Box sx={{ display: 'flex', height: '100vh', width: '100%', overflow: 'hidden', bgcolor: 'background.default' }}>
      {/* Sidebar */}
      <ChatSidebar
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onNewChat={handleNewChat}
        onOpenSettings={onOpenSettings}
        onChatSelect={handleChatSelect}
        drawerWidth={DRAWER_WIDTH}
        isMobile={isMobile}
      />

      {/* Main Chat Area */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          display: 'flex',
          flexDirection: 'column',
          height: '100vh',
          overflow: 'hidden',
          bgcolor: 'background.default',
          transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          ml: !isMobile && sidebarOpen ? `${DRAWER_WIDTH}px` : 0,
          position: 'relative',
        }}
      >
        <ChatHeader
          title={currentChat?.title || 'New Chat'}
          onToggleSidebar={handleToggleSidebar}
          sidebarOpen={sidebarOpen}
        />

        <MessageList
          messages={currentChat?.messages || []}
          isStreaming={isStreaming}
          streamingContent={streamingContent}
          activeModel={activeModel}
          onRegenerate={handleRegenerate}
          onShowInfo={handleShowInfo}
          onSwitchBranch={switchBranch}
        />

        <MessageInput
          onSend={handleSendMessage}
          disabled={isStreaming || !activeModel}
          isStreaming={isStreaming}
        />
      </Box>

      {/* Metadata Popover */}
      <Popover
        open={Boolean(metaAnchorEl)}
        anchorEl={metaAnchorEl}
        onClose={handleCloseMeta}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'center',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'center',
        }}
        slotProps={{
          paper: {
            sx: { p: 2, maxWidth: 300, borderRadius: 2, bgcolor: 'background.paper', backgroundImage: 'none' }
          }
        }}
      >
        <Typography variant="subtitle2" sx={{ mb: 1, fontWeight: 600 }}>Message Details</Typography>
        <Divider sx={{ mb: 1 }} />
        {metaMessage?.metadata && Object.entries(metaMessage.metadata).map(([key, value]) => (
          <Box key={key} sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
            <Typography variant="caption" sx={{ color: 'text.secondary', mr: 2 }}>{key}:</Typography>
            <Typography variant="caption" sx={{ fontFamily: 'monospace' }}>
              {typeof value === 'number' && key.includes('duration') 
                ? `${(value / 1e9).toFixed(2)}s` 
                : String(value)}
            </Typography>
          </Box>
        ))}
      </Popover>
    </Box>
  );
}
