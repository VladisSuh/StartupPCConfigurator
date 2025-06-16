import { categoryLabels, SelectedComponentListProps } from '../../types/index';
import styles from './SelectedComponentList.module.css';

const SelectedComponentList = ({ selectedComponents, onRemove }: SelectedComponentListProps) => {
    return (
        <div className={styles.list}>
            {selectedComponents.map(([category, component]) => (
                <div key={category} className={styles.item}>
                    <div>
                        <p className={styles.category}>{categoryLabels[category] || category}</p>
                        <p className={styles.name}>{component.name}</p>
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