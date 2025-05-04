import React from 'react';
import styles from './WelcomeMessage.module.css';

const WelcomeMessage: React.FC = () => {
  return (
    <div className={styles.welcomeText}>
      <span className={styles.welcomeLine1}>How was <em className={styles.emphasisText}>your</em> day</span>
      <span className={styles.welcomeLine2}>today?</span>
    </div>
  );
};

export default WelcomeMessage;