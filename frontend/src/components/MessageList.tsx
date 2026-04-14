import { useEffect, useRef } from 'react';
import { Box, Typography } from '@mui/material';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import { motion, AnimatePresence } from 'framer-motion';
import MessageBubble from './MessageBubble';
import type { Message } from '../types/chat';

interface MessageListProps {
  messages: Message[];
  isStreaming: boolean;
  streamingContent: string;
  activeModel?: string;
  onRegenerate?: (messageId: string) => void;
  onShowInfo?: (message: Message, event: React.MouseEvent<HTMLElement>) => void;
  onSwitchBranch?: (messageId: string) => void;
}

export default function MessageList({
  messages,
  isStreaming,
  streamingContent,
  activeModel,
  onRegenerate,
  onShowInfo,
  onSwitchBranch
}: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom on new messages or streaming
  useEffect(() => {
    if (bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, streamingContent]);

  const activeMessages = messages.filter(m => m.is_active);

  // Empty state
  if (activeMessages.length === 0 && !isStreaming) {
    return (
      <Box
        sx={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 3,
          p: 6,
          bgcolor: 'background.default',
        }}
      >
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5 }}
        >
          <Box
            sx={{
              width: 80,
              height: 80,
              borderRadius: '24px',
              bgcolor: 'background.paper',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              border: '1px solid',
              borderColor: 'divider',
              boxShadow: '0 4px 20px rgba(0,0,0,0.03)',
              mb: 2,
            }}
          >
            <AutoAwesomeIcon sx={{ fontSize: 36, color: 'primary.main', opacity: 0.8 }} />
          </Box>
        </motion.div>
        
        <Typography 
          variant="h4" 
          sx={{ 
            fontWeight: 700, 
            color: 'text.primary', 
            fontFamily: '"Outfit", sans-serif',
            letterSpacing: '-0.03em',
          }}
        >
          Initiate Inquiry
        </Typography>
        <Typography 
          variant="body1" 
          sx={{ 
            color: 'text.secondary', 
            textAlign: 'center', 
            maxWidth: 480,
            opacity: 0.7,
            lineHeight: 1.8,
          }}
        >
          Enter a prompt to begin your session. Flow-AI will synthesize insights from the selected model and allow you to explore branching paths of thought.
        </Typography>
      </Box>
    );
  }

  return (
    <Box
      ref={containerRef}
      sx={{
        flex: 1,
        overflow: 'auto',
        px: { xs: 1, sm: 2, md: 4 },
        py: 4,
      }}
    >
      <Box sx={{ maxWidth: 840, mx: 'auto', width: '100%' }}>
        <AnimatePresence initial={false}>
          {activeMessages.map((message, index) => {
            // Calculate siblings
            const siblings = messages.filter(m => m.parent_id === message.parent_id);
            const sortedSiblings = [...siblings].sort(
              (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
            );
            const currentSiblingIndex = sortedSiblings.findIndex(s => s.id === message.id);
            const prevSiblingId = currentSiblingIndex > 0 ? sortedSiblings[currentSiblingIndex - 1].id : undefined;
            const nextSiblingId = currentSiblingIndex < sortedSiblings.length - 1 ? sortedSiblings[currentSiblingIndex + 1].id : undefined;

            return (
              <motion.div
                key={message.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                transition={{ duration: 0.3 }}
              >
                <MessageBubble
                  message={message}
                  isLast={index === activeMessages.length - 1 && !isStreaming}
                  onRegenerate={onRegenerate}
                  onShowInfo={onShowInfo}
                  branchInfo={siblings.length > 1 ? {
                    total: siblings.length,
                    current: currentSiblingIndex + 1,
                    prevId: prevSiblingId,
                    nextId: nextSiblingId,
                  } : undefined}
                  onSwitchBranch={onSwitchBranch}
                />
              </motion.div>
            );
          })}

          {/* Streaming message */}
          {isStreaming && (
            <motion.div
              layout
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
            >
              <MessageBubble
                message={{
                  id: 'streaming',
                  content: streamingContent,
                  role: 'assistant',
                  model: activeModel || '',
                  parent_id: '',
                  metadata: {},
                  timestamp: new Date().toISOString(),
                  is_active: true,
                }}
                isStreaming
                isLast
              />
            </motion.div>
          )}
        </AnimatePresence>

        <div ref={bottomRef} />
      </Box>
    </Box>
  );
}
