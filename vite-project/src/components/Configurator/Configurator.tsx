import { useState } from "react";
import CategoryTabs from "../CategoryTabs/CategoryTabs";
import ComponentList from "../ComponentList/ComponentList";
import styles from "./Configurator.module.css";
import { SelectedBuild } from "../SelectedBuild/SelectedBuild";
import { CategoryType, Component } from "../../types/index";
import { useConfig } from "../../ConfigContext";

const Configurator = () => {
    const [selectedCategory, setSelectedCategory] = useState<CategoryType>("cpu");
    const [selectedComponents, setSelectedComponents] = useState<Record<string, Component | null>>({
        cpu: null,
        gpu: null,
        motherboard: null,
        ram: null,
        hdd: null,
        ssd: null,
        cooler: null,
        case: null,
        psu: null,
    });

    const { isLoading } = useConfig();

    if (isLoading) return <div>Загрузка...</div>;

    return (
        <div className={styles.container}>
            <CategoryTabs
                onSelect={setSelectedCategory}
                selectedComponents={selectedComponents}
            />

            <ComponentList
                selectedCategory={selectedCategory}
                selectedComponents={selectedComponents}
                setSelectedComponents={setSelectedComponents}
            />

            <SelectedBuild
                selectedComponents={selectedComponents}
                setSelectedComponents={setSelectedComponents}
            />
        </div>
    );
}

export default Configurator;