package domain

import "strings"

// ComponentCategory представляет тип компонента
type ComponentCategory string

// Константы категорий
const (
	CategoryCPU           ComponentCategory = "cpu"
	CategoryMotherboard   ComponentCategory = "motherboard"
	CategoryRAM           ComponentCategory = "ram"
	CategoryGPU           ComponentCategory = "gpu"
	CategoryPSU           ComponentCategory = "psu"
	CategoryCase          ComponentCategory = "case"
	CategoryCooler        ComponentCategory = "cpu_cooler"
	CategorySSD           ComponentCategory = "ssd"
	CategoryHDD           ComponentCategory = "hdd"
	CategoryCaseFan       ComponentCategory = "case_fan"
	CategoryWiFiAdapter   ComponentCategory = "wifi_adapter"
	CategoryOpticalDrive  ComponentCategory = "optical_drive"
	CategorySoundCard     ComponentCategory = "sound_card"
	CategoryCaptureCard   ComponentCategory = "capture_card"
	CategoryRGBController ComponentCategory = "rgb_controller"
)

// AllCategories — полный список допустимых компонент
var AllCategories = []ComponentCategory{
	CategoryCPU,
	CategoryMotherboard,
	CategoryRAM,
	CategoryGPU,
	CategoryPSU,
	CategoryCase,
	CategoryCooler,
	CategorySSD,
	CategoryHDD,
	CategoryCaseFan,
	CategoryWiFiAdapter,
	CategoryOpticalDrive,
	CategorySoundCard,
	CategoryCaptureCard,
	CategoryRGBController,
}

// CategorySet — для быстрой проверки через map
var CategorySet = func() map[string]struct{} {
	m := make(map[string]struct{})
	for _, cat := range AllCategories {
		m[strings.ToLower(string(cat))] = struct{}{}
	}
	return m
}()

// IsValidCategory проверяет, допустима ли категория
func IsValidCategory(input string) bool {
	_, ok := CategorySet[strings.ToLower(input)]
	return ok
}

// ExpectedSpecs описывает обязательные поля в specs для каждой категории
var ExpectedSpecs = map[ComponentCategory][]string{
	CategoryCPU:           {"socket", "tdp", "supported_ram_type", "max_ram_freq"},
	CategoryMotherboard:   {"socket", "ram_type", "form_factor"},
	CategoryRAM:           {"ram_type", "frequency", "capacity"},
	CategoryGPU:           {"length_mm", "power_draw"},
	CategoryPSU:           {"power", "pcie_connectors"},
	CategoryCase:          {"form_factor_support", "gpu_max_length", "cooler_max_height"},
	CategoryCooler:        {"supported_sockets", "height_mm"},
	CategorySSD:           {"interface", "form_factor", "nvme"},
	CategoryHDD:           {"interface", "rpm", "capacity"},
	CategoryCaseFan:       {"diameter", "rpm", "airflow", "connector_type"},
	CategoryWiFiAdapter:   {"interface", "wifi_standard"},
	CategoryOpticalDrive:  {"interface", "form_factor"},
	CategorySoundCard:     {"interface", "channels"},
	CategoryCaptureCard:   {"interface", "input_type"},
	CategoryRGBController: {"interface", "channels", "software_support"},
}

// ValidateSpecs проверяет, что в specs есть все ожидаемые поля для заданной категории
func ValidateSpecs(category ComponentCategory, specs map[string]interface{}) []string {
	missing := []string{}
	expected, ok := ExpectedSpecs[category]
	if !ok {
		return missing // нет требований
	}
	for _, key := range expected {
		if _, ok := specs[key]; !ok {
			missing = append(missing, key)
		}
	}
	return missing
}
