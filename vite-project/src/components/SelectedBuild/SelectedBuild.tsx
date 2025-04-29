import { Component } from '../../types';
import SelectedComponentList from '../SelectedComponentList/SelectedComponentList';
import styles from './SelectedBuild.module.css';

interface SelectedBuildProps {
    selectedComponents: Record<string, Component | null>;
    setSelectedComponents: React.Dispatch<React.SetStateAction<Record<string, Component | null>>>;
}

export const SelectedBuild = ({ selectedComponents, setSelectedComponents }: SelectedBuildProps) => {
    const selectedList = Object.entries(selectedComponents).filter(
        ([_, component]) => component !== null
    ) as [string, Component][];

    const handleRemove = (category: string) => {
        setSelectedComponents((prev) => ({
            ...prev,
            [category]: null,
        }));
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

            <div className={styles.button}>
                Сохранить в личном кабинете
            </div>
        </div>
    );
};
