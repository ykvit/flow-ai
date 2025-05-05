import React, { useState } from 'react';
import styles from './ChatInput.module.css';
import PlusIcon from '../../assets/plus-icon.svg?react';
import BulbIcon from '../../assets/bulb-icon.svg?react';
import SendIcon from '../../assets/Arrow.svg?react';
import SurseButton from '../../assets/planet-icon.svg?react'; 

interface ChatInputProps {
  onSendMessage: (message: string) => void;
  isLoading: boolean;
  inputRef: React.Ref<HTMLInputElement>; 
}

const ChatInput: React.FC<ChatInputProps> = ({ onSendMessage, isLoading, inputRef }) => {
  const [inputValue, setInputValue] = useState<string>('');

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(event.target.value);
  };

  const handleKeyPress = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter' && !event.shiftKey && !isLoading) {
      event.preventDefault();
      sendMessage();
    }
  };

  const sendMessage = () => {
    const messageToSend = inputValue.trim();
    if (messageToSend) {
      onSendMessage(messageToSend);
      setInputValue(''); 
    }
  };

  return (

    <div className={styles.inputAreaContainer}>
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
        disabled={isLoading} 
        placeholder=""
      />

      {/* button */}
      <button className={`${styles.actionButton} ${styles.addButton}`} aria-label="Add file">
        <PlusIcon />
      </button>
      <button className={`${styles.actionButton} ${styles.sourcesButton}`} aria-label="Sources">
        <SurseButton />
      </button>
      <button className={`${styles.actionButton} ${styles.reasoningButton}`} aria-label="Reasoning">
        <BulbIcon /> <span>Reasoning</span>
      </button>
      <button
        className={`${styles.actionButton} ${styles.sendButton}`}
        onClick={sendMessage}
        disabled={isLoading || !inputValue.trim()}
        aria-label="Send message"
      >
        <SendIcon />
      </button>

    </div>
  );
};

export default ChatInput;