import React, { useState, useEffect, useRef } from 'react';
import styles from './CustomDropdown.module.css';
import ArrowDownIcon from '../../assets/settings/arrow-down.svg?react';

export interface DropdownOption {
  value: string;
  label: string;
  disabled?: boolean;
}

interface CustomDropdownProps {
  options: DropdownOption[];
  selectedValue: string | null;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  id?: string;
  className?: string;
}

const CustomDropdown: React.FC<CustomDropdownProps> = ({
  options,
  selectedValue,
  onChange,
  placeholder = "Select...",
  disabled = false,
  id,
  className = ''
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const selectedLabel = options.find(option => option.value === selectedValue)?.label;
  
  const toggleDropdown = () => {
    if (!disabled) {
      setIsOpen(!isOpen);
    }
  };
  
  const handleOptionClick = (value: string, isDisabled?: boolean) => {
    if (isDisabled) return;
    onChange(value);
    setIsOpen(false);
  };
  
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    } else {
      document.removeEventListener('mousedown', handleClickOutside);
    }
    
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);
  
  return (
    <div 
      className={`${styles.dropdownContainer} ${disabled ? styles.disabled : ''} ${className}`} 
      ref={dropdownRef} 
      id={id}
    >
      <button
        type="button"
        className={styles.dropdownHeader}
        onClick={toggleDropdown}
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
      >
        <span className={styles.selectedValue}>
          {selectedLabel || placeholder}
        </span>
        <ArrowDownIcon className={`${styles.arrowIcon} ${isOpen ? styles.open : ''}`} />
      </button>
      
      {isOpen && (
        <ul className={styles.optionsList} role="listbox">
          {options.map((option) => (
            <li
              key={option.value}
              className={`${styles.optionItem} ${option.value === selectedValue ? styles.selected : ''} ${option.disabled ? styles.disabledOption : ''}`}
              onClick={() => handleOptionClick(option.value, option.disabled)}
              role="option"
              aria-selected={option.value === selectedValue}
              aria-disabled={option.disabled}
            >
              {option.label}
            </li>
          ))}
          {options.length === 0 && (
            <li className={`${styles.optionItem} ${styles.disabledOption}`}>
              No options available
            </li>
          )}
        </ul>
      )}
    </div>
  );
};

export default CustomDropdown;