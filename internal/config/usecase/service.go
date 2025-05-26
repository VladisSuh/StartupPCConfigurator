package usecase

import (
	"StartupPCConfigurator/internal/config/repository"
	"StartupPCConfigurator/internal/config/usecase/rules"
	"StartupPCConfigurator/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrConfigNotFound = errors.New("configuration not found")
	ErrForbidden      = errors.New("not owner of configuration")
)

var buildLabels = []string{
	"Бюджетная сборка",
	"Приемлемая сборка",
	"Сбалансированная сборка",
	"Улучшенная сборка",
	"Максимальная сборка",
}

type ConfigService interface {
	FetchComponents(category, search, brand, usecase string) ([]domain.Component, error)
	CreateConfiguration(userId uuid.UUID, name string, comps []domain.ComponentRef) (domain.Configuration, error)
	FetchCompatibleComponentsMulti(category string, bases []domain.ComponentRef, brand *string, usecase *string) ([]domain.Component, error)
	FetchUserConfigurations(userId uuid.UUID) ([]domain.Configuration, error)
	UpdateConfiguration(userId uuid.UUID, configId string, name string, comps []domain.ComponentRef) (domain.Configuration, error)
	DeleteConfiguration(userId uuid.UUID, configId string) error
	GenerateConfigurations(refs []domain.ComponentRef) ([][]domain.Component, error)
	GetUseCaseBuild(usecaseName string, limit int) ([]domain.NamedBuild, error)
	GenerateUseCaseConfigs(usecaseName string, refs []domain.ComponentRef) ([][]domain.Component, error)
	ListUseCases() ([]domain.UseCase, error)
	ListBrands(category string) ([]string, error)
}

type configService struct {
	repo repository.ConfigRepository
}

func NewConfigService(r repository.ConfigRepository) ConfigService {
	return &configService{repo: r}
}

// service.go
func (s *configService) FetchComponents(
	category, search, brand, usecase string,
) ([]domain.Component, error) {
	// 1. Получили «сырые» данные по категории/search/brand
	comps, err := s.repo.GetComponents(category, search, brand)
	if err != nil {
		return nil, err
	}

	// 2. Если сценарий не указан — отдаем всё
	if usecase == "" {
		return comps, nil
	}
	rule, ok := rules.ScenarioRules[usecase]
	if !ok {
		return nil, fmt.Errorf("unknown usecase %q", usecase)
	}

	// 3. Фильтруем ровно по той части правила, что относится к category
	var out []domain.Component
	for _, c := range comps {
		switch strings.ToLower(c.Category) {
		case "cpu":
			if cpuMatches(c, rule) {
				out = append(out, c)
			}
		case "motherboard":
			if mbMatches(c, rule) {
				out = append(out, c)
			}
		case "ram":
			if ramMatches(c, rule) {
				out = append(out, c)
			}
		case "gpu":
			if gpuMatches(c, rule) {
				out = append(out, c)
			}
		case "psu":
			if psuMatches(c, rule) {
				out = append(out, c)
			}
		case "case":
			if caseMatches(c, rule) {
				out = append(out, c)
			}
		case "ssd":
			if ssdMatches(c, rule) {
				out = append(out, c)
			}

		default:
			// остальные категории — без фильтрации по сценарию
			out = append(out, c)
		}
	}
	return out, nil
}

func (s *configService) FetchCompatibleComponentsMulti(
	category string,
	bases []domain.ComponentRef,
	brand *string,
	usecase *string,
) ([]domain.Component, error) {
	// 1) Получаем «пул» кандидатов по category + опционально brand
	var candidates []domain.Component
	var err error
	if brand != nil {
		candidates, err = s.repo.GetComponentsByFilters(category, brand)
	} else {
		candidates, err = s.repo.GetComponentsByCategory(category)
	}
	if err != nil {
		return nil, err
	}

	// 2) Применяем фильтрацию по сценарию, если он указан
	if usecase != nil && *usecase != "" {
		rule, ok := rules.ScenarioRules[*usecase]
		if !ok {
			return nil, fmt.Errorf("unknown usecase %q", *usecase)
		}
		var filtered []domain.Component
		for _, comp := range candidates {
			switch strings.ToLower(comp.Category) {
			case "cpu":
				if cpuMatches(comp, rule) {
					filtered = append(filtered, comp)
				}
			case "motherboard":
				if mbMatches(comp, rule) {
					filtered = append(filtered, comp)
				}
			case "ram":
				if ramMatches(comp, rule) {
					filtered = append(filtered, comp)
				}
			case "gpu":
				if gpuMatches(comp, rule) {
					filtered = append(filtered, comp)
				}
			case "psu":
				if psuMatches(comp, rule) {
					filtered = append(filtered, comp)
				}
			case "case":
				if caseMatches(comp, rule) {
					filtered = append(filtered, comp)
				}
			case "ssd":
				if ssdMatches(comp, rule) {
					filtered = append(filtered, comp)
				}

			default:
				filtered = append(filtered, comp)
			}
		}
		candidates = filtered
	}

	// 2) Специальная логика для кулеров
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

	// 3) Собираем merged-спецификацию и totalDraw для PSU
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
		for k, v := range m {
			merged[k] = v
		}
		if strings.EqualFold(category, "psu") &&
			(strings.EqualFold(ref.Category, "cpu") || strings.EqualFold(ref.Category, "gpu")) {
			if d, ok := m["power_draw"].(float64); ok {
				totalDraw += d
			}
		}
	}

	// 4) Адаптируем merged-спецификацию под целевую категорию
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
		merged = map[string]interface{}{"power": totalDraw + 150}
	case "ssd":
		if v, ok := merged["pcie_version"]; ok {
			merged["interface"] = v
		} else if _, ok := merged["sata_ports"]; ok {
			merged["interface"] = "SATA III"
		}
		if slots, ok := merged["m2_slots"].(float64); ok && slots >= 1 {
			merged["form_factor"] = "M.2"
		} else {
			merged["form_factor"] = "2.5"
		}
	case "cooler":
		height, _ := merged["cooler_max_height"]
		socket, _ := merged["socket"]
		merged = map[string]interface{}{
			"socket":    socket,
			"height_mm": height,
		}
	}

	// 5) Фильтруем наш пул по совместимости с учётом merged specs
	filter := domain.CompatibilityFilter{
		Category: category,
		Specs:    merged,
	}
	return s.repo.FilterPoolByCompatibility(candidates, filter)
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
	var selected []domain.Component
	for _, ref := range refs {
		comp, err := s.repo.GetComponentByName(ref.Category, ref.Name)
		if err != nil {
			return nil, fmt.Errorf("component not found: %s / %s", ref.Category, ref.Name)
		}
		selected = append(selected, comp)
	}

	all, err := s.repo.GetComponents("", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to load components: %w", err)
	}

	grouped := make(map[string][]domain.Component)
	for _, comp := range all {
		grouped[strings.ToLower(comp.Category)] = append(grouped[strings.ToLower(comp.Category)], comp)
	}

	candidates := generateLimitedCombos(grouped, 100)
	seen := make(map[string]struct{})
	filtered := make([][]domain.Component, 0, 10)

	for _, combo := range candidates {
		ok := true
		for _, want := range selected {
			found := false
			for _, have := range combo {
				if strings.EqualFold(have.Category, want.Category) && have.Name == want.Name {
					found = true
					break
				}
			}
			if !found {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}

		hash := hashCombo(combo)
		if _, exists := seen[hash]; exists {
			continue
		}
		seen[hash] = struct{}{}
		filtered = append(filtered, combo)

		if len(filtered) >= 10 {
			break
		}
	}

	return filtered, nil
}

func generateLimitedCombos(grouped map[string][]domain.Component, max int) [][]domain.Component {
	required := []string{"cpu", "motherboard", "ram", "gpu", "psu", "case", "ssd"}
	result := make([][]domain.Component, 0, max)

	var dfs func(depth int, combo []domain.Component)
	dfs = func(depth int, combo []domain.Component) {
		if len(result) >= max {
			return
		}
		if depth == len(required) {
			if CheckCompatibility(combo) == nil {
				cp := make([]domain.Component, len(combo))
				copy(cp, combo)
				result = append(result, cp)
			}
			return
		}
		cat := required[depth]
		for _, comp := range grouped[cat] {
			dfs(depth+1, append(combo, comp))
		}
	}

	dfs(0, nil)
	return result
}

func hashCombo(combo []domain.Component) string {
	sort.Slice(combo, func(i, j int) bool {
		return combo[i].Category < combo[j].Category
	})
	var sb strings.Builder
	for _, c := range combo {
		sb.WriteString(strings.ToLower(c.Category))
		sb.WriteString(":" + c.Name + ";")
	}
	return sb.String()
}

// 4) Функция matchesScenario рядом, в том же файле
func (s *configService) GetUseCaseBuild(usecaseName string, limit int) ([]domain.NamedBuild, error) {
	rule, ok := rules.ScenarioRules[usecaseName]
	if !ok {
		return nil, fmt.Errorf("unknown use case %q", usecaseName)
	}

	all, err := s.repo.GetComponentsByUseCase(usecaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to load components: %w", err)
	}

	combos := GenerateConfigurations(all)
	seen := make(map[string]struct{})
	ranked := make(map[int][]domain.Component) // rank 0..4 → first combo

	for _, combo := range combos {
		if !matchesScenario(combo, rule) {
			continue
		}
		hash := hashCombo(combo)
		if _, ok := seen[hash]; ok {
			continue
		}
		seen[hash] = struct{}{}

		rank := rankBuildByScenario(rule, combo)
		if rank >= len(buildLabels) {
			rank = len(buildLabels) - 1
		}
		// добавим только если ещё нет такой сборки
		if _, exists := ranked[rank]; !exists {
			ranked[rank] = combo
		}

		if len(ranked) == len(buildLabels) {
			break // все 5 типов собраны
		}
	}

	if len(ranked) == 0 {
		return nil, fmt.Errorf("no builds found for use case %q", usecaseName)
	}

	var builds []domain.NamedBuild
	for rank := 0; rank < len(buildLabels); rank++ {
		if combo, ok := ranked[rank]; ok {
			builds = append(builds, domain.NamedBuild{
				Name:       buildLabels[rank],
				Components: combo,
			})
		}
	}

	return builds, nil
}

var caseSupportedMap = map[string][]string{
	"ATX":       {"ATX", "Micro-ATX", "Mini-ITX"},
	"Micro-ATX": {"Micro-ATX", "Mini-ITX"},
	"Mini-ITX":  {"Mini-ITX"},
}

// matchesScenario проверяет, подходит ли комбинация компонентов под правило rules.ScenarioRule.
func matchesScenario(combo []domain.Component, rule rules.ScenarioRule) bool {
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
	// Все обязательные компоненты (кроме GPU) должны присутствовать
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

	// 5) GPU: требуем карту только если есть жесткое минимальное требование
	if rule.MinGPUMemory > 0 {
		if gpu == nil {
			return false
		}
		specsGPU := parseSpecs(gpu.Specs)
		memRaw, ok := specsGPU["memory_gb"].(float64)
		if !ok {
			return false
		}
		mem := int(memRaw)
		if mem < rule.MinGPUMemory {
			return false
		}
		if rule.MaxGPUMemory > 0 && mem > rule.MaxGPUMemory {
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

	// 7) Case ↔ MB: поддержка форм-факторов без изменения JSON
	caseForm, _ := specsCase["form_factor"].(string) // например "ATX"
	mbForm, _ := specsMB["form_factor"].(string)     // например "Micro-ATX"
	allowed, ok := caseSupportedMap[strings.ToUpper(caseForm)]
	if !ok {
		// если вдруг корпус нестандартный — отбраковываем
		return false
	}
	if !contains(allowed, mbForm) {
		return false
	}

	// 8) Case: корпус сам должен быть в разрешённом списке форм-факторов
	if len(rule.CaseFormFactors) > 0 {
		if caseForm, _ := specsCase["form_factor"].(string); !contains(rule.CaseFormFactors, caseForm) {
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
	rule, ok := rules.ScenarioRules[usecaseName]
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
	filtered, err := s.repo.GetComponentsByUseCase(usecaseName)
	if err != nil {
		return nil, err
	}
	combos := GenerateConfigurations(filtered)
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

func (s *configService) ListBrands(category string) ([]string, error) {
	if category == "" {
		return nil, errors.New("category required")
	}
	return s.repo.GetBrandsByCategory(category)
}

func cpuMatches(c domain.Component, rule rules.ScenarioRule) bool {
	specs := parseSpecs(c.Specs)
	// 1) socket
	sock, _ := specs["socket"].(string)
	if !contains(rule.CPUSocketWhitelist, sock) {
		return false
	}
	// 2) TDP
	if tdpRaw, ok := specs["tdp"].(float64); ok {
		tdp := int(tdpRaw)
		if rule.MinCPUTDP > 0 && tdp < rule.MinCPUTDP {
			return false
		}
		if rule.MaxCPUTDP > 0 && tdp > rule.MaxCPUTDP {
			return false
		}
	}
	return true
}

func mbMatches(c domain.Component, rule rules.ScenarioRule) bool {
	specs := parseSpecs(c.Specs)
	// RAM-type согласовать
	if rule.RAMType != "" {
		if mt, _ := specs["ram_type"].(string); mt != rule.RAMType {
			return false
		}
	}
	// форм-фактор MB — должен поддерживаться корпусом, но тут мы не знаем корпус,
	// поэтому только проверим, что MB форм-фактор входит в allowed-list корпуса из правила
	// (опционально — если хотите)
	return true
}

func ramMatches(c domain.Component, rule rules.ScenarioRule) bool {
	specs := parseSpecs(c.Specs)
	if rt, _ := specs["ram_type"].(string); rt != rule.RAMType {
		return false
	}
	if capRaw, ok := specs["capacity"].(float64); ok {
		cap := int(capRaw)
		if rule.MinRAM > 0 && cap < rule.MinRAM {
			return false
		}
		if rule.MaxRAM > 0 && cap > rule.MaxRAM {
			return false
		}
	}
	return true
}

func gpuMatches(c domain.Component, rule rules.ScenarioRule) bool {
	// GPU необязателен, но если MinGPUMemory>0 — проверяем
	if rule.MinGPUMemory == 0 {
		return true
	}
	specs := parseSpecs(c.Specs)
	if memRaw, ok := specs["memory_gb"].(float64); ok {
		mem := int(memRaw)
		if mem < rule.MinGPUMemory {
			return false
		}
		if rule.MaxGPUMemory > 0 && mem > rule.MaxGPUMemory {
			return false
		}
		return true
	}
	return false
}

func psuMatches(c domain.Component, rule rules.ScenarioRule) bool {
	specs := parseSpecs(c.Specs)
	if powRaw, ok := specs["power"].(float64); ok {
		p := int(powRaw)
		if rule.MinPSUPower > 0 && p < rule.MinPSUPower {
			return false
		}
		if rule.MaxPSUPower > 0 && p > rule.MaxPSUPower {
			return false
		}
	}
	return true
}

func caseMatches(c domain.Component, rule rules.ScenarioRule) bool {
	specs := parseSpecs(c.Specs)
	cf, _ := specs["form_factor"].(string)
	// 1) корпус сам должен быть разрешён в сценарии
	if len(rule.CaseFormFactors) > 0 && !contains(rule.CaseFormFactors, cf) {
		return false
	}
	return true
}

func ssdMatches(c domain.Component, rule rules.ScenarioRule) bool {
	var specs struct {
		MaxThroughput float64 `json:"max_throughput"`
		FormFactor    string  `json:"form_factor"`
	}
	if err := json.Unmarshal(c.Specs, &specs); err != nil {
		return false
	}
	if specs.MaxThroughput < float64(rule.MinSSDThroughput) {
		return false
	}
	for _, ff := range rule.SSDFormFactors {
		if strings.EqualFold(specs.FormFactor, ff) {
			return true
		}
	}
	return false
}

func rankBuildByScenario(rule rules.ScenarioRule, combo []domain.Component) int {
	score := 0.0
	total := 0.0

	get := func(cat string) *domain.Component {
		for _, c := range combo {
			if strings.EqualFold(c.Category, cat) {
				return &c
			}
		}
		return nil
	}

	// CPU
	if cpu := get("cpu"); cpu != nil {
		specs := parseSpecs(cpu.Specs)
		if tdp, ok := specs["tdp"].(float64); ok {
			tdpInt := int(tdp)
			total += 1
			score += normalize(tdpInt, rule.MinCPUTDP, rule.MaxCPUTDP)
		}
	}

	// RAM
	if ram := get("ram"); ram != nil {
		specs := parseSpecs(ram.Specs)
		if cap, ok := specs["capacity"].(float64); ok {
			capInt := int(cap)
			total += 1
			score += normalize(capInt, rule.MinRAM, rule.MaxRAM)
		}
	}

	// GPU
	if gpu := get("gpu"); gpu != nil {
		specs := parseSpecs(gpu.Specs)
		if mem, ok := specs["memory_gb"].(float64); ok {
			memInt := int(mem)
			total += 1
			score += normalize(memInt, rule.MinGPUMemory, rule.MaxGPUMemory)
		}
	}

	// PSU
	if psu := get("psu"); psu != nil {
		specs := parseSpecs(psu.Specs)
		if pwr, ok := specs["power"].(float64); ok {
			pwrInt := int(pwr)
			total += 1
			score += normalize(pwrInt, rule.MinPSUPower, rule.MaxPSUPower)
		}
	}

	// SSD
	if ssd := get("ssd"); ssd != nil {
		specs := parseSpecs(ssd.Specs)
		if tput, ok := specs["max_throughput"].(float64); ok {
			tputInt := int(tput)
			total += 1
			score += normalize(tputInt, rule.MinSSDThroughput, rule.MinSSDThroughput*4) // разумный потолок
		}
	}

	if total == 0 {
		return 0
	}
	avg := score / total

	// Mapping: [0..0.2]=0, [0.2..0.4]=1, ..., [0.8..1.0]=4
	switch {
	case avg < 0.2:
		return 0
	case avg < 0.4:
		return 1
	case avg < 0.6:
		return 2
	case avg < 0.8:
		return 3
	default:
		return 4
	}
}

func normalize(val, min, max int) float64 {
	if max == min {
		return 1
	}
	return float64(val-min) / float64(max-min)
}
