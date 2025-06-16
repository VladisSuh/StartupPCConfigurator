import { Configuration, UsecaseObject, UsecasesResponse } from "../../types";
import styles from './UsecaseConfiguration.module.css';
import SavedComponentCard from "../SaveComponentCard/SavedComponentCard";
import { useState } from "react";
import { useConfig } from "../../ConfigContext";

const UsecaseConfiguration = ({ configuration }: { configuration: UsecaseObject }) => {
    const [componentPrices, setComponentPrices] = useState<{ [id: string]: number }>({});
    const { theme } = useConfig()

    const handlePriceLoad = (componentId: string, price: number) => {
        setComponentPrices(prev => ({ ...prev, [componentId]: price }));
    };

    const totalPrice = Object.values(componentPrices).reduce((sum, price) => sum + price, 0);


    return (
        <div className={styles.configuration}>

            <div className={styles.configHeader}>
                <div className={`${styles.configTitle} ${styles[theme]}`}>
                    {configuration.name}
                </div>
            </div>

            {configuration.components.map((component, index) => (
                <div key={component.id}>
                    <SavedComponentCard
                        key={index}
                        component={component}
                        onPriceLoad={handlePriceLoad}
                    />
                </div>
            ))}

            <div className={`${styles.footer} ${styles[theme]}`}>
                <div className={styles.totalContainer}>
                    {Object.keys(componentPrices).length === configuration.components.length && (
                        <div className={styles.totalPrice}>
                            Сумма от <strong>{totalPrice.toLocaleString()} ₽</strong>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

export default UsecaseConfiguration;