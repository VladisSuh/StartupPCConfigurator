import { useState } from "react";
import styles from './CategoryTabs.module.css';
import { CategoryTabsProps, CategoryType, categories, categoryLabels } from "../../types/index";




const CategoryTabs = ({ onSelect, selectedComponents }: CategoryTabsProps) => {
    const [activeCategory, setActiveCategory] = useState<CategoryType>("cpu");

    const handleTabClick = (category: CategoryType) => {
        setActiveCategory(category);
        onSelect(category);
    };

    const getTabClass = (category: CategoryType) => {
        const isActive = activeCategory === category;
        const hasComponent = !!selectedComponents[category];

        return [
            styles.tab,
            isActive ? styles.tabActive : '',
            !isActive && hasComponent ? styles.tabWithComponent : ''
        ].join(' ').trim();
    };

    return (
        <div className={styles.tabsContainer}>
            <div>
                {categories.map((category) => (
                    <div
                        key={category}
                        onClick={() => handleTabClick(category)}
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