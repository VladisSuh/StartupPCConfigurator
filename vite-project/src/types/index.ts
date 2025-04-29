
export interface Component {
    id: string;
    name: string;
    category: string;
    brand: string;
    specs: Record<string, string>;
}



export type CategoryType =
    | "cpu"
    | "gpu"
    | "motherboard"
    | "ram"
    | "storage"
    | "cooler"
    | "case"
    | "soundcard"
    | "power_supply";


export interface CategoryTabsProps {
    onSelect: (category: CategoryType) => void;
}

export const categories: CategoryType[] = [
    "cpu",
    "gpu",
    "motherboard",
    "ram",
    "storage",
    "cooler",
    "case",
    "soundcard",
    "power_supply",
];

export const categoryLabels: Record<CategoryType, string> = {
    cpu: "Процессор",
    gpu: "Видеокарта",
    motherboard: "Материнская плата",
    ram: "Оперативная память",
    storage: "Хранилище",
    cooler: "Охлаждение",
    case: "Корпус",
    soundcard: "Звуковая карта",
    power_supply: "Блок питания",
};

export interface CategoryTabsProps {
    onSelect: (category: CategoryType) => void;
}



export interface ComponentListProps {
    selectedCategory: string;
    selectedComponents: Record<string, Component | null>;
    setSelectedComponents: React.Dispatch<React.SetStateAction<Record<string, Component | null>>>;
}


export interface ComponentCardProps {
    component: Component;
    onSelect: () => void;
    selected: boolean;
}