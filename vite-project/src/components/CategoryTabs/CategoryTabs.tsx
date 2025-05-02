import { useState } from "react";
import styles from './CategoryTabs.module.css';
import { CategoryTabsProps, CategoryType, categories, categoryLabels } from "../../types/index";




const CategoryTabs = ({ onSelect }: CategoryTabsProps) => {
    const [activeCategory, setActiveCategory] = useState<CategoryType>("cpu");

    const handleTabClick = (category: CategoryType) => {
        setActiveCategory(category);
        onSelect(category);
    };

    return (
        <div className={styles.tabsContainer}>
            <div className={styles.tabsItems}>
                {categories.map((category) => (
                    <div
                        key={category}
                        onClick={() => handleTabClick(category)}
                        className={`${styles.tab} ${activeCategory === category ? styles.tabActive : ''
                            }`}
                    >
                        <span className={styles.categoryLabel}>
                            {categoryLabels[category]}
                        </span>
                    </div>
                ))}
            </div>

        </div>
    );
}

export default CategoryTabs;