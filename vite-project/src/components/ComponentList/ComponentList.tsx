import { useEffect, useState } from 'react';
import styles from './ComponentList.module.css';
import { ComponentCard } from '../ComponentCard/ComponentCard';
import { Component, ComponentListProps } from '../../types/index';
import Filters from '../Filters/Filters';



const ComponentList = ({
    selectedCategory,
    selectedComponents,
    setSelectedComponents,
}: ComponentListProps) => {

    const [allComponents, setAllComponents] = useState<Component[]>([]);
    const [сompatibleComponents, setCompatibleComponents] = useState<Component[]>([]);

    const [activeBrandTab, setActiveBrandTab] = useState<string>('Все бренды');
    const [activeUsecaseTab, setActiveUsecaseTab] = useState<string>('all');
    const [showCompatibleOnly, setShowCompatibleOnly] = useState(true);

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [componentPrices, setComponentPrices] = useState<Record<string, number>>({});

    const [filteredCompatibleComponents, setFilteredCompatibleComponents] = useState<Component[]>([]);
    const [filteredAllComponents, setFilteredAllComponents] = useState<Component[]>([]);
    

    const saveComponentPrice = (componentId: string, price: number) => {
        setComponentPrices(prev => ({
            ...prev,
            [componentId]: price
        }));
    };

    useEffect(() => {
        setActiveBrandTab('Все бренды');
        setActiveUsecaseTab('all');
    }, [selectedCategory]);


    useEffect(() => {
        const fetchComponents = async () => {
            try {
                setError('');
                console.log('Fetching components for category:', selectedCategory);

                const compatible = await fetchCompatibleComponents(selectedCategory, selectedComponents, activeUsecaseTab, activeBrandTab);
                setCompatibleComponents(compatible)

                const all = await fetchAllComponents(selectedCategory, activeUsecaseTab, activeBrandTab);
                setAllComponents(all);
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
            <Filters
                сompatibleComponents={сompatibleComponents}
                allComponents={allComponents}
                setFilteredCompatibleComponents={setFilteredCompatibleComponents}
                setFilteredAllComponents={setFilteredAllComponents}
                componentPrices={componentPrices}
                activeBrandTab={activeBrandTab}
                setActiveBrandTab={setActiveBrandTab}
                activeUsecaseTab={activeUsecaseTab}
                setActiveUsecaseTab={setActiveUsecaseTab}
                selectedCategory={selectedCategory}
                showCompatibleOnly={showCompatibleOnly}
                setShowCompatibleOnly={setShowCompatibleOnly}
            />

            <div >
                {loading ? (
                    <div className={styles.loading}>Загрузка...</div>
                ) : (
                    <div>
                        {(showCompatibleOnly ? filteredCompatibleComponents : filteredAllComponents || []).map((component) => (
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
                                onPriceLoaded={(price) => saveComponentPrice(component.id, price)}
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
    console.log('Response fetchAllComponents:', data);

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