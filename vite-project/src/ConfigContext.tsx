import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { CategoryType, categories } from './types';

interface ConfigContextType {
    getUsecases: () => string[];
    getBrands: (category: CategoryType) => string[];
    isLoading: boolean;
}

const ConfigContext = createContext<ConfigContextType | undefined>(undefined);

export const useConfig = () => {
    const context = useContext(ConfigContext);
    if (!context) {
        throw new Error('useConfig должен использоваться внутри ConfigProvider');
    }
    return context;
};

export const ConfigProvider = ({ children }: { children: ReactNode }) => {
    const [usecases, setUsecases] = useState<string[]>([]);
    const [brands, setBrands] = useState<Record<CategoryType, string[]>>({
        cpu: [], gpu: [], motherboard: [], ram: [],
        hdd: [], ssd: [], cooler: [], case: [], psu: []
    });
    const [isLoading, setIsLoading] = useState(true);
    const getBrands = (category: CategoryType) => {
        return brands[category] || [];
    };
    const getUsecases = () => {
        return usecases;
    };

    useEffect(() => {
        const fetchUsecases = async () => {
            try {
                const res = await fetch('http://localhost:8080/config/usecases');
                if (!res.ok) throw new Error(`Ошибка: ${res.status}`);
                const data = await res.json();
                console.log('Сценарии', data);

                setUsecases(data);
            } catch (e) {
                console.error('Ошибка при загрузке сценариев', e);
            }
        };

        const fetchBrands = async () => {
            const all: Record<CategoryType, string[]> = { ...brands };
            for (const cat of categories) {
                try {
                    console.log('Загрузка брендов', cat);
                    const res = await fetch(`http://localhost:8080/config/brands?category=${cat}`);
                    if (!res.ok) throw new Error(`Ошибка: ${res.status}`);
                    const data = await res.json();
                    all[cat] = data.brands;
                } catch (e) {
                    console.error(`Ошибка при загрузке брендов ${cat}`, e);
                }
            }
            setBrands(all);
        };

        const load = async () => {
            await fetchUsecases();
            await fetchBrands();
            setIsLoading(false);
        };

        load();
    }, []);

    return (
        <ConfigContext.Provider value={{ getUsecases, getBrands, isLoading }}>
            {children}
        </ConfigContext.Provider>
    );
};


