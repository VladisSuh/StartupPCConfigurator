export const mockComponents = [
    // CPU
    {
        id: "cpu-1",
        name: "AMD Ryzen 9 7950X",
        category: "cpu",
        brand: "AMD",
        specs: {
            cores: 16,
            threads: 32,
            base_clock: "4.5 GHz",
            boost_clock: "5.7 GHz",
            tdp: "170W",
            socket: "AM5"
        }
    },
    {
        id: "cpu-2",
        name: "Intel Core i9-13900K",
        category: "cpu",
        brand: "Intel",
        specs: {
            cores: 24,
            threads: 32,
            base_clock: "3.0 GHz",
            boost_clock: "5.8 GHz",
            tdp: "125W",
            socket: "LGA1700"
        }
    },

    // GPU
    {
        id: "gpu-1",
        name: "NVIDIA GeForce RTX 4090",
        category: "gpu",
        brand: "NVIDIA",
        specs: {
            memory: "24GB GDDR6X",
            bus_width: "384-bit",
            core_clock: "2235 MHz",
            tdp: "450W",
            ports: "3x DisplayPort, 1x HDMI"
        }
    },
    {
        id: "gpu-2",
        name: "AMD Radeon RX 7900 XTX",
        category: "gpu",
        brand: "AMD",
        specs: {
            memory: "24GB GDDR6",
            bus_width: "384-bit",
            core_clock: "2300 MHz",
            tdp: "355W",
            ports: "2x DisplayPort, 1x HDMI"
        }
    },

    // Motherboard
    {
        id: "mb-1",
        name: "ASUS ROG Maximus Z790 Hero",
        category: "motherboard",
        brand: "ASUS",
        specs: {
            socket: "LGA1700",
            chipset: "Intel Z790",
            form_factor: "ATX",
            ram_slots: 4,
            m2_slots: 5
        }
    },
    {
        id: "mb-2",
        name: "MSI MAG B650 Tomahawk",
        category: "motherboard",
        brand: "MSI",
        specs: {
            socket: "AM5",
            chipset: "AMD B650",
            form_factor: "ATX",
            ram_slots: 4,
            m2_slots: 4
        }
    },

    // RAM
    {
        id: "ram-1",
        name: "Corsair Dominator Platinum RGB 32GB",
        category: "ram",
        brand: "Corsair",
        specs: {
            capacity: "32GB (2x16GB)",
            speed: "DDR5-6000",
            timings: "CL36",
            voltage: "1.35V",
            rgb: true
        }
    },
    {
        id: "ram-2",
        name: "Kingston Fury Beast 64GB Kit",
        category: "ram",
        brand: "Kingston",
        specs: {
            capacity: "64GB (2x32GB)",
            speed: "DDR4-3200",
            timings: "CL16",
            voltage: "1.35V",
            rgb: false
        }
    },

    // Storage
    {
        id: "ssd-1",
        name: "Samsung 990 Pro 2TB",
        category: "storage",
        brand: "Samsung",
        specs: {
            capacity: "2TB",
            type: "NVMe PCIe 4.0",
            read_speed: "7450 MB/s",
            write_speed: "6900 MB/s",
            endurance: "1200 TBW"
        }
    },
    {
        id: "ssd-2",
        name: "Western Digital Black SN850X 1TB",
        category: "storage",
        brand: "WD",
        specs: {
            capacity: "1TB",
            type: "NVMe PCIe 4.0",
            read_speed: "7300 MB/s",
            write_speed: "6300 MB/s",
            endurance: "600 TBW"
        }
    },

    // Cooler
    {
        id: "cooler-1",
        name: "Noctua NH-D15 Chromax Black",
        category: "cooler",
        brand: "Noctua",
        specs: {
            type: "Air Cooler",
            height: "165mm",
            fans: 2,
            noise_level: "24.6 dB(A)",
            tdp: "220W"
        }
    },
    {
        id: "cooler-2",
        name: "Cooler Master MasterLiquid ML360R",
        category: "cooler",
        brand: "Cooler Master",
        specs: {
            type: "AIO Liquid Cooler",
            radiator_size: "360mm",
            fans: 3,
            noise_level: "30 dB(A)",
            tdp: "300W"
        }
    },

    // Case
    {
        id: "case-1",
        name: "NZXT H9 Elite",
        category: "case",
        brand: "NZXT",
        specs: {
            form_factor: "Full Tower",
            motherboard_support: "E-ATX/ATX/mATX",
            fans: 4,
            rgb: true,
            weight: "12.5 kg"
        }
    },
    {
        id: "case-2",
        name: "Fractal Design North",
        category: "case",
        brand: "Fractal Design",
        specs: {
            form_factor: "Mid Tower",
            motherboard_support: "ATX/mATX",
            fans: 2,
            rgb: false,
            weight: "8.2 kg"
        }
    },

    // Sound Card
    {
        id: "sound-1",
        name: "Creative Sound Blaster AE-7",
        category: "soundcard",
        brand: "Creative",
        specs: {
            channels: "7.1",
            snr: "127dB",
            sample_rate: "384 kHz",
            interface: "PCIe",
            dac: "ESS SABRE-class"
        }
    },
    {
        id: "sound-2",
        name: "ASUS Xonar AE",
        category: "soundcard",
        brand: "ASUS",
        specs: {
            channels: "7.1",
            snr: "110dB",
            sample_rate: "192 kHz",
            interface: "PCIe",
            dac: "Cirrus Logic CS4382"
        }
    },

    // Power Supply
    {
        id: "psu-1",
        name: "Seasonic PRIME TX-1000",
        category: "power_supply",
        brand: "Seasonic",
        specs: {
            wattage: "1000W",
            efficiency: "80+ Titanium",
            modular: "Full",
            connectors: "1x 24-pin, 2x EPS, 6x PCIe",
            warranty: "12 years"
        }
    },
    {
        id: "psu-2",
        name: "EVGA SuperNOVA 850 G6",
        category: "power_supply",
        brand: "EVGA",
        specs: {
            wattage: "850W",
            efficiency: "80+ Gold",
            modular: "Full",
            connectors: "1x 24-pin, 2x EPS, 4x PCIe",
            warranty: "10 years"
        }
    }
];