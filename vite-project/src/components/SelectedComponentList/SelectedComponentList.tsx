import { categoryLabels } from '../../types/index';
import { CategoryType } from '../../types/index';
import { Component } from '../../types';
import styles from './SelectedComponentList.module.css';

interface SelectedComponentListProps {
    selectedComponents: [string, Component][];
    onRemove: (category: string) => void;
}

const SelectedComponentList = ({ selectedComponents, onRemove }: SelectedComponentListProps) => {
    return (
        <div className={styles.list}>
            {selectedComponents.map(([category, component]) => (
                <div key={category} className={styles.item}>
                    <div className={styles.text}>
                        <strong>{categoryLabels[category as CategoryType] || category}:</strong>{' '}
                        {component.name}
                    </div>
                    <button
                        className={styles.removeButton}
                        onClick={() => onRemove(category)}
                    >
                        &times;
                    </button>
                </div>
            ))}
        </div>
    );
};

export default SelectedComponentList;