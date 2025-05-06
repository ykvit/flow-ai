import React, { useState, useRef, useEffect, useCallback } from 'react';
import styles from './ChatInput.module.css';
import PlusIcon from '../../assets/plus-icon.svg?react';
import BulbIcon from '../../assets/bulb-icon.svg?react';
import SendIcon from '../../assets/Arrow.svg?react';
import SurseButton from '../../assets/planet-icon.svg?react';
import StopIcon from '../../assets/stop-icon.svg?react';

interface ChatInputProps {
  onSendMessage: (message: string) => void;
  isLoading: boolean;
  inputRef: React.Ref<HTMLTextAreaElement>;
  onStopGenerating?: () => void;
}

// Константи, які можна налаштувати (мають відповідати CSS)
const TEXTAREA_BASE_BORDER_BOX_HEIGHT_ONE_LINE = 50; // Приблизна висота textarea для одного рядка (напр. 24px контент + 2*13px паддінг)
const TEXTAREA_MAX_BORDER_BOX_HEIGHT = 226; // Приблизна макс. висота textarea (напр. 10 рядків * 24px + 2*13px паддінг = 266px, але CSS мав 226px)
                                          // Узгодьте з CSS .chatInput max-height + vertical paddings
const CONTAINER_VERTICAL_SPACING = 60; // Загальний вертикальний простір у контейнері поза textarea (10px зверху textarea + 50px знизу для кнопок)
const INITIAL_CONTAINER_HEIGHT = TEXTAREA_BASE_BORDER_BOX_HEIGHT_ONE_LINE + CONTAINER_VERTICAL_SPACING; // e.g. 50 + 60 = 110px

const ChatInput: React.FC<ChatInputProps> = ({
  onSendMessage,
  isLoading,
  inputRef, // Цей ref тепер для textarea, як і раніше
  onStopGenerating,
}) => {
  const [inputValue, setInputValue] = useState<string>('');
  const hasText = inputValue.trim().length > 0;

  const textareaRefInternal = useRef<HTMLTextAreaElement | null>(null);
  const inputAreaContainerRef = useRef<HTMLDivElement | null>(null);
  const inputBgWhiteRef = useRef<HTMLDivElement | null>(null);
  const inputBgBlurRef = useRef<HTMLDivElement | null>(null);

  // Комбінований реф для зовнішнього та внутрішнього використання textarea
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
      
      ta.style.height = 'auto'; // Скидаємо для розрахунку scrollHeight
      let newTextareaHeight = ta.scrollHeight;

      // Обмежуємо висоту textarea
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

      // Розраховуємо нову висоту контейнера
      const newContainerHeight = newTextareaHeight + CONTAINER_VERTICAL_SPACING;

      inputAreaContainerRef.current.style.height = `${newContainerHeight}px`;
      inputBgWhiteRef.current.style.height = `${newContainerHeight}px`;
      inputBgBlurRef.current.style.height = `${newContainerHeight}px`;
    }
  }, []); // useCallback, оскільки залежностей немає (константи ззовні)

  useEffect(() => {
    // Початкове налаштування розміру при монтуванні та при зміні inputValue
    adjustInputAreaLayout();
  }, [inputValue, adjustInputAreaLayout]);

  const handleInputChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInputValue(event.target.value);
    // adjustInputAreaLayout буде викликано через useEffect [inputValue]
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
      setInputValue(''); // Це викличе useEffect -> adjustInputAreaLayout
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
    <div ref={inputAreaContainerRef} className={`${styles.inputAreaContainer} ${isLoading ? styles.loading : ''}`}>
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

      <button className={`${styles.actionButton} ${styles.addButton}`} aria-label="Add file" disabled={isLoading}>
        <PlusIcon />
      </button>
      <button className={`${styles.actionButton} ${styles.sourcesButton}`} aria-label="Sources" disabled={isLoading}>
        <SurseButton />
      </button>
      <button className={`${styles.actionButton} ${styles.reasoningButton}`} aria-label="Reasoning" disabled={isLoading}>
        <BulbIcon /> <span>Reasoning</span>
      </button>

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