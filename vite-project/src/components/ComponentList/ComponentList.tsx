import { useEffect, useState } from 'react';
import styles from './ComponentList.module.css';
import { ComponentCard } from '../ComponentCard/ComponentCard';
//import { mockComponents } from '../mockData/mock';
import { Component, ComponentListProps, UsecaseLabels, Usecases } from '../../types/index';
import { useConfig } from '../../ConfigContext';



const ComponentList = ({
    selectedCategory,
    selectedComponents,
    setSelectedComponents,
}: ComponentListProps) => {
    const [allComponents, setAllComponents] = useState<Component[]>([]);
    const [сompatibleComponents, setCompatibleComponents] = useState<Component[]>([]);
    const [showCompatibleOnly, setShowCompatibleOnly] = useState(true);
    const [isFiltersVisible, setIsFiltersVisible] = useState(false);
    const [brands, setBrands] = useState<string[]>([]);
    const { getBrands, isLoading, getUsecases } = useConfig()
    const [activeBrandTab, setActiveBrandTab] = useState<string>('Все бренды');
    const [activeUsecaseTab, setActiveUsecaseTab] = useState<string>('all');

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    //const [searchQuery, setSearchQuery] = useState('');

    useEffect(() => {
        const fetchComponents = async () => {
            try {
                //setLoading(true);
                setError('');

                const compatible = await fetchCompatibleComponents(selectedCategory, selectedComponents, activeUsecaseTab, activeBrandTab);
                setCompatibleComponents(compatible)

                const all = await fetchAllComponents(selectedCategory, activeUsecaseTab, activeBrandTab);
                setAllComponents(all);

                const brandsWithAll = ['Все бренды', ...getBrands(selectedCategory)];
                setBrands(brandsWithAll);

                console.log('Brands:', brandsWithAll);

            } catch (err) {
                setError(err instanceof Error ? err.message : 'Неизвестная ошибка');
            } finally {
                setLoading(false);
            }
        };

        fetchComponents();
    }, [selectedCategory, activeUsecaseTab, activeBrandTab]);

    if (error) return <div className={styles.error}>{error}</div>;

    return (
        <div className={styles.container}>
            {<div className={styles.tabs}>
                {Usecases.map(usecase => (
                    <button
                        key={usecase}
                        className={`${styles.tab} ${usecase === activeUsecaseTab ? styles.activeTab : ''}`}
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
                        className={`${styles.tab} ${brand === activeBrandTab ? styles.activeTab : ''}`}
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
                        checked={showCompatibleOnly}
                        onChange={() => setShowCompatibleOnly(!showCompatibleOnly)}
                    />
                    Совместимые товары
                </label>
            </div>

            <div >
                {loading ? (
                    <div className={styles.loading}>Загрузка...</div>
                ) : (
                    <div>
                        {(showCompatibleOnly ? сompatibleComponents || [] : allComponents || []).map((component) => (
                            <ComponentCard
                                key={component.id}
                                component={component}
                                onSelect={() =>
                                    setSelectedComponents((prev) => {
                                        const isSelected = prev[component.category]?.id === component.id;
                                        return {
                                            ...prev,
                                            [component.category]: isSelected ? null : component,
                                        };
                                    })
                                }
                                selected={selectedComponents[component.category]?.id === component.id}
                            />
                        ))}
                    </div>
                )}
            </div>


        </div>
    );
};

export default ComponentList;



const fetchAllComponents = async (
    category: string,
    activeUsecaseTab: string,
    activeBrandTab: string
) => {
    const params = new URLSearchParams();
    params.append('category', category);

    if (activeUsecaseTab !== 'all') {
        params.append('usecase', activeUsecaseTab);
    }

    if (activeBrandTab !== 'Все бренды') {
        params.append('brand', activeBrandTab);
    }

    const url = `http://localhost:8080/config/components?${params.toString()}`;

    const response = await fetch(url, {
        method: 'GET',
        headers: {
            'Accept': 'application/json',
        },
    });

    if (!response.ok) {
        throw new Error('Ошибка загрузки компонентов');
    }

    const data = await response.json();
    console.log('Response fetchAllComponents:', data); // Debugging line

    return data;
};

const fetchCompatibleComponents = async (
    category: string,
    selectedComponents: Record<string, Component | null>,
    activeUsecaseTab: string,
    activeBrandTab: string
) => {
    const url = 'http://localhost:8080/config/compatible';

    const bases = Object.entries(selectedComponents)
        .filter(([_, component]) => component !== null)
        .map(([key, component]) => {
            return {
                category: key,
                name: component!.name
            };
        });

    const requestBody: Record<string, any> = {
        category,
        bases
    };

    if (activeUsecaseTab !== 'all') {
        requestBody['usecase'] = activeUsecaseTab;
    }

    if (activeBrandTab !== 'Все бренды') {
        requestBody['brand'] = activeBrandTab;
    }



    console.log('Request body fetchCompatibleComponents:', requestBody);

    const response = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestBody)
    });

    const data = await response.json();

    console.log('Response fetchCompatibleComponents:', data);

    if (!response.ok) {
        throw new Error('Ошибка загрузки совместимых компонентов');
    }

    return data;
};