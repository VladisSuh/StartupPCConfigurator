package usecase

import (
	"encoding/json"
	"fmt"
	"strings"

	"StartupPCConfigurator/internal/domain"
)

// CheckCompatibility проверяет полную сборку и возвращает список ошибок, если что-то не совместимо.
// Поддерживаем проверки:
//   - CPU ↔ MB (socket)
//   - RAM ↔ MB (ram_type)
//   - SSD ↔ MB (interface + form_factor)
//   - GPU ↔ Case (length_mm)
//   - CPU ↔ Case (cooler_height)
//   - PSU ↔ (CPU power_draw + GPU power_draw + 150)
//   - Case ↔ MB (form_factor входит в список допустимых у корпуса)
func CheckCompatibility(components []domain.Component) []string {
	errs := []string{}
	// Распарсим все specs в map по категории
	specsByCat := map[string]map[string]interface{}{}
	for _, c := range components {
		var m map[string]interface{}
		if err := json.Unmarshal(c.Specs, &m); err != nil {
			continue
		}
		specsByCat[strings.ToLower(c.Category)] = m
	}

	cpu := specsByCat["cpu"]
	mb := specsByCat["motherboard"]
	ram := specsByCat["ram"]
	gpu := specsByCat["gpu"]
	psu := specsByCat["psu"]
	cs := specsByCat["case"]
	ssd := specsByCat["ssd"]
	hdd := specsByCat["hdd"]

	// 1) CPU ↔ MB: socket
	if cpu != nil && mb != nil {
		if cpu["socket"] != mb["socket"] {
			errs = append(errs, fmt.Sprintf("CPU.socket %v ≠ MB.socket %v", cpu["socket"], mb["socket"]))
		}
	}
	// 2) RAM ↔ MB: ram_type
	if ram != nil && mb != nil {
		if ram["ram_type"] != mb["ram_type"] {
			errs = append(errs, fmt.Sprintf("RAM.ram_type %v ≠ MB.ram_type %v", ram["ram_type"], mb["ram_type"]))
		}
	}
	// 3) SSD ↔ MB: interface + form_factor
	if ssd != nil && mb != nil {
		// интерфейс
		if ssd["interface"] != mb["pcie_version"] && ssd["interface"] != mb["interface"] {
			errs = append(errs, fmt.Sprintf("SSD.interface %v ≠ MB.pcie_version %v", ssd["interface"], mb["pcie_version"]))
		}
		// форм-фактор (M.2 vs число слотов)
		if ssdFF, ok := ssd["form_factor"].(string); ok {
			if ssdFF == "M.2" {
				if m2, ok := mb["m2_slots"].(float64); !ok || m2 < 1 {
					errs = append(errs, "SSD требует M.2-слот, а MB не поддерживает")
				}
			}
		}
	}

	// 3.1) HDD ↔ MB: интерфейс + наличие портов
	if hdd != nil && mb != nil {
		// a) HDD должен быть SATA-семейства
		if iface, _ := hdd["interface"].(string); !strings.HasPrefix(strings.ToUpper(iface), "SATA") {
			errs = append(errs, fmt.Sprintf("HDD.interface %v ≠ SATA", iface))
		}

		// b) На плате обязательно ≥1 SATA-порт
		if ports, ok := mb["sata_ports"].(float64); !ok || ports < 1 {
			errs = append(errs, "HDD требует SATA-порт, а MB не поддерживает")
		}
	}

	// 3.2) HDD ↔ Case: есть ли свободный 3.5-бей
	if hdd != nil && cs != nil {
		if bays, ok := cs["drive_bays_3_5"].(float64); ok && bays < 1 {
			errs = append(errs, "Case не имеет 3.5\" отсеков для HDD")
		}
	}

	// 4) GPU ↔ Case: длина
	if gpu != nil && cs != nil {
		if gl, ok1 := gpu["length_mm"].(float64); ok1 {
			if cm, ok2 := cs["gpu_max_length"].(float64); ok2 && gl > cm {
				errs = append(errs, fmt.Sprintf("GPU.length_mm %.0f > Case.gpu_max_length %.0f", gl, cm))
			}
		}
	}
	// 5) CPU ↔ Case: высота кулера
	if cpu != nil && cs != nil {
		if ch, ok1 := cpu["cooler_height"].(float64); ok1 {
			if cmh, ok2 := cs["cooler_max_height"].(float64); ok2 && ch > cmh {
				errs = append(errs, fmt.Sprintf("CPU.cooler_height %.0f > Case.cooler_max_height %.0f", ch, cmh))
			}
		}
	}
	// 6) Case ↔ MB: form_factor входит в список
	if cs != nil && mb != nil {
		if allowed, ok := cs["max_motherboard_form_factors"].([]interface{}); ok {
			ff := mb["form_factor"]
			okf := false
			for _, x := range allowed {
				if x == ff {
					okf = true
					break
				}
			}
			if !okf {
				errs = append(errs, fmt.Sprintf("MB.form_factor %v не поддерживается Case", ff))
			}
		}
	}
	// 7) PSU ↔ CPU+GPU: мощность
	if psu != nil {
		need := 150.0 // запас
		if cpu != nil {
			if d, ok := cpu["power_draw"].(float64); ok {
				need += d
			}
		}
		if gpu != nil {
			if d, ok := gpu["power_draw"].(float64); ok {
				need += d
			}
		}
		if p, ok := psu["power"].(float64); ok && p < need {
			errs = append(errs, fmt.Sprintf("PSU.power %.0f < required %.0f", p, need))
		}
	}
	// PSU ↔ Case form-factor
	if psu != nil && cs != nil {
		if ff, _ := psu["form_factor"].(string); ff != "" {
			if cf, _ := cs["psu_form_factor"].(string); !strings.EqualFold(ff, cf) {
				errs = append(errs,
					fmt.Sprintf("PSU.form_factor %v ≠ Case.psu_form_factor %v", ff, cf))
			}
		}
	}

	return errs
}

// GenerateConfigurations возвращает все возможные сборки из всех категорий
// с учётом CheckCompatibility. Делает бэктрекинг с ранней отбраковкой.
func GenerateConfigurations(allComponents []domain.Component) [][]domain.Component {
	// сгруппируем по категориям
	categories := map[string][]domain.Component{}
	for _, c := range allComponents {
		key := strings.ToLower(c.Category)
		categories[key] = append(categories[key], c)
	}
	// список категорий, которые участвуют в сборке
	order := []string{"cpu", "motherboard", "ram", "ssd", "hdd", "gpu", "psu", "case", "cooler"}

	var results [][]domain.Component
	var current []domain.Component

	var backtrack func(idx int)
	backtrack = func(idx int) {
		if idx == len(order) {
			// коньченый набор — проверяем всю сборку
			if len(CheckCompatibility(current)) == 0 {
				combo := make([]domain.Component, len(current))
				copy(combo, current)
				results = append(results, combo)
			}
			return
		}
		cat := order[idx]
		for _, comp := range categories[cat] {
			current = append(current, comp)
			// частичная проверка: если хоть одна ошибка — отсекаем
			if len(CheckCompatibility(current)) == 0 {
				backtrack(idx + 1)
			}
			current = current[:len(current)-1]
		}
	}

	backtrack(0)
	return results
}
