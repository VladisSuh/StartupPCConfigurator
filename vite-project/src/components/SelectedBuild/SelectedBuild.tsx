import { useState } from 'react';
import { useAuth } from '../../AuthContext';
import { Component } from '../../types';
import Login from '../Login/component';
import { Modal } from '../Modal/Modal';
import Register from '../Register/component';
import SelectedComponentList from '../SelectedComponentList/SelectedComponentList';
import styles from './SelectedBuild.module.css';

interface SelectedBuildProps {
    selectedComponents: Record<string, Component | null>;
    setSelectedComponents: React.Dispatch<React.SetStateAction<Record<string, Component | null>>>;
}

export const SelectedBuild = ({ selectedComponents, setSelectedComponents }: SelectedBuildProps) => {

    const { isAuthenticated, getToken } = useAuth();
    const [isVisible, setIsVisible] = useState(false);
    const [openComponent, setOpenComponent] = useState('login');
    const [showLoginMessage, setShowLoginMessage] = useState(false);
    const [isSaveResponseVisible, setIsSaveResponseVisible] = useState(false);
    const [isSaveError, setIsSaveError] = useState(false);
    const [ErrorMessage, setErrorMessage] = useState('');


    const selectedList = Object.entries(selectedComponents).filter(
        ([_, component]) => component !== null
    ) as [string, Component][];

    const handleRemove = (category: string) => {
        setSelectedComponents((prev) => ({
            ...prev,
            [category]: null,
        }));
    };

    const handleSave = async () => {
        if (!isAuthenticated) {
            setIsVisible(true);
            setOpenComponent('login');
            setShowLoginMessage(true);
        }
        const token = getToken();

        const configName = prompt('Введите название конфигурации:');
        if (!configName) return;

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
                name: configName,
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
                    name: configName,
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
                //alert(`Ошибка: ${responseJson.message || 'Неизвестная ошибка'}`);
                return;
            }

            console.log('Конфигурация сохранена:', responseJson);
            //alert('Конфигурация успешно сохранена!');
        } catch (error) {
            setIsSaveError(true);
            console.error(error);
            //alert('Не удалось сохранить конфигурацию');
        }
        setIsSaveResponseVisible(true);
    };

    return (
        <div className={styles.buildContainer}>
            <div className={styles.summary}>
                <h2 className={styles.title}>Итоговая сборка</h2>
                {selectedList.length === 0 ? (
                    <p className={styles.empty}>Компоненты не выбраны</p>
                ) : (
                    <SelectedComponentList selectedComponents={selectedList} onRemove={handleRemove} />
                )}
            </div>

            <div className={styles.button} onClick={handleSave}>
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

            <Modal isOpen={isSaveResponseVisible} onClose={() => setIsSaveResponseVisible(false)}>
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

