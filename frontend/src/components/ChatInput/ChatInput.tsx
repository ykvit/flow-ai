import React, { useState, useRef, useEffect, useCallback } from 'react';
import styles from './ChatInput.module.css';
// import PlusIcon from '../../assets/plus-icon.svg?react?react';
// import BulbIcon from '../../assets/bulb-icon.svg?react?react';
// import SendIcon from '../../assets/Vector.svg?react';
// import SurseButton from '../../assets/planet-icon.svg?react?react';
// import StopIcon from '../../assets/stop-icon.svg?react?react';

interface ChatInputProps {
  onSendMessage: (message: string) => void;
  isLoading: boolean;
  inputRef: React.Ref<HTMLTextAreaElement>;
  onStopGenerating?: () => void;
  isSidebarOpen: boolean; 
}

const TEXTAREA_BASE_BORDER_BOX_HEIGHT_ONE_LINE = 50; 
const TEXTAREA_MAX_BORDER_BOX_HEIGHT = 226;
const CONTAINER_VERTICAL_SPACING = 60;


const ChatInput: React.FC<ChatInputProps> = ({
  onSendMessage,
  isLoading,
  inputRef, 
  onStopGenerating,
  isSidebarOpen,
}) => {
  const [inputValue, setInputValue] = useState<string>('');
  const hasText = inputValue.trim().length > 0;

  const textareaRefInternal = useRef<HTMLTextAreaElement | null>(null);
  const inputAreaContainerRef = useRef<HTMLDivElement | null>(null);
  const inputBgWhiteRef = useRef<HTMLDivElement | null>(null);
  const inputBgBlurRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (typeof inputRef === 'function') {
      inputRef(textareaRefInternal.current);
    } else if (inputRef && 'current' in inputRef) {
      (inputRef as React.MutableRefObject<HTMLTextAreaElement | null>).current = textareaRefInternal.current;
    }
  }, [inputRef]);


  const adjustInputAreaLayout = useCallback(() => {
    if (textareaRefInternal.current && inputAreaContainerRef.current && inputBgWhiteRef.current && inputBgBlurRef.current) {
      const ta = textareaRefInternal.current;
      
      ta.style.height = 'auto';
      let newTextareaHeight = ta.scrollHeight;

      if (newTextareaHeight < TEXTAREA_BASE_BORDER_BOX_HEIGHT_ONE_LINE) {
        newTextareaHeight = TEXTAREA_BASE_BORDER_BOX_HEIGHT_ONE_LINE;
      }
      
      if (newTextareaHeight > TEXTAREA_MAX_BORDER_BOX_HEIGHT) {
        newTextareaHeight = TEXTAREA_MAX_BORDER_BOX_HEIGHT;
        ta.style.overflowY = 'scroll';
      } else {
        ta.style.overflowY = 'hidden';
      }
      ta.style.height = `${newTextareaHeight}px`;

      const newContainerHeight = newTextareaHeight + CONTAINER_VERTICAL_SPACING;

      inputAreaContainerRef.current.style.height = `${newContainerHeight}px`;
      inputBgWhiteRef.current.style.height = `${newContainerHeight}px`;
      inputBgBlurRef.current.style.height = `${newContainerHeight}px`;
    }
  }, []); 

  useEffect(() => {
    adjustInputAreaLayout();
  }, [inputValue, adjustInputAreaLayout]);

  const handleInputChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInputValue(event.target.value);
  };

  const handleKeyPress = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
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
    if (isLoading) return styles.stopButton;
    if (hasText) return styles.sendButtonActive;
    return styles.sendButtonInactive;
  };

  return (
<div ref={inputAreaContainerRef} className={` ${styles.inputAreaContainer}  ${isLoading ? styles.loading : ''}  ${isSidebarOpen ? styles.sidebarIsOpen : ''}  `}  >
      <div ref={inputBgBlurRef} className={`${styles.inputBgBlur} ${isLoading ? styles.loadingGradientAnimation : ''}`}></div>
      <div ref={inputBgWhiteRef} className={styles.inputBgWhite}></div>

      {inputValue === '' && !isLoading && (
        <div className={styles.inputPlaceholderText}>Ask anything...</div>
      )}

      <textarea
        ref={textareaRefInternal}
        className={styles.chatInput}
        value={inputValue}
        onChange={handleInputChange}
        onKeyPress={handleKeyPress}
        placeholder=""
        rows={1}
      />


<div className={styles.leftActionButtonsContainer}>
  {/* <button 
    className={`${styles.actionButton} ${styles.addButton}`} 
    aria-label="Add file" 
    disabled={isLoading}
  >
    <PlusIcon />
  </button> */}

  {/* <button 
    className={`${styles.actionButton} ${styles.reasoningButton}`} 
    aria-label="Reasoning" 
    disabled={isLoading}
  >
    <BulbIcon /> <span>Reasoning</span>
  </button> */}


  {/* <button 
    className={`${styles.actionButton} ${styles.sourcesButton}`} 
    aria-label="Sources" 
    disabled={isLoading}
  >
    <SurseButton />
  </button> */}


</div>
      <button
        className={`${styles.actionButton} ${getSendButtonStyle()}`}
        onClick={handleSendStopClick}
        disabled={isLoading ? false : !hasText}
        aria-label={isLoading ? "Stop generating" : "Send message"}
      >
        {/* {isLoading ? <StopIcon /> : <SendIcon />} */}
      </button>
    </div>
  );
};

export default ChatInput;