
export interface Component {
    id: string;
    name: string;
    category: CategoryType;
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
    selectedCategory: CategoryType;
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

export type Page = 'configurator' | 'account' | 'usecases';


export type ResponseComponent = {
    category: CategoryType;
    id: string;
    name: string;
    specs: Record<string, string>;
};

export type Configuration = {
    ID: string;
    Name: string;
    OwnerId: string;
    components: Component[];
    CreatedAt: string;
    UpdatedAt: string;
};

export type Configurations = Configuration[];

export type UsecaseObject = {
    name: string;
    components: Component[];
};

export type UsecasesResponse = {
    components: UsecaseObject[];
}

export type UsecaseType =
    | 'all'
    | 'office'
    | 'htpc'
    | 'gaming'
    | 'streamer'
    | 'design'
    | 'video'
    | 'cad'
    | 'dev'
    | 'enthusiast'
    | 'nas';

export const Usecases: UsecaseType[] = [
    'all',
    "office",
    "htpc",
    "gaming",
    "streamer",
    "design",
    "video",
    "cad",
    "dev",
    "enthusiast",
    "nas"
];

export const UsecaseLabels: Record<UsecaseType, string> = {
    all: "Все сценарии",
    office: "Офис",
    htpc: "Медиаплеер",
    gaming: "Игры",
    streamer: "Стриминг",
    design: "Дизайн",
    video: "Видеомонтаж",
    cad: "3D-моделирование",
    dev: "Разработка",
    enthusiast: "Топовая сборка",
    nas: "Домашний сервер",
};


export type Notification = {
    id: string;
    userId: string;
    title: string;
    message: string;
    isRead: boolean;
    createdAt: string; 
}
