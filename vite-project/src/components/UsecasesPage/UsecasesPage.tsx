import { useEffect, useState } from "react";
import { useAuth } from "../../AuthContext";
import { Configurations, UsecaseLabels, UsecaseObject, Usecases, UsecasesResponse } from "../../types";
import styles from './UsecasesPage.module.css';
import SavedConfig from "../SavedConfig/SavedConfig";
import UsecaseConfiguration from "../UsecaseConfiguration/UsecaseConfiguration";
import { useConfig } from "../../ConfigContext";


const UsecasesPage = () => {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [UserConfigs, setUserConfigs] = useState<Configurations>([]);
    const [activeUsecaseTab, setActiveUsecaseTab] = useState<string>('office');
    const [UsecaseList, setUsecaseList] = useState<UsecaseObject[]>([]);
    const {theme } = useConfig()


    useEffect(() => {
        const fetchUsecases = async () => {
            try {
                //setLoading(true);
                setError('');

                setLoading(true);
                setError('');

                /* if (!isAuthenticated) {
                    throw new Error('Требуется авторизация');
                } */

                const response = await fetch(`http://localhost:8080/config/usecase/${activeUsecaseTab}`);

                if (!response.ok) {
                    throw new Error('Не удалось получить cборки по сценарию');
                }

                const data = await response.json();
                console.log('User configs:', data);
                setUsecaseList(data.components);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Неизвестная ошибка');
            } finally {
                setLoading(false);
            }
        };

        fetchUsecases();
    }, [activeUsecaseTab]);

    return (
        <div className={styles.container}>

            {<div className={styles.tabs}>
                {Usecases.map(usecase => {
                    if (usecase === 'all') return null;
                    return (
                        <button
                            key={usecase}
                            className={`${styles.usecaseTab} ${usecase === activeUsecaseTab ? styles.activeUsecaseTab : ''} ${styles[theme]}`}
                            onClick={() => setActiveUsecaseTab(usecase)}
                        >
                            {UsecaseLabels[usecase] || usecase}
                        </button>
                    );
                })}
            </div>}



            <div className={styles.content}>
                {loading ? (
                    <div className={styles.loading}>Загрузка...</div>
                ) : (
                    <div>
                        {!UsecaseList || UsecaseList.length === 0 ? (
                            <div className={styles.emptyMessage}>Нет подходящих сборок</div>
                        ) : (
                            UsecaseList.map((configuration) => (
                                <UsecaseConfiguration key={configuration.name} configuration={configuration} />
                            ))
                        )}
                    </div>
                )}
            </div>
        </div >
    );
}

export default UsecasesPage;