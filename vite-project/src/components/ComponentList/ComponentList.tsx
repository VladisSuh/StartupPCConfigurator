import { useEffect, useState } from 'react';
import styles from './ComponentList.module.css';
import { ComponentCard } from '../ComponentCard/ComponentCard';
//import { mockComponents } from '../mockData/mock';
import { Component, ComponentListProps } from '../../types/index';



const ComponentList = ({
    selectedCategory,
    selectedComponents,
    setSelectedComponents,
}: ComponentListProps) => {
    const [allComponents, setAllComponents] = useState<Component[]>([]);
    const [сompatibleComponents, setCompatibleComponents] = useState<Component[]>([]);
    const [showCompatibleOnly, setShowCompatibleOnly] = useState(true);

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    //const [searchQuery, setSearchQuery] = useState('');

    useEffect(() => {
        const fetchComponents = async () => {
            try {
                setLoading(true);
                setError('');

                const all = await fetchAllComponents(selectedCategory);
                setAllComponents(all);
                const compatible = await fetchCompatibleComponents(selectedCategory, selectedComponents)

                setCompatibleComponents(compatible)

            } catch (err) {
                setError(err instanceof Error ? err.message : 'Неизвестная ошибка');
            } finally {
                setLoading(false);
            }
        };

        fetchComponents();
    }, [selectedCategory]);

    if (error) return <div className={styles.error}>{error}</div>;

    return (
        <div className={styles.container}>

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



const fetchAllComponents = async (category: string) => {
    let url = `http://localhost:8080/config/compatible?category=${encodeURIComponent(category)}`;


    const response = await fetch(url);

    let d = await response.json()

    console.log('Response fetchAllComponents:', d); // Debugging line

    if (!response.ok) {
        throw new Error('Ошибка загрузки компонентов');
    }
    return d;
};

const fetchCompatibleComponents = async (
    category: string,
    selectedComponents: Record<string, Component | null>
) => {
    let url = `http://localhost:8080/config/compatible?category=${encodeURIComponent(category)}`;

    Object.entries(selectedComponents).forEach(([category, component]) => {
        if (component !== null) {
            if (category=='cpu'){
                url += `&${category}=${encodeURIComponent(component.specs.socket)}`
            }else{
                url += `&${category}=${encodeURIComponent(component.name)}`
            }
            
        }
    });

    console.log('url', url)

    const response = await fetch(url);
    let d = await response.json()

    console.log('Response fetchCompatibleComponents:', d);
    if (!response.ok) {
        throw new Error('Ошибка загрузки совместимых компонентов');
    }
    return d;
};
