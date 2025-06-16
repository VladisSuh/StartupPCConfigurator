import { useState } from "react";
import styles from './CategoryTabs.module.css';
import { CategoryTabsProps, CategoryType, categories, categoryLabels } from "../../types/index";
import { useConfig } from "../../ConfigContext";

const CategoryTabs = ({ onSelect, selectedComponents }: CategoryTabsProps) => {
    const [activeCategory, setActiveCategory] = useState<CategoryType>("cpu");
    const { theme } = useConfig()

    const handleTabSelect = (category: CategoryType) => {
        setActiveCategory(category);
        onSelect(category);
    };

    const getTabClass = (category: CategoryType) => {
        const isActive = activeCategory === category;
        const hasComponent = !!selectedComponents[category];

        return [
            styles.tab,
            isActive ? styles.tabActive : '',
            !isActive && hasComponent ? styles.tabWithComponent : '',
            styles[theme],
        ].join(' ').trim();
    };

    return (
        <div className={styles.tabsContainer}>
            <div className={`${styles.tabsContent} ${styles[theme]}`}>
                {categories.map((category) => (
                    <div
                        key={category}
                        onClick={() => handleTabSelect(category)}
                        className={getTabClass(category)}
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