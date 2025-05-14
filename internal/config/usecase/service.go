package usecase

import (
	"StartupPCConfigurator/internal/config/repository"
	"StartupPCConfigurator/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrConfigNotFound = errors.New("configuration not found")
	ErrForbidden      = errors.New("not owner of configuration")
)

type ConfigService interface {
	FetchComponents(category, search string) ([]domain.Component, error)
	CreateConfiguration(userId uuid.UUID, name string, comps []domain.ComponentRef) (domain.Configuration, error)
	FetchCompatibleComponentsMulti(category string, bases []domain.ComponentRef) ([]domain.Component, error)
	FetchUserConfigurations(userId uuid.UUID) ([]domain.Configuration, error)
	UpdateConfiguration(userId uuid.UUID, configId string, name string, comps []domain.ComponentRef) (domain.Configuration, error)
	DeleteConfiguration(userId uuid.UUID, configId string) error
	GenerateConfigurations(refs []domain.ComponentRef) ([][]domain.Component, error)
	GetUseCaseBuild(usecaseName string) ([]domain.Component, error)
	GenerateUseCaseConfigs(usecaseName string, refs []domain.ComponentRef) ([][]domain.Component, error)
	ListUseCases() ([]domain.UseCase, error)
}

type configService struct {
	repo repository.ConfigRepository
}

func NewConfigService(r repository.ConfigRepository) ConfigService {
	return &configService{repo: r}
}

func (s *configService) FetchComponents(category, search string) ([]domain.Component, error) {
	return s.repo.GetComponents(category, search)
}

func (s *configService) FetchCompatibleComponentsMulti(
	category string,
	bases []domain.ComponentRef,
) ([]domain.Component, error) {
	// 1) Если ищем кулеры — сразу отфильтруем bases так, чтобы в нём были только CPU и Case
	if strings.EqualFold(category, "cooler") {
		var filtered []domain.ComponentRef
		for _, ref := range bases {
			if strings.EqualFold(ref.Category, "cpu") ||
				strings.EqualFold(ref.Category, "case") {
				filtered = append(filtered, ref)
			}
		}
		bases = filtered
	}

	// 2) Собираем все specs в одну карту и считаем нагрузку на PSU
	merged := make(map[string]interface{})
	var totalDraw float64

	for _, ref := range bases {
		base, err := s.repo.GetComponentByName(ref.Category, ref.Name)
		if err != nil {
			return nil, fmt.Errorf("component not found: %s/%s", ref.Category, ref.Name)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(base.Specs, &m); err != nil {
			return nil, fmt.Errorf("invalid specs for %s/%s: %w", ref.Category, ref.Name, err)
		}

		// мёржим всё
		for k, v := range m {
			merged[k] = v
		}

		// если целевая категория PSU — аккумулируем power_draw от CPU и GPU
		if strings.EqualFold(category, "psu") &&
			(strings.EqualFold(ref.Category, "cpu") || strings.EqualFold(ref.Category, "gpu")) {
			if d, ok := m["power_draw"].(float64); ok {
				totalDraw += d
			}
		}
	}

	// 3) Преобразования под каждую категорию
	switch strings.ToLower(category) {
	case "case":
		if h, ok := merged["cooler_height"]; ok {
			merged["cooler_max_height"] = h
		}

	case "gpu":
		if v, ok := merged["pcie_version"]; ok {
			merged["interface"] = v
		}

	case "psu":
		// только требуемая мощность
		merged = map[string]interface{}{
			"power": totalDraw + 150,
		}

	case "ssd":
		// интерфейс: NVMe или SATA
		if v, ok := merged["pcie_version"]; ok {
			merged["interface"] = v
		} else if _, ok := merged["sata_ports"]; ok {
			merged["interface"] = "SATA III"
		}
		// форм-фактор: M.2 или 2.5"
		if slots, ok := merged["m2_slots"].(float64); ok && slots >= 1 {
			merged["form_factor"] = "M.2"
		} else {
			merged["form_factor"] = "2.5"
		}

	case "cooler":
		// собираем только socket + height_mm
		height, _ := merged["cooler_max_height"]
		socket, _ := merged["socket"]
		merged = map[string]interface{}{
			"socket":    socket,
			"height_mm": height,
		}
	}

	// 4) Формируем и отправляем фильтр
	filter := domain.CompatibilityFilter{
		Category: category,
		Specs:    merged,
	}
	return s.repo.GetCompatibleComponents(filter)
}

func (s *configService) CreateConfiguration(userId uuid.UUID, name string, refs []domain.ComponentRef) (domain.Configuration, error) {
	if name == "" {
		return domain.Configuration{}, errors.New("name is required")
	}
	if len(refs) == 0 {
		return domain.Configuration{}, errors.New("at least one component required")
	}

	var fullComps []domain.Component
	for _, ref := range refs {
		comp, err := s.repo.GetComponentByName(ref.Category, ref.Name)
		if err != nil {
			return domain.Configuration{}, fmt.Errorf("component not found: %s / %s", ref.Category, ref.Name)
		}
		fullComps = append(fullComps, comp)
	}

	errs := CheckCompatibility(fullComps)
	if len(errs) > 0 {
		return domain.Configuration{}, fmt.Errorf("сборка несовместима: %s", strings.Join(errs, "; "))
	}

	return s.repo.CreateConfiguration(userId, name, fullComps)
}

func (s *configService) UpdateConfiguration(userId uuid.UUID, configId string, name string, refs []domain.ComponentRef) (domain.Configuration, error) {
	var fullComps []domain.Component
	for _, ref := range refs {
		comp, err := s.repo.GetComponentByName(ref.Category, ref.Name)
		if err != nil {
			return domain.Configuration{}, fmt.Errorf("component not found: %s / %s", ref.Category, ref.Name)
		}
		fullComps = append(fullComps, comp)
	}

	errs := CheckCompatibility(fullComps)
	if len(errs) > 0 {
		return domain.Configuration{}, fmt.Errorf("сборка несовместима: %s", strings.Join(errs, "; "))
	}

	updated, err := s.repo.UpdateConfiguration(userId, configId, name, fullComps)
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			return domain.Configuration{}, domain.ErrConfigNotFound
		} else if errors.Is(err, domain.ErrForbidden) {
			return domain.Configuration{}, domain.ErrForbidden
		}
		return domain.Configuration{}, err
	}
	return updated, nil
}

func (s *configService) FetchUserConfigurations(userId uuid.UUID) ([]domain.Configuration, error) {
	return s.repo.GetUserConfigurations(userId)
}

func (s *configService) DeleteConfiguration(userId uuid.UUID, configId string) error {
	err := s.repo.DeleteConfiguration(userId, configId)
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			return domain.ErrConfigNotFound
		} else if errors.Is(err, domain.ErrForbidden) {
			return domain.ErrForbidden
		}
		return err
	}
	return nil
}

func (s *configService) GenerateConfigurations(refs []domain.ComponentRef) ([][]domain.Component, error) {
	// 1) По refs собираем выбранные компоненты
	var selected []domain.Component
	for _, ref := range refs {
		comp, err := s.repo.GetComponentByName(ref.Category, ref.Name)
		if err != nil {
			return nil, fmt.Errorf("component not found: %s / %s", ref.Category, ref.Name)
		}
		selected = append(selected, comp)
	}

	// 2) Подгружаем все компоненты из репозитория
	all, err := s.repo.GetComponents("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to load components: %w", err)
	}

	// 3) Генерируем все совместимые сборки из полного списка
	configs := GenerateConfigurations(all)

	// 4) Фильтруем, оставляя только сборки, содержащие все выбранные компоненты
	//    при этом сравниваем категории в нижнем регистре
	filtered := make([][]domain.Component, 0, len(configs))
	for _, combo := range configs {
		ok := true
		for _, want := range selected {
			wantCat := strings.ToLower(want.Category) // <-- нормализация want
			found := false
			for _, have := range combo {
				if strings.ToLower(have.Category) == wantCat && have.Name == want.Name {
					found = true
					break
				}
			}
			if !found {
				ok = false
				break
			}
		}
		if ok {
			filtered = append(filtered, combo)
		}
	}

	return filtered, nil
}

// 4) Функция matchesScenario рядом, в том же файле
func (s *configService) GetUseCaseBuild(usecaseName string) ([]domain.Component, error) {
	rule, ok := ScenarioRules[usecaseName]
	if !ok {
		return nil, fmt.Errorf("unknown use case %q", usecaseName)
	}

	all, err := s.repo.GetComponents("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to load components: %w", err)
	}

	combos := GenerateConfigurations(all)
	for _, combo := range combos {
		if matchesScenario(combo, rule) {
			return combo, nil
		}
	}
	return nil, fmt.Errorf("no builds found for use case %q", usecaseName)
}

// matchesScenario проверяет, подходит ли комбинация компонентов под правило ScenarioRule.
func matchesScenario(combo []domain.Component, rule ScenarioRule) bool {
	var cpu, mb, ram, gpu, psu, cs *domain.Component

	// Разбираем combo по категориям
	for i := range combo {
		switch strings.ToLower(combo[i].Category) {
		case "cpu":
			cpu = &combo[i]
		case "motherboard":
			mb = &combo[i]
		case "ram":
			ram = &combo[i]
		case "gpu":
			gpu = &combo[i]
		case "psu":
			psu = &combo[i]
		case "case":
			cs = &combo[i]
		}
	}
	// Все ключевые компоненты должны присутствовать
	if cpu == nil || mb == nil || ram == nil || psu == nil || cs == nil {
		return false
	}

	// Распаковываем JSON specs в map[string]interface{}
	specsCPU := parseSpecs(cpu.Specs)
	specsMB := parseSpecs(mb.Specs)
	specsRAM := parseSpecs(ram.Specs)
	specsPSU := parseSpecs(psu.Specs)
	specsCase := parseSpecs(cs.Specs)

	// 1) CPU: socket в белом списке
	sock, _ := specsCPU["socket"].(string)
	if !contains(rule.CPUSocketWhitelist, sock) {
		return false
	}
	// 2) CPU: TDP (мин/макс)
	if tdpRaw, ok := specsCPU["tdp"].(float64); ok {
		tdp := int(tdpRaw)
		if rule.MinCPUTDP > 0 && tdp < rule.MinCPUTDP {
			return false
		}
		if rule.MaxCPUTDP > 0 && tdp > rule.MaxCPUTDP {
			return false
		}
	}

	// 3) Motherboard ↔ RAM: тип памяти
	if rule.RAMType != "" {
		if mbType, _ := specsMB["ram_type"].(string); mbType != rule.RAMType {
			return false
		}
		if ramType, _ := specsRAM["ram_type"].(string); ramType != rule.RAMType {
			return false
		}
	}
	// 4) RAM: объём (мин/макс)
	if capRaw, ok := specsRAM["capacity"].(float64); ok {
		cap := int(capRaw)
		if rule.MinRAM > 0 && cap < rule.MinRAM {
			return false
		}
		if rule.MaxRAM > 0 && cap > rule.MaxRAM {
			return false
		}
	}

	// 5) GPU: память (мин/макс)
	if rule.MinGPUMemory > 0 || rule.MaxGPUMemory > 0 {
		if gpu == nil {
			return false
		}
		specsGPU := parseSpecs(gpu.Specs)
		if memRaw, ok := specsGPU["memory_gb"].(float64); ok {
			mem := int(memRaw)
			if rule.MinGPUMemory > 0 && mem < rule.MinGPUMemory {
				return false
			}
			if rule.MaxGPUMemory > 0 && mem > rule.MaxGPUMemory {
				return false
			}
		} else {
			return false
		}
	}

	// 6) PSU: мощность (мин/макс)
	if powRaw, ok := specsPSU["power"].(float64); ok {
		pow := int(powRaw)
		if rule.MinPSUPower > 0 && pow < rule.MinPSUPower {
			return false
		}
		if rule.MaxPSUPower > 0 && pow > rule.MaxPSUPower {
			return false
		}
	}

	// 7) Case ↔ Motherboard: проверяем, что форм-фактор материнской платы поддерживается корпусом
	if arr, ok := specsCase["max_motherboard_form_factors"].([]interface{}); ok {
		mfMB, _ := specsMB["form_factor"].(string)
		supported := false
		for _, v := range arr {
			if s, ok := v.(string); ok && strings.EqualFold(s, mfMB) {
				supported = true
				break
			}
		}
		if !supported {
			return false
		}
	}

	return true
}

// Вспомогательная: распаковка JSONB в map[string]interface{}
func parseSpecs(raw json.RawMessage) map[string]interface{} {
	var m map[string]interface{}
	json.Unmarshal(raw, &m)
	return m
}

// Вспомогательная: поиск строки в срезе
func contains(ss []string, s string) bool {
	for _, x := range ss {
		if strings.EqualFold(x, s) {
			return true
		}
	}
	return false
}

func (s *configService) GenerateUseCaseConfigs(usecaseName string, refs []domain.ComponentRef) ([][]domain.Component, error) {
	// 1) Получаем правило
	rule, ok := ScenarioRules[usecaseName]
	if !ok {
		return nil, fmt.Errorf("unknown use case %q", usecaseName)
	}
	// 2) Собираем все компоненты по refs (как в GenerateConfigurations)
	var selected []domain.Component
	for _, ref := range refs {
		comp, err := s.repo.GetComponentByName(ref.Category, ref.Name)
		if err != nil {
			return nil, fmt.Errorf("component not found: %s/%s", ref.Category, ref.Name)
		}
		selected = append(selected, comp)
	}
	// 3) Загружаем всё и получаем все совместимые комбо
	all, err := s.repo.GetComponents("", "")
	if err != nil {
		return nil, err
	}
	combos := GenerateConfigurations(all)
	// 4) Фильтр: содержит все selected **и** соответствует сценарию
	var out [][]domain.Component
	for _, combo := range combos {
		if containsAll(combo, selected) && matchesScenario(combo, rule) {
			out = append(out, combo)
		}
	}
	return out, nil
}

// вспомогательный: combo содержит все want
func containsAll(combo, want []domain.Component) bool {
	for _, w := range want {
		ok := false
		for _, c := range combo {
			if strings.EqualFold(c.Category, w.Category) && c.Name == w.Name {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

// ListUseCases возвращает все сценарии из БД
func (s *configService) ListUseCases() ([]domain.UseCase, error) {
	return s.repo.GetUseCases()
}
