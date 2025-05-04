import React, { useEffect, useRef } from 'react';
import { Message } from '../../types';
import styles from './MessageList.module.css';

interface MessageListProps {
  messages: Message[];
}

const MessageList: React.FC<MessageListProps> = ({ messages }) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  return (
    <div className={styles.messagesArea}>
      {messages.map((msg, index) => (
        <div key={index} className={`${styles.messageBubble} ${styles[msg.sender]}`}>
          <p>{msg.text}</p>
        </div>
      ))}
      <div ref={messagesEndRef} />
    </div>
  );
};

export default MessageList;