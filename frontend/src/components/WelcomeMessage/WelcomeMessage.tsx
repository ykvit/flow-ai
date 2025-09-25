import React, { useState, useEffect, useMemo } from 'react';
import styles from './WelcomeMessage.module.css';

const welcomePhrases = [
  { text: "How's your day going so far?", highlight: "your" },
  { text: "Hope your day's been treating you well!", highlight: "your" },
  { text: "Good to see you! How was your day?", highlight: "you" },
  { text: "Did you have a good day today?", highlight: "you" },
  { text: "How have you been? How was your day?", highlight: "you" },
  { text: "What made your day special today?", highlight: "your" },
  { text: "How did your day unfold?", highlight: "your" },
  { text: "Is there anything from your day you'd like to share?", highlight: "your" },
  { text: "How are you feeling after your day?", highlight: "you" },
  { text: "Was your day as awesome as you are?", highlight: "your" },
  { text: "How was your day today?", highlight: "your" }, 
];

interface WelcomeMessageProps { 
  isSidebarOpen: boolean;
}

const WelcomeMessage: React.FC<WelcomeMessageProps> = ({ isSidebarOpen }) => { 
  const [currentPhrase, setCurrentPhrase] = useState<{ text: string; highlight: string } | null>(null);

  useEffect(() => {
    const phrase = getRandomElement(welcomePhrases);
    if (phrase) {
      setCurrentPhrase(phrase);
    }
  }, []);


const getRandomElement = <T,>(arr: T[]): T | undefined => {
  if (!arr || arr.length === 0) {
    return undefined;
  }
  const randomIndex = Math.floor(Math.random() * arr.length);
  return arr[randomIndex];
};



  const applyHighlight = (line: string, wordToHighlight: string): React.ReactNode => {
    const lowerLine = line.toLowerCase();
    const lowerHighlight = wordToHighlight.toLowerCase();
    const startIndex = lowerLine.indexOf(lowerHighlight);

    if (startIndex === -1) {
      return line; 
    }

    const endIndex = startIndex + wordToHighlight.length;
    const before = line.slice(0, startIndex);
    const highlightedWord = line.slice(startIndex, endIndex);
    const after = line.slice(endIndex);

    return (
      <>
        {before}
        <em className={styles.emphasisText}>{highlightedWord}</em>
        {after}
      </>
    );
  };

  const formattedTextElement = useMemo(() => {
    if (!currentPhrase) {
      return null; 
    }

    const { text, highlight } = currentPhrase;
    let line1 = text;
    let line2 = ''; 

    const words = text.split(' ');
    if (words.length > 4) { 
        const splitIndex = Math.ceil(words.length / 2);
        line1 = words.slice(0, splitIndex).join(' ');
        line2 = words.slice(splitIndex).join(' ');
    }
    const renderedLine1 = applyHighlight(line1, highlight);
    const renderedLine2 = applyHighlight(line2, highlight);

    return (
      <>
        <span className={styles.welcomeLine}>{renderedLine1}</span>
        {line2 && <span className={styles.welcomeLine}>{renderedLine2}</span>} 
      </>
    );
  }, [currentPhrase]);

  return (
    <div className={`${styles.welcomeContainer} ${isSidebarOpen ? styles.sidebarIsOpen : ''}`}> 
    <div className={styles.welcomeGreeting}>
      {formattedTextElement}
//     </div>
//   </div>
);
};

export default WelcomeMessage;