
export interface Component {
    id: string;
    name: string;
    category: string;
    brand: string;
    specs: Record<string, string>;
}



/* export type CategoryType =
    | "cpu"
    | "gpu"
    | "motherboard"
    | "ram"
    | "storage"
    | "cooler"
    | "case"
    | "soundcard"
    | "power_supply";

 */
export interface CategoryTabsProps {
    onSelect: (category: CategoryType) => void;
    selectedComponents: Record<string, Component | null>;
}

/* export const categories: CategoryType[] = [
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
}; */

export type CategoryType =
    | 'cpu'
    | 'gpu'
    | 'motherboard'
    | 'ram'
    | 'hdd'
    | 'ssd'
    | 'cooler'
    | 'case'
    | 'psu';

export const categories: CategoryType[] = [
    "cpu",
    "gpu",
    "motherboard",
    "ram",
    'hdd',
    'ssd',
    "cooler",
    "case",
    'psu'
];

export const categoryLabels: Record<CategoryType, string> = {
    cpu: "Процессор",
    gpu: "Видеокарта",
    motherboard: "Материнская плата",
    ram: "Оперативная память",
    hdd: "Жёсткий диск",
    ssd: "Твердотельный накопитель",
    cooler: "Охлаждение",
    case: "Корпус",
    psu: "Блок питания",
};

export const specs: Record<string, string> = {
    cooler_height: "Высота кулера",
    cores: "Количество ядер",
    socket: "Сокет",
    tdp: "TDP",
    threads: "Количество потоков",
    length_mm: "Длина (мм)",
    memory_gb: "Оперативная память (ГБ)",
    power_draw: "Потребляемая мощность",
    form_factor: "Форм-фактор",
    ram_type: "Тип ОЗУ",
    capacity: "Ёмкость",
    frequency: "Частота",
    cooler_max_height: "Максимальная высота кулера",
    gpu_max_length: "Максимальная длина GPU",
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

export interface Offer {
    shopName: string;
    price: number;
    availability: string;
    url: string;
}

export interface OffersListProps {
    offers: Offer[];
}