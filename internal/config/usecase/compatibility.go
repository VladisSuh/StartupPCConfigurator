package usecase

import (
	"StartupPCConfigurator/internal/domain"
	"encoding/json"
)

func CheckCompatibility(components []domain.Component) []string {
	errors := []string{}

	var parsed []struct {
		Category string
		Specs    map[string]interface{}
	}

	for _, c := range components {
		var specs map[string]interface{}
		if err := json.Unmarshal(c.Specs, &specs); err != nil {
			continue
		}
		parsed = append(parsed, struct {
			Category string
			Specs    map[string]interface{}
		}{
			Category: c.Category,
			Specs:    specs,
		})
	}

	var cpu, mb, ram, gpu, psu, caseComp *map[string]interface{}

	for _, c := range parsed {
		switch c.Category {
		case "cpu":
			cpu = &c.Specs
		case "motherboard":
			mb = &c.Specs
		case "ram":
			ram = &c.Specs
		case "gpu":
			gpu = &c.Specs
		case "psu":
			psu = &c.Specs
		case "case":
			caseComp = &c.Specs
		}
	}

	if cpu != nil && mb != nil {
		if (*cpu)["socket"] != (*mb)["socket"] {
			errors = append(errors, "CPU и Motherboard несовместимы (socket)")
		}
	}

	if ram != nil && mb != nil {
		if (*ram)["ram_type"] != (*mb)["ram_type"] {
			errors = append(errors, "RAM и Motherboard несовместимы (тип памяти)")
		}
	}

	if gpu != nil && caseComp != nil {
		gpuLen, ok1 := (*gpu)["length_mm"].(float64)
		caseMaxLen, ok2 := (*caseComp)["gpu_max_length"].(float64)
		if ok1 && ok2 && gpuLen > caseMaxLen {
			errors = append(errors, "GPU слишком длинная для корпуса")
		}
	}

	if cpu != nil && caseComp != nil {
		coolerHeight, ok1 := (*cpu)["cooler_height"].(float64)
		caseMaxHeight, ok2 := (*caseComp)["cooler_max_height"].(float64)
		if ok1 && ok2 && coolerHeight > caseMaxHeight {
			errors = append(errors, "Кулер не помещается в корпус по высоте")
		}
	}

	if psu != nil && gpu != nil {
		gpuPower, ok1 := (*gpu)["power_draw"].(float64)
		psuPower, ok2 := (*psu)["power"].(float64)
		if ok1 && ok2 && psuPower < gpuPower+150 {
			errors = append(errors, "PSU может быть слишком слабым для GPU")
		}
	}

	return errors
}
