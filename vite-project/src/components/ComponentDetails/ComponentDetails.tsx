import { useConfig } from "../../ConfigContext";
import { Component, specs } from "../../types";
import styles from "./ComponentDetails.module.css";


const ComponentDetails = ({ component }: { component: Component }) => {
    const { theme } = useConfig();
    
    return (
        <div className={`${styles.componentDetails} ${styles[theme]}`} >
            <h2 className={styles.componentTitle}>{component.name}</h2>

            <div className={styles.componentMeta}>
                <span className={styles.componentBrand}>Бренд: {component.brand}</span>
            </div>

            <div className={styles.componentSpecs}>
                <h3>Характеристики:</h3>
                <ul className="specs-list">
                    {Object.entries(component.specs).map(([key, value]) => (
                        <li key={key} className="spec-item">
                            <span className={styles.specName}>{specs[key] || key}: </span>
                            <span className={styles.specValue}>{value}</span>
                        </li>
                    ))}
                </ul>
            </div>
        </div>
    );
};

export default ComponentDetails;