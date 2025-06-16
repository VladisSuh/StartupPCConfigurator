import { useEffect, useMemo, useState } from 'react';
import { CategoryType, Component, UsecaseLabels, Usecases } from '../../types/index';
import styles from './Filters.module.css';
import { useConfig } from '../../ConfigContext';

type FiltersProps = {
    сompatibleComponents: Component[];
    allComponents: Component[];
    setFilteredAllComponents: (components: Component[]) => void;
    setFilteredCompatibleComponents: (components: Component[]) => void;
    componentPrices: Record<string, number>;
    activeBrandTab: string;
    setActiveBrandTab: (brand: string) => void;
    activeUsecaseTab: string;
    setActiveUsecaseTab: (usecase: string) => void;
    selectedCategory: CategoryType;
    showCompatibleOnly: boolean;
    setShowCompatibleOnly: (show: boolean) => void;
};

const Filters = ({
    сompatibleComponents,
    allComponents,
    setFilteredAllComponents,
    setFilteredCompatibleComponents,
    componentPrices,
    activeBrandTab,
    setActiveBrandTab,
    activeUsecaseTab,
    setActiveUsecaseTab,
    selectedCategory,
    showCompatibleOnly,
    setShowCompatibleOnly
}: FiltersProps) => {

    const [searchQuery, setSearchQuery] = useState('');
    const [sortOption, setSortOption] = useState<'default' | 'increasing' | 'decreasing'>('default');
    const [brands, setBrands] = useState<string[]>([]);
    const { getBrands, theme } = useConfig()

    const filteredComponents = useMemo(() => {
        return (components: Component[]) => {
            if (!components) return [];
            let filtered = [...components];

            if (searchQuery) {
                const query = searchQuery.toLowerCase();
                filtered = filtered.filter(c =>
                    c.name.toLowerCase().includes(query)
                );
            }

            switch (sortOption) {
                case 'increasing':
                    filtered.sort((a, b) => (componentPrices[a.id] || 0) - (componentPrices[b.id] || 0));
                    break;
                case 'decreasing':
                    filtered.sort((a, b) => (componentPrices[b.id] || 0) - (componentPrices[a.id] || 0));
                    break;
                default:
                    break;
            }

            return filtered;
        };
    }, [ sortOption, componentPrices, searchQuery]);

    useEffect(() => {
        setFilteredCompatibleComponents(filteredComponents(сompatibleComponents));
        setFilteredAllComponents(filteredComponents(allComponents));
    }, [filteredComponents, сompatibleComponents, allComponents]);


    useEffect(() => {
        setBrands(['Все бренды', ...getBrands(selectedCategory)]);
    }, [selectedCategory, getBrands]);

    return (
        <div className={styles.filtersContainer}>
            <div className={styles.searchContainer}>
                <input
                    type="text"
                    placeholder="Поиск по названию или бренду..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className={`${styles.searchInput} ${styles[theme]}`}
                />
            </div>

            {<div className={styles.tabs}>
                {Usecases.map(usecase => (
                    <button
                        key={usecase}
                        className={`${styles.usecaseTab} ${usecase === activeUsecaseTab ? styles.activeUsecaseTab : ''} ${styles[theme]}`}
                        onClick={() => setActiveUsecaseTab(usecase)}
                    >
                        {UsecaseLabels[usecase] || usecase}
                    </button>
                ))}
            </div>}

            {<div className={styles.brandTabs}>
                {brands.map(brand => (
                    <button
                        key={brand}
                        className={`${styles.brandTab} ${brand === activeBrandTab ? styles.activeBrandTab : ''} ${styles[theme]}`}
                        onClick={() => setActiveBrandTab(brand)}
                    >
                        {brand}
                    </button>
                ))}
            </div>}

            <div className={styles.controls}>
                <label className={styles.checkboxLabel}>
                    <input
                        type="checkbox"
                        className={styles.checkbox}
                        checked={showCompatibleOnly}
                        onChange={() => setShowCompatibleOnly(!showCompatibleOnly)}
                    />
                    Совместимые товары
                </label>
                <div>
                    <label htmlFor="sort" className={styles.label}>Сортировка:</label>
                    <select
                        id="sort"
                        name="sort"
                        className={`${styles.select} ${styles[theme]}`}
                        value={sortOption}
                        onChange={(e) => setSortOption(e.target.value as any)}
                    >
                        <option value="default">По умолчанию</option>
                        <option value="increasing">По возрастанию цены</option>
                        <option value="decreasing">По убыванию цены</option>
                    </select>
                </div>
            </div>
        </div>
    );
};

export default Filters;