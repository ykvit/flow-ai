import React from 'react';
import styles from './SettingsNav.module.css';
import { SettingsCategory } from './SettingsModal'; 

interface Category {
  id: SettingsCategory;
  label: string;
  icon: React.FC<React.SVGProps<SVGSVGElement>>;
}

interface SettingsNavProps {
  categories: Category[];
  activeCategory: SettingsCategory;
  onSelectCategory: (category: SettingsCategory) => void;
}

const SettingsNav: React.FC<SettingsNavProps> = ({
  categories,
  activeCategory,
  onSelectCategory,
}) => {
  return (
    <nav className={styles.nav}>
      <ul className={styles.navList}>
        {categories.map((category) => (
          <li key={category.id}>
            <button
              className={`${styles.navButton} ${
                activeCategory === category.id ? styles.active : ''
              }`}
              onClick={() => onSelectCategory(category.id)}
            >
              <category.icon className={styles.navIcon} />
              <span className={styles.navLabel}>{category.label}</span>
            </button>
          </li>
        ))}
      </ul>
    </nav>
  );
};

export default SettingsNav;