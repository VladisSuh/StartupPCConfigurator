package rules

// Расширяем ScenarioRule новыми полями Min/Max для CPU-TDP, RAM, GPU-памяти и PSU
type ScenarioRule struct {
	CPUSocketWhitelist             []string
	MinCPUTDP, MaxCPUTDP           int
	RAMType                        string
	MinRAM, MaxRAM                 int
	MinGPUMemory, MaxGPUMemory     int
	MinPSUPower, MaxPSUPower       int
	MinHDDCapacity, MaxHDDCapacity int // ёмкость HDD в ГБ
	CaseFormFactors                []string
	MinSSDThroughput               int      // минимальная пропускная способность, МБ/с
	SSDFormFactors                 []string // допустимые форм-факторы: "M.2", "2.5", и т.д.
}

var ScenarioRules = map[string]ScenarioRule{
	"office": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700"},
		MinCPUTDP:          0, MaxCPUTDP: 35,
		RAMType: "DDR4",
		MinRAM:  4, MaxRAM: 16,
		MinGPUMemory: 0, MaxGPUMemory: 4,
		MinPSUPower: 150, MaxPSUPower: 300,
		CaseFormFactors:  []string{"Mini-ITX", "Micro-ATX"},
		MinSSDThroughput: 0,
		SSDFormFactors:   []string{"M.2", "2.5"},
		MinHDDCapacity:   0, MaxHDDCapacity: 0,
	},

	"gaming": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700", "AM5"},
		MinCPUTDP:          65, MaxCPUTDP: 125,
		RAMType: "DDR4",
		MinRAM:  16, MaxRAM: 64,
		MinGPUMemory: 6, MaxGPUMemory: 16,
		MinPSUPower: 650, MaxPSUPower: 1000,
		CaseFormFactors:  []string{"ATX", "Micro-ATX"},
		MinSSDThroughput: 3000,
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   0, MaxHDDCapacity: 0,
	},

	"htpc": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700"},
		MinCPUTDP:          0, MaxCPUTDP: 35,
		RAMType: "DDR4",
		MinRAM:  8, MaxRAM: 16,
		MinGPUMemory: 0, MaxGPUMemory: 4,
		MinPSUPower: 300, MaxPSUPower: 500,
		CaseFormFactors:  []string{"Mini-ITX"},
		MinSSDThroughput: 0,
		SSDFormFactors:   []string{"M.2", "2.5"},
		MinHDDCapacity:   500, MaxHDDCapacity: 8000, // фильмы и музыка
	},

	"streamer": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700", "AM5"},
		MinCPUTDP:          65, MaxCPUTDP: 125,
		RAMType: "DDR4",
		MinRAM:  32, MaxRAM: 64,
		MinGPUMemory: 6, MaxGPUMemory: 12,
		MinPSUPower: 750, MaxPSUPower: 1200,
		CaseFormFactors:  []string{"ATX", "Micro-ATX"},
		MinSSDThroughput: 2000,
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   0, MaxHDDCapacity: 0,
	},

	"design": {
		CPUSocketWhitelist: []string{"AM5"},
		MinCPUTDP:          95, MaxCPUTDP: 170,
		RAMType: "DDR5",
		MinRAM:  32, MaxRAM: 128,
		MinGPUMemory: 10, MaxGPUMemory: 24,
		MinPSUPower: 650, MaxPSUPower: 1200,
		CaseFormFactors:  []string{"ATX"},
		MinSSDThroughput: 2000,
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   256, MaxHDDCapacity: 8000, // большие проекты и библиотека
	},

	"video": {
		CPUSocketWhitelist: []string{"AM5"},
		MinCPUTDP:          95, MaxCPUTDP: 170,
		RAMType: "DDR5",
		MinRAM:  64, MaxRAM: 128,
		MinGPUMemory: 10, MaxGPUMemory: 24,
		MinPSUPower: 750, MaxPSUPower: 1200,
		CaseFormFactors:  []string{"ATX"},
		MinSSDThroughput: 3000,
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   1000, MaxHDDCapacity: 16000, // видеопроекты под архив
	},

	"cad": {
		CPUSocketWhitelist: []string{"AM5", "LGA1700"},
		MinCPUTDP:          95, MaxCPUTDP: 170,
		RAMType: "DDR5",
		MinRAM:  64, MaxRAM: 128,
		MinGPUMemory: 4, MaxGPUMemory: 12,
		MinPSUPower: 650, MaxPSUPower: 1000,
		CaseFormFactors:  []string{"ATX"},
		MinSSDThroughput: 2500,
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   0, MaxHDDCapacity: 0,
	},

	"dev": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700"},
		MinCPUTDP:          35, MaxCPUTDP: 95,
		RAMType: "DDR4",
		MinRAM:  16, MaxRAM: 64,
		MinGPUMemory: 0, MaxGPUMemory: 8,
		MinPSUPower: 550, MaxPSUPower: 800,
		CaseFormFactors:  []string{"ATX", "Micro-ATX", "Mini-ITX"},
		MinSSDThroughput: 1000,
		SSDFormFactors:   []string{"M.2", "2.5"},
		MinHDDCapacity:   0, MaxHDDCapacity: 8000,
	},

	"enthusiast": {
		CPUSocketWhitelist: []string{"AM5", "LGA1700"},
		MinCPUTDP:          65, MaxCPUTDP: 170,
		RAMType: "DDR5",
		MinRAM:  32, MaxRAM: 128,
		MinGPUMemory: 8, MaxGPUMemory: 24,
		MinPSUPower: 650, MaxPSUPower: 1500,
		CaseFormFactors:  []string{"ATX"},
		MinSSDThroughput: 4000,
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   0, MaxHDDCapacity: 0,
	},

	"nas": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700"},
		MinCPUTDP:          0, MaxCPUTDP: 65,
		RAMType: "DDR4",
		MinRAM:  8, MaxRAM: 32,
		MinGPUMemory: 0, MaxGPUMemory: 4,
		MinPSUPower: 300, MaxPSUPower: 500,
		CaseFormFactors:  []string{"Mini-ITX", "Micro-ATX"},
		MinSSDThroughput: 0,
		SSDFormFactors:   []string{"2.5", "M.2"},
		MinHDDCapacity:   2000, MaxHDDCapacity: 16000, // первичный сценарий для HDD
	},
}
