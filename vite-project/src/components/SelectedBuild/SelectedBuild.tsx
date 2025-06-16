import { useState } from 'react';
import { useAuth } from '../../AuthContext';
import { CategoryType, Component, SelectedBuildProps } from '../../types';
import Login from '../Login/component';
import { Modal } from '../Modal/Modal';
import Register from '../Register/component';
import SelectedComponentList from '../SelectedComponentList/SelectedComponentList';
import styles from './SelectedBuild.module.css';
import { useConfig } from '../../ConfigContext';



export const SelectedBuild = ({ selectedComponents, setSelectedComponents }: SelectedBuildProps) => {

    const { isAuthenticated, getToken } = useAuth();
    const [isVisible, setIsVisible] = useState(false);
    const [openComponent, setOpenComponent] = useState('login');
    const [showLoginMessage, setShowLoginMessage] = useState(false);
    const [isSaveResponseVisible, setIsSaveResponseVisible] = useState(false);
    const [isSaveError, setIsSaveError] = useState(false);
    const [ErrorMessage, setErrorMessage] = useState('');
    const [NameInputVisible, setNameInputVisible] = useState(false);
    const [buildName, setBuildName] = useState('');
    const { theme } = useConfig();


    const selectedList = Object.entries(selectedComponents).filter(
        ([_, component]) => component !== null
    ) as [CategoryType, Component][];

    const handleRemove = (category: CategoryType) => {
        setSelectedComponents((prev) => ({
            ...prev,
            [category]: null,
        }));
    };

    const handleSave = async () => {
        const token = getToken();

        if (!buildName) return;

        const components = selectedList.map(([category, component]) => ({
            category,
            name: component.name,
        }));

        console.log({
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                name: buildName,
                components,
            }),
        })

        try {
            const response = await fetch('http://localhost:8080/config/newconfig', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    name: buildName,
                    components,
                }),
            });

            const responseJson = await response.json(); // читаем один раз

            if (!response.ok) {
                console.error('Ошибка сервера:', response.status, response.statusText);
                if (responseJson.error) {
                    const str = responseJson.error.charAt(0).toUpperCase() + responseJson.error.slice(1)
                    setErrorMessage(str);
                }
                console.error('Тело ответа:', await response.text());
                return;
            }

            console.log('Конфигурация сохранена:', responseJson);
        } catch (error) {
            setIsSaveError(true);
            console.error(error);
        }
        setIsSaveResponseVisible(true);
    };

    return (
        <div className={styles.selectedBuildContainer}>
            <div className={styles.summary}>
                <p className={styles.title}>Итоговая сборка</p>
                {selectedList.length === 0 ? (
                    <p className={styles.empty}>Компоненты не выбраны</p>
                ) : (
                    <SelectedComponentList selectedComponents={selectedList} onRemove={handleRemove} />
                )}
            </div>
            <div className={`${styles.saveButton} ${styles[theme]}`} onClick={() => {
                if (!isAuthenticated) {
                    setIsVisible(true);
                    setOpenComponent('login');
                    setShowLoginMessage(true);
                }else{
                    setNameInputVisible(true);
                }
            }}>
                Сохранить в личном кабинете
            </div>

            <Modal isOpen={isVisible} onClose={() => setIsVisible(false)}>
                {openComponent === 'register' ? (
                    <Register
                        setOpenComponent={(component) => {
                            setOpenComponent(component);
                            setShowLoginMessage(false);
                        }}
                        onClose={() => setIsVisible(false)}
                    />
                ) : (
                    <Login
                        setOpenComponent={(component) => {
                            setOpenComponent(component);
                            setShowLoginMessage(false);
                        }}
                        onClose={() => setIsVisible(false)}
                        message={showLoginMessage ? 'Сохранить сборку можно после авторизации' : undefined}
                    />
                )}
            </Modal>

            <Modal isOpen={NameInputVisible} onClose={() => {
                setNameInputVisible(false);
            }}>
                <h3>Название вашей сборки</h3>
                <input
                    type="text"
                    value={buildName}
                    onChange={(e) => setBuildName(e.target.value)}
                    placeholder="Например: 'Игровой ПК'"
                    className={`${styles.input} ${styles[theme]}`}
                    autoFocus
                />
                <button className={styles.submitButton} disabled={buildName === ''} onClick={() => {
                    setNameInputVisible(false);
                    handleSave();
                    setBuildName('');
                }}>Сохранить</button>
            </Modal>

            <Modal isOpen={isSaveResponseVisible} onClose={() => {
                setIsSaveResponseVisible(false)
                setIsSaveError(false);
                setErrorMessage('');
            }}>
                {isSaveError ? (
                    <>
                        <h3>Произошла ошибка при сохранении конфигурации.</h3>
                        {ErrorMessage ? <p>{ErrorMessage}</p> : <p>Попробуйте еще раз.</p>}
                    </>
                ) : (
                    <>
                        <p>Конфигурация успешно сохранена!</p>
                    </>
                )}
            </Modal>

        </div>
    );
};

