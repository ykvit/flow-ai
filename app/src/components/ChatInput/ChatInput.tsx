import React, { useState } from 'react';
import styles from './ChatInput.module.css';
import PlusIcon from '../../assets/plus-icon.svg?react';
import BulbIcon from '../../assets/bulb-icon.svg?react';
import SendIcon from '../../assets/Arrow.svg?react';
import SurseButton from '../../assets/planet-icon.svg?react'; 
import StopIcon from '../../assets/stop-icon.svg?react';

interface ChatInputProps {
  onSendMessage: (message: string) => void;
  isLoading: boolean; 
  inputRef: React.Ref<HTMLInputElement>;
  onStopGenerating?: () => void;
}

const ChatInput: React.FC<ChatInputProps> = ({
  onSendMessage,
  isLoading,
  inputRef,
  onStopGenerating,
}) => {
  const [inputValue, setInputValue] = useState<string>('');
  const hasText = inputValue.trim().length > 0;

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(event.target.value);
  };

  const handleKeyPress = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter' && !event.shiftKey && !isLoading && hasText) {
      event.preventDefault();
      sendMessage();
    }
  };

  const sendMessage = () => {
    if (hasText && !isLoading) {
      onSendMessage(inputValue.trim());
      setInputValue('');
    }
  };

  const handleSendStopClick = () => {
    if (isLoading && onStopGenerating) {
      onStopGenerating();
    } else if (hasText && !isLoading) {
      sendMessage();
    }
  };

  const getSendButtonStyle = () => {
    if (isLoading) {
      return styles.stopButton;
    } else if (hasText) {
      return styles.sendButtonActive;
    } else {
      return styles.sendButtonInactive;
    }
  };

  return (

    <div className={`${styles.inputAreaContainer} ${isLoading ? styles.loading : ''}`}>
      <div
        className={`${styles.inputBgBlur} ${isLoading ? styles.loadingGradientAnimation : ''}`}
      ></div>
      <div className={styles.inputBgWhite}></div>

      {inputValue === '' && !isLoading && (
        <div className={styles.inputPlaceholderText}>Ask anything...</div>
      )}

      <input
        ref={inputRef}
        type="text"
        className={styles.chatInput}
        value={inputValue}
        onChange={handleInputChange}
        onKeyPress={handleKeyPress}
        placeholder=""
      />


      <button
        className={`${styles.actionButton} ${styles.addButton}`}
        aria-label="Add file"
        disabled={isLoading} 
      >
        <PlusIcon />
      </button>
      <button
        className={`${styles.actionButton} ${styles.sourcesButton}`}
        aria-label="Sources" 
        disabled={isLoading}
      >
        <SurseButton />
      </button>
      <button
        className={`${styles.actionButton} ${styles.reasoningButton}`}
        aria-label="Reasoning"
        disabled={isLoading} 
      >
        <BulbIcon /> <span>Reasoning</span>
      </button>

      {/* ---  Send/Stop --- */}
      <button
        className={`${styles.actionButton} ${getSendButtonStyle()}`}
        onClick={handleSendStopClick}
        disabled={isLoading ? false : !hasText}
        aria-label={isLoading ? "Stop generating" : "Send message"}
      >
        {isLoading ? <StopIcon /> : <SendIcon />}
      </button>

    </div>
  );
};

export default ChatInput;