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
		MinCPUTDP:          15, MaxCPUTDP: 65,
		RAMType: "DDR4", MinRAM: 8, MaxRAM: 16,
		MinGPUMemory: 0, MaxGPUMemory: 0, // iGPU достаточно
		MinPSUPower: 250, MaxPSUPower: 400,
		CaseFormFactors:  []string{"Micro-ATX"},
		MinSSDThroughput: 500, // обычный SATA-SSD
		SSDFormFactors:   []string{"2.5", "SATA"},
		MinHDDCapacity:   0, MaxHDDCapacity: 0, // HDD за ненадобностью
	},
	"gaming": {
		CPUSocketWhitelist: []string{"AM5", "LGA1700"},
		MinCPUTDP:          65, MaxCPUTDP: 125,
		RAMType: "DDR5", MinRAM: 16, MaxRAM: 32,
		MinGPUMemory: 8, MaxGPUMemory: 16,
		MinPSUPower: 650, MaxPSUPower: 1000,
		CaseFormFactors:  []string{"ATX"},
		MinSSDThroughput: 3500, // NVMe Gen3/4
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   1000, MaxHDDCapacity: 6000, // под библиотеку игр
	},
	"htpc": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700"},
		MinCPUTDP:          15, MaxCPUTDP: 45,
		RAMType: "DDR4", MinRAM: 8, MaxRAM: 16,
		MinGPUMemory: 2, MaxGPUMemory: 4,
		MinPSUPower: 200, MaxPSUPower: 350,
		CaseFormFactors:  []string{"Mini-ITX"},
		MinSSDThroughput: 500,
		SSDFormFactors:   []string{"M.2", "2.5"},
		MinHDDCapacity:   2000, MaxHDDCapacity: 8000, // медиатека
	},
	"streamer": {
		CPUSocketWhitelist: []string{"AM5", "LGA1700"},
		MinCPUTDP:          95, MaxCPUTDP: 125,
		RAMType: "DDR5", MinRAM: 32, MaxRAM: 64,
		MinGPUMemory: 10, MaxGPUMemory: 12,
		MinPSUPower: 750, MaxPSUPower: 1000,
		CaseFormFactors:  []string{"ATX"},
		MinSSDThroughput: 3500,
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   2000, MaxHDDCapacity: 8000, // хранение VOD
	},
	"design": {
		CPUSocketWhitelist: []string{"AM5"},
		MinCPUTDP:          95, MaxCPUTDP: 170,
		RAMType: "DDR5", MinRAM: 32, MaxRAM: 128,
		MinGPUMemory: 8, MaxGPUMemory: 24,
		MinPSUPower: 650, MaxPSUPower: 1200,
		CaseFormFactors:  []string{"ATX"},
		MinSSDThroughput: 3000,
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   0, MaxHDDCapacity: 4000, // по желанию под архив
	},
	"video": {
		CPUSocketWhitelist: []string{"AM5"},
		MinCPUTDP:          95, MaxCPUTDP: 170,
		RAMType: "DDR5", MinRAM: 64, MaxRAM: 128,
		MinGPUMemory: 12, MaxGPUMemory: 24,
		MinPSUPower: 750, MaxPSUPower: 1200,
		CaseFormFactors:  []string{"ATX"},
		MinSSDThroughput: 3500,
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   4000, MaxHDDCapacity: 16000, // много исходников
	},
	"cad": {
		CPUSocketWhitelist: []string{"AM5", "LGA1700"},
		MinCPUTDP:          95, MaxCPUTDP: 170,
		RAMType: "DDR5", MinRAM: 32, MaxRAM: 64,
		MinGPUMemory: 6, MaxGPUMemory: 12, // проф. CAD-карты
		MinPSUPower: 650, MaxPSUPower: 1000,
		CaseFormFactors:  []string{"ATX"},
		MinSSDThroughput: 3000,
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   0, MaxHDDCapacity: 0,
	},
	"dev": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700"},
		MinCPUTDP:          65, MaxCPUTDP: 105,
		RAMType: "DDR4", MinRAM: 16, MaxRAM: 32,
		MinGPUMemory: 4, MaxGPUMemory: 8,
		MinPSUPower: 550, MaxPSUPower: 750,
		CaseFormFactors:  []string{"ATX", "Micro-ATX"},
		MinSSDThroughput: 2000,
		SSDFormFactors:   []string{"M.2", "2.5"},
		MinHDDCapacity:   0, MaxHDDCapacity: 2000,
	},
	"enthusiast": {
		CPUSocketWhitelist: []string{"AM5", "LGA1700"},
		MinCPUTDP:          95, MaxCPUTDP: 170,
		RAMType: "DDR5", MinRAM: 32, MaxRAM: 128,
		MinGPUMemory: 12, MaxGPUMemory: 24,
		MinPSUPower: 850, MaxPSUPower: 1500,
		CaseFormFactors:  []string{"ATX"},
		MinSSDThroughput: 5000, // Gen5 -> запас
		SSDFormFactors:   []string{"M.2"},
		MinHDDCapacity:   0, MaxHDDCapacity: 0,
	},
	"nas": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700"},
		MinCPUTDP:          0, MaxCPUTDP: 45,
		RAMType: "DDR4", MinRAM: 8, MaxRAM: 32,
		MinGPUMemory: 0, MaxGPUMemory: 0,
		MinPSUPower: 250, MaxPSUPower: 500,
		CaseFormFactors:  []string{"Micro-ATX"},
		MinSSDThroughput: 100,
		SSDFormFactors:   []string{"2.5", "M.2"},
		MinHDDCapacity:   4000, MaxHDDCapacity: 16000, // роль хранилища
	},
}
