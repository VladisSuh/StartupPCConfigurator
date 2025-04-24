package usecase

type ScenarioRule struct {
	CPUSocketWhitelist []string // допустимые сокеты CPU
	MaxCPUTDP          int      // максимальный TDP ЦПУ (Вт)
	RAMType            string
	MinRAM             int
	MinGPUMemory       int
	MinPSUPower        int
	CaseFormFactors    []string
}

var ScenarioRules = map[string]ScenarioRule{
	"office": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700"},
		MaxCPUTDP:          65, // офисным хватит 65 Вт и ниже
		RAMType:            "DDR4",
		MinRAM:             8,
		MinGPUMemory:       0,
		MinPSUPower:        300,
		CaseFormFactors:    []string{"ATX", "mATX", "Mini-ITX"},
	},
	"htpc": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700"},
		MaxCPUTDP:          35, // очень низкое тепловыделение
		RAMType:            "DDR4",
		MinRAM:             8,
		MinGPUMemory:       0,
		MinPSUPower:        300,
		CaseFormFactors:    []string{"Mini-ITX"},
	},
	"gaming": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700", "AM5"},
		MaxCPUTDP:          125, // допускаем «тяжёлые» ЦПУ
		RAMType:            "DDR4",
		MinRAM:             16,
		MinGPUMemory:       6,
		MinPSUPower:        650,
		CaseFormFactors:    []string{"ATX", "mATX"},
	},
	"streamer": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700", "AM5"},
		MaxCPUTDP:          125,
		RAMType:            "DDR4",
		MinRAM:             32,
		MinGPUMemory:       6,
		MinPSUPower:        750,
		CaseFormFactors:    []string{"ATX", "mATX"},
	},
	"design": {
		CPUSocketWhitelist: []string{"AM5"},
		MaxCPUTDP:          170,
		RAMType:            "DDR5",
		MinRAM:             32,
		MinGPUMemory:       10,
		MinPSUPower:        650,
		CaseFormFactors:    []string{"ATX"},
	},
	"video": {
		CPUSocketWhitelist: []string{"AM5"},
		MaxCPUTDP:          170,
		RAMType:            "DDR5",
		MinRAM:             64,
		MinGPUMemory:       10,
		MinPSUPower:        750,
		CaseFormFactors:    []string{"ATX"},
	},
	"cad": {
		CPUSocketWhitelist: []string{"AM5", "LGA1700"},
		MaxCPUTDP:          170,
		RAMType:            "DDR5",
		MinRAM:             64,
		MinGPUMemory:       4,
		MinPSUPower:        650,
		CaseFormFactors:    []string{"ATX"},
	},
	"dev": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700"},
		MaxCPUTDP:          95, // средний CPU до 95 Вт
		RAMType:            "DDR4",
		MinRAM:             16,
		MinGPUMemory:       0,
		MinPSUPower:        550,
		CaseFormFactors:    []string{"ATX", "mATX", "Mini-ITX"},
	},
	"enthusiast": {
		CPUSocketWhitelist: []string{"AM5", "LGA1700"},
		MaxCPUTDP:          170,
		RAMType:            "DDR5",
		MinRAM:             32,
		MinGPUMemory:       8,
		MinPSUPower:        850,
		CaseFormFactors:    []string{"ATX"},
	},
	"nas": {
		CPUSocketWhitelist: []string{"AM4", "LGA1700"},
		MaxCPUTDP:          65,
		RAMType:            "DDR4",
		MinRAM:             8,
		MinGPUMemory:       0,
		MinPSUPower:        300,
		CaseFormFactors:    []string{"Mini-ITX"},
	},
}
