package usecase

import (
	"StartupPCConfigurator/internal/config/repository"
	"StartupPCConfigurator/internal/config/usecase/rules"
	"StartupPCConfigurator/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
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
	GetUseCaseBuild(usecaseName string, limit int) ([]domain.NamedBuild, error)
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

func (s *configService) GetUseCaseBuild(usecaseName string, limit int) ([]domain.NamedBuild, error) {
	rule, ok := rules.ScenarioRules[usecaseName]
	if !ok {
		return nil, fmt.Errorf("unknown use case %q", usecaseName)
	}

	all, err := s.repo.GetComponents("", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to load components: %w", err)
	}

	// 1) Пред-фильтрация по категориям
	var cpus, mbs, rams, psus, cases, gpus, ssds []domain.Component
	for _, c := range all {
		switch strings.ToLower(c.Category) {
		case "cpu":
			if cpuMatches(c, rule) {
				cpus = append(cpus, c)
			}
		case "motherboard":
			if mbMatches(c, rule) { // NEW: фильтрация плат сразу
				mbs = append(mbs, c)
			}
		case "ram":
			if ramMatches(c, rule) {
				rams = append(rams, c)
			}
		case "psu":
			if psuMatches(c, rule) {
				psus = append(psus, c)
			}
		case "case":
			if caseMatches(c, rule) {
				cases = append(cases, c)
			}
		case "gpu":
			if gpuMatches(c, rule) {
				gpus = append(gpus, c)
			}
		case "ssd":
			if ssdMatches(c, rule) {
				ssds = append(ssds, c)
			}
		}
	}

	// 1-бис) GPU/SSD опциональны, только если это правда нужно
	gpuPool := gpus
	if rule.MinGPUMemory == 0 {
		gpuPool = append([]domain.Component{{}}, gpuPool...)
	}
	ssdPool := ssds
	if rule.MinSSDThroughput == 0 {
		ssdPool = append([]domain.Component{{}}, ssdPool...)
	}

	seen := make(map[string]struct{})
	ranked := make(map[int][]domain.Component)
	need := limit

	// 2) Генерация комбинаций
	for _, cpu := range cpus {
		specCPU := parseSpecs(cpu.Specs) // MOVED: парсим один раз
		for _, mb := range mbs {
			specMB := parseSpecs(mb.Specs)
			// NEW: проверяем совместимость сокетов
			if specCPU["socket"] != specMB["socket"] {
				continue
			}

			for _, ram := range rams {
				for _, psu := range psus {
					for _, cs := range cases {
						specsCase := parseSpecs(cs.Specs)
						// MB ↔ Case
						if !contains(caseSupportedMap[strings.ToUpper(specsCase["form_factor"].(string))],
							specMB["form_factor"].(string)) {
							continue
						}

						// GPU / SSD
						for _, gpu := range gpuPool {
							for _, ssd := range ssdPool {

								combo := []domain.Component{cpu, mb, ram, psu, cs}
								if gpu.ID != 0 {
									combo = append(combo, gpu)
								}
								if ssd.ID != 0 {
									combo = append(combo, ssd)
								}

								// дедуп по «ключевым» категориям
								h := hashCombo(combo)
								if _, ok := seen[h]; ok {
									continue
								}
								seen[h] = struct{}{}

								// ранжирование
								rank := rankBuildByScenario(rule, combo, usecaseName)
								if rank >= len(buildLabels) {
									rank = len(buildLabels) - 1
								}
								if _, ok := ranked[rank]; !ok {
									ranked[rank] = combo
									need--
									if need == 0 {
										goto DONE
									}
								}
							}
						}
					}
				}
			}
		}
	}
DONE:

	if len(ranked) == 0 {
		return nil, fmt.Errorf("no builds found for use case %q", usecaseName)
	}

	// 3) Возвращаем от 0-го до limit-1-го ранга
	var out []domain.NamedBuild
	for i := 0; i < limit; i++ {
		if combo, ok := ranked[i]; ok {
			out = append(out, domain.NamedBuild{Name: buildLabels[i], Components: combo})
		}
	}
	return out, nil
}

var caseSupportedMap = map[string][]string{
	"ATX":       {"ATX", "Micro-ATX", "Mini-ITX"},
	"MICRO-ATX": {"Micro-ATX", "Mini-ITX"},
	"MINI-ITX":  {"Mini-ITX"},
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

	// 1) сокет платы обязан быть из whitelist сценария
	if sock, _ := specs["socket"].(string); sock != "" &&
		!contains(rule.CPUSocketWhitelist, sock) {
		return false
	}

	// 2) тип памяти
	if rule.RAMType != "" {
		if mt, _ := specs["ram_type"].(string); mt != rule.RAMType {
			return false
		}
	}
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

// веса одного сценария
type Weights struct {
	CPU, GPU, RAM, SSD, PSU float64
}

// глобальная карта «сценарий → веса»
var scenarioWeights = map[string]Weights{
	"office": {
		CPU: 1, GPU: 0.5, RAM: 1, SSD: 0.5, PSU: 0.5,
	},
	"gaming": {
		CPU: 3, GPU: 5, RAM: 2, SSD: 2, PSU: 1,
	},
	"htpc": {
		CPU: 2, GPU: 1, RAM: 1, SSD: 1, PSU: 0.5,
	},
	"streamer": {
		CPU: 4, GPU: 3, RAM: 3, SSD: 2, PSU: 1.5,
	},
	"design": {
		CPU: 3, GPU: 4, RAM: 2, SSD: 1, PSU: 1,
	},
	"video": {
		CPU: 4, GPU: 5, RAM: 4, SSD: 2, PSU: 1.5,
	},
	"cad": {
		CPU: 4, GPU: 3, RAM: 4, SSD: 2, PSU: 1,
	},
	"dev": {
		CPU: 3, GPU: 1, RAM: 3, SSD: 2, PSU: 1,
	},
	"enthusiast": {
		CPU: 4, GPU: 5, RAM: 3, SSD: 1.5, PSU: 1,
	},
	"nas": {
		CPU: 1, GPU: 0.5, RAM: 2, SSD: 1, PSU: 1,
	},
}

func rankBuildByScenario(rule rules.ScenarioRule, combo []domain.Component, scen string) int {
	w := scenarioWeights[scen] // <-- берём весы

	// ищем нужные компоненты
	get := func(cat string) *domain.Component {
		for i := range combo {
			if strings.EqualFold(combo[i].Category, cat) {
				return &combo[i]
			}
		}
		return nil
	}

	score, total := 0.0, 0.0

	// ---------- CPU ----------
	if cpu := get("cpu"); cpu != nil {
		spec := parseSpecs(cpu.Specs)
		cores := toInt(spec["cores"])
		threads := toInt(spec["threads"])
		perf := cores*2 + threads   // грубая «оценка мощности»
		s := normalize(perf, 8, 48) // 8 – минимум, 48 – верх для WS
		score += s * w.CPU
		total += w.CPU
	}

	// ---------- GPU ----------
	if gpu := get("gpu"); gpu != nil {
		mem := toInt(parseSpecs(gpu.Specs)["memory_gb"])
		s := normalize(mem, rule.MinGPUMemory, rule.MaxGPUMemory)
		score += s * w.GPU
		total += w.GPU
	}

	// ---------- RAM ----------
	if ram := get("ram"); ram != nil {
		cap := toInt(parseSpecs(ram.Specs)["capacity"])
		s := normalize(cap, rule.MinRAM, rule.MaxRAM)
		score += s * w.RAM
		total += w.RAM
	}

	// ---------- SSD ----------
	if ssd := get("ssd"); ssd != nil {
		tput := toInt(parseSpecs(ssd.Specs)["max_throughput"])
		s := normalize(tput, rule.MinSSDThroughput, 8000) // 8 GB/s = PCIe 4 «потолок»
		score += s * w.SSD
		total += w.SSD
	}

	// ---------- PSU ----------
	if psu := get("psu"); psu != nil {
		specP := parseSpecs(psu.Specs)
		pwr := toInt(specP["power"])

		// суммарный потреб-т для оценки запаса
		cpuTDP, gpuTDP := 0, 0
		if cpu := get("cpu"); cpu != nil {
			cpuTDP = toInt(parseSpecs(cpu.Specs)["tdp"])
		}
		if gpu := get("gpu"); gpu != nil {
			gpuTDP = toInt(parseSpecs(gpu.Specs)["power_draw"])
		}
		headroom := pwr - (cpuTDP + gpuTDP)
		s := normalize(headroom, 100, 450) // 100 Вт – минимум запаса
		score += s * w.PSU
		total += w.PSU
	}

	if total == 0 {
		return 0 // теоретически не случится
	}
	avg := score / total

	// 0..1 делим на 5 интервалов
	return int(avg * 5) // 0–4 автоматически
}

// helper
func toInt(v any) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	if i, ok := v.(int); ok {
		return i
	}
	return 0
}

func normalize(val, min, max int) float64 {
	if max == min {
		return 1
	}
	f := float64(val-min) / float64(max-min)
	if f < 0 { // меньше минимума
		return 0
	}
	if f > 1 { // выше максимума
		return 1
	}
	return f
}

func hashCombo(combo []domain.Component) string {
	// --- клонируем, чтобы не трогать оригинал ---------------------------- // NEW
	tmp := make([]domain.Component, len(combo))
	copy(tmp, combo)

	// сортируем по категории
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Category < tmp[j].Category
	})

	var sb strings.Builder
	for _, c := range tmp {
		if c.ID == 0 { // пропускаем «заглушки» GPU/SSD/etc.   // NEW
			continue
		}
		sb.WriteString(strings.ToLower(c.Category))
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(c.ID))
		sb.WriteString(";")
	}
	return sb.String()
}
