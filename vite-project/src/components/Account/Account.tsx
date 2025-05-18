import { useEffect, useState } from "react";
import { useAuth } from "../../AuthContext";
import { Configurations } from "../../types";
import styles from './Account.module.css';
import SavedComponentCard from "../SaveComponentCard/SavedComponentCard";


const Account = () => {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const { isAuthenticated, getToken } = useAuth();
    const [UserConfigs, setUserConfigs] = useState<Configurations>([]);

    useEffect(() => {
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
                    throw new Error('Failed to fetch user configs');
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

        fetchUserConfigs();
    }, []);

    return (
        <div>
            <div className={styles.container}>

                <h3 className={styles.title}>Сохраненные сборки</h3>
                <div >
                    {loading ? (
                        <div className={styles.loading}>Загрузка...</div>
                    ) : (
                        <div>
                            {UserConfigs.map((Configuration) => (
                                <div key={Configuration.ID}>
                                    <div className={styles.configTitle}>
                                        {'Название конфигурации: '}{Configuration.Name}
                                    </div>
                                    {Configuration.components.map((component) => (
                                        <div key={component.id}>
                                            <SavedComponentCard
                                                key={Configuration.ID}
                                                component={component}
                                            />
                                        </div>
                                    ))}
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

export default Account;