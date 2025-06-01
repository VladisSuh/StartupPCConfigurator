import { useEffect, useState } from "react";
import { useAuth } from "../../AuthContext";
import { Configurations } from "../../types";
import styles from './Account.module.css';
import SavedComponentCard from "../SaveComponentCard/SavedComponentCard";
import SavedConfig from "../SavedConfig/SavedConfig";


const Account = () => {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const { isAuthenticated, getToken } = useAuth();
    const [UserConfigs, setUserConfigs] = useState<Configurations>([]);

    const fetchUserConfigs = async () => {
        try {
            setLoading(true);
            setError('');

            if (!isAuthenticated) {
                throw new Error('Требуется авторизация');
            }

            const token = getToken();
            const response = await fetch('http://localhost:8080/config/userconf', {
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            });

            if (!response.ok) {
                throw new Error('Не удалось получить конфигурации');
            }

            const data = await response.json();
            console.log('User configs:', data);
            setUserConfigs(data);

        } catch (err) {
            console.error('Error fetching user configs:', err);
            setError(err instanceof Error ? err.message : 'Неизвестная ошибка');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchUserConfigs();
    }, []);

    return (
        <div className={styles.container}>

            <h3 className={styles.title}>Сохраненные сборки</h3>
            <div className={styles.content}>
                {loading ? (
                    <div className={styles.loading}>Загрузка...</div>
                ) : (
                    <div>
                        {!UserConfigs || UserConfigs.length === 0 ? (
                            <div className={styles.emptyMessage}>У вас нет сохранённых сборок</div>
                        ) : (
                            UserConfigs.map((configuration) => (
                                <SavedConfig
                                    key={configuration.ID}
                                    configuration={configuration}
                                    onDelete={fetchUserConfigs}
                                />
                            ))
                        )}
                    </div>
                )}
            </div>
        </div >
    );
}

export default Account;