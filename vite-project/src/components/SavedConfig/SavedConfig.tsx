import { Configuration } from "../../types";
import styles from './SavedConfig.module.css';
import SavedComponentCard from "../SaveComponentCard/SavedComponentCard";
import { useAuth } from "../../AuthContext";
import { useState } from "react";

const SavedConfig = ({ configuration, onDelete }: { configuration: Configuration, onDelete?: () => void; }) => {
    const { isAuthenticated, getToken } = useAuth();
    const [componentPrices, setComponentPrices] = useState<{ [id: string]: number }>({});

    const handlePriceLoad = (componentId: string, price: number) => {
        setComponentPrices(prev => ({ ...prev, [componentId]: price }));
    };

    const totalPrice = Object.values(componentPrices).reduce((sum, price) => sum + price, 0);

    function handleDeleteConfiguration(): void {
        const deleteConfiguration = async () => {
            try {
                if (!isAuthenticated) {
                    throw new Error('Требуется авторизация');
                }

                const token = getToken();
                const response = await fetch(`http://localhost:8080/config/newconfig/${configuration.ID}`, {
                    method: 'DELETE',
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                });

                if (!response.ok) {
                    throw new Error('Failed to fetch user configs');
                }

                onDelete && onDelete();

            } catch (err) {
                console.error('Error fetching user configs:', err);
            }
        };

        deleteConfiguration();

    }

    return (
        <div className={styles.configuration}>

            <div className={styles.configHeader}>
                <div className={styles.configTitle}>
                    {'Название: '}<span className={styles.configName}>{configuration.Name}</span>
                </div>
                {onDelete && <img
                    src="src/assets/icon-delete.png"
                    className={styles.trashIcon}
                    alt="Удалить"
                    onClick={handleDeleteConfiguration}
                />}
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

            <div className={styles.footer}>
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

export default SavedConfig;