package usecase

import (
	"StartupPCConfigurator/internal/config/repository"
	"StartupPCConfigurator/internal/config/usecase/rules"
	"StartupPCConfigurator/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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

	// ---------- 1. Кандидаты по категории + brand ------------------
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

	// ---------- 2. Pre-filter by use-case --------------------------
	if usecase != nil && *usecase != "" {
		rule, ok := rules.ScenarioRules[*usecase]
		if !ok {
			return nil, fmt.Errorf("unknown usecase %q", *usecase)
		}
		var filtered []domain.Component
		for _, comp := range candidates {
			ok := false
			switch strings.ToLower(comp.Category) {
			case "cpu":
				ok = cpuMatches(comp, rule)
			case "motherboard":
				ok = mbMatches(comp, rule)
			case "ram":
				ok = ramMatches(comp, rule)
			case "gpu":
				ok = gpuMatches(comp, rule)
			case "psu":
				ok = psuMatches(comp, rule)
			case "case":
				ok = caseMatches(comp, rule)
			case "ssd":
				ok = ssdMatches(comp, rule)
			case "hdd":
				ok = hddMatches(comp, rule)
			default:
				ok = true
			}
			if ok {
				filtered = append(filtered, comp)
			}
		}
		candidates = filtered
	}

	// ---------- 3. specsByCat — всё, что уже выбрано ----------------
	specsByCat := map[string]map[string]interface{}{}
	totalDraw := 0.0

	for _, ref := range bases {
		base, err := s.repo.GetComponentByName(ref.Category, ref.Name)
		if err != nil {
			return nil, fmt.Errorf("component not found: %s/%s", ref.Category, ref.Name)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(base.Specs, &m); err != nil {
			return nil, fmt.Errorf("invalid specs for %s/%s: %w",
				ref.Category, ref.Name, err)
		}
		specsByCat[strings.ToLower(ref.Category)] = m

		// суммируем power_draw для PSU
		if strings.EqualFold(category, "psu") &&
			(strings.EqualFold(ref.Category, "cpu") || strings.EqualFold(ref.Category, "gpu")) {

			if d, ok := m["power_draw"].(float64); ok {
				totalDraw += d
			}
		}
	}

	// ---------- 3-bis. Доп. фильтр только для HDD -------------------
	if strings.EqualFold(category, "hdd") {
		// если выбран корпус, смотрим, есть ли 3.5″-отсеки
		if csSpecs, ok := specsByCat["case"]; ok {
			bays3, _ := csSpecs["drive_bays_3_5"].(float64)

			// если отсеков нет, убираем все 3.5″ диски из candidates
			if bays3 < 1 {
				tmp := candidates[:0] // переиспользуем срез без аллокации
				for _, c := range candidates {
					var s struct {
						Form string `json:"form_factor"`
					}
					_ = json.Unmarshal(c.Specs, &s)
					if !strings.EqualFold(s.Form, "3.5") {
						tmp = append(tmp, c)
					}
				}
				candidates = tmp
			}
		}
	}

	// ---------- 4. Собираем filter-map только из нужных полей -------
	filterSpecs := buildFilterForCategory(category, specsByCat, totalDraw)

	// ---------- 5. Фильтрация кандидатов ----------------------------
	return s.repo.FilterPoolByCompatibility(candidates, domain.CompatibilityFilter{
		Category: category,
		Specs:    filterSpecs,
	})
}

// ------------------------------------------------------------------
// buildFilterForCategory – единая точка, где мы решаем,
// какие поля действительно проверять для каждой категории.
// ------------------------------------------------------------------
func buildFilterForCategory(
	target string,
	specsByCat map[string]map[string]interface{},
	totalDraw float64,
) map[string]interface{} {

	switch strings.ToLower(target) {

	case "case":
		out := map[string]interface{}{}
		if mb, ok := specsByCat["motherboard"]; ok {
			out["form_factor"] = mb["form_factor"]
		}
		if cpu, ok := specsByCat["cpu"]; ok {
			out["cooler_max_height"] = cpu["cooler_height"]
		}
		if gpu, ok := specsByCat["gpu"]; ok {
			out["gpu_max_length"] = gpu["length_mm"]
		}
		if bays, ok := specsByCat["hdd"]["form_factor"]; ok && bays == "3.5" {
			// если уже выбран 3.5" HDD – нужен хотя бы один 3.5-бэй
			out["drive_bays_3_5"] = 1.0
		}
		return out

	case "gpu":
		if mb, ok := specsByCat["motherboard"]; ok {
			return map[string]interface{}{
				"interface": mb["pcie_version"],
			}
		}
		return nil

	case "ssd":
		if mb, ok := specsByCat["motherboard"]; ok {
			ff := "2.5"
			if slots, _ := mb["m2_slots"].(float64); slots >= 1 {
				ff = "M.2"
			}
			if v, ok := mb["pcie_version"]; ok {
				return map[string]interface{}{
					"interface":   v,
					"form_factor": ff,
				}
			}
			return map[string]interface{}{
				"interface":   "SATA III",
				"form_factor": ff,
			}
		}
		return nil

	case "hdd":
		out := map[string]interface{}{"interface": "SATA III"}
		if mb, ok := specsByCat["motherboard"]; ok {
			out["sata_ports"] = mb["sata_ports"]
		}
		if cs, ok := specsByCat["case"]; ok {
			out["drive_bays_3_5"] = cs["drive_bays_3_5"]
		}
		return out

	case "psu":
		return map[string]interface{}{"power": totalDraw + 150}

	case "cooler":
		mbSocket := ""
		if cpu, ok := specsByCat["cpu"]; ok {
			mbSocket, _ = cpu["socket"].(string)
		}
		maxH := 0.0
		if cs, ok := specsByCat["case"]; ok {
			maxH, _ = cs["cooler_max_height"].(float64)
		}
		return map[string]interface{}{
			"socket":    mbSocket,
			"height_mm": maxH,
		}

	default:
		// для прочих категорий ничего особого не нужно
		return nil
	}
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

// ========== вспом. кэш ===============
var specCache = sync.Map{} // id -> map[string]interface{}

func specs(c domain.Component) map[string]interface{} {
	if v, ok := specCache.Load(c.ID); ok {
		return v.(map[string]interface{})
	}
	m := parseSpecs(c.Specs)
	specCache.Store(c.ID, m)
	return m
}

// mini-ключ для пары «форм-фактор корпуса / форм-фактор платы»
type ffPair struct{ caseFF, mbFF string }

// дикт, в котором TRUE означает «плата mbFF физически помещается в корпус caseFF»
var formOK = map[ffPair]bool{
	{"ATX", "ATX"}:       true,
	{"ATX", "MICRO-ATX"}: true,
	{"ATX", "MINI-ITX"}:  true,

	{"MICRO-ATX", "MICRO-ATX"}: true,
	{"MICRO-ATX", "MINI-ITX"}:  true,

	{"MINI-ITX", "MINI-ITX"}: true,
}

// ==========================================================================
// helpers.go (или в том же файле — как удобнее)
// -------------------------------------------------------------------------
// заглушка + усечение «второстепенных» пулов (GPU / SSD / HDD)
func capped(list []domain.Component, need bool, max int) []domain.Component {
	if !need { // вес = 0  →  только пустышка
		return []domain.Component{{}}
	}
	if max > 0 && len(list) > max { // обрезаем хвост
		list = list[:max]
	}
	return append([]domain.Component{{}}, list...)
}

// усечение «обязательных» пулов (CPU / MB / RAM / PSU / Case)
func capPrimary(list []domain.Component, max int) []domain.Component {
	if max > 0 && len(list) > max {
		rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
		return list[:max]
	}
	return list
}

// быстрый ключ без sort.Slice / аллокаций
func fastKey(combo []domain.Component) string {
	var b strings.Builder
	for _, c := range combo { // порядок фиксирован: CPU, MB, RAM…
		if c.ID == 0 {
			continue
		} // пропускаем «заглушки»
		b.WriteString(c.Category) // category уже в правильном регистре
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(c.ID))
		b.WriteByte(';')
	}
	return b.String()
}

func init() { rand.Seed(time.Now().UnixNano()) }

// ==========================================================================
// service.go  (обновлённая GetUseCaseBuild)
// --------------------------------------------------------------------------

func (s *configService) GetUseCaseBuild(usecase string, limit int) ([]domain.NamedBuild, error) {
	rule, ok := rules.ScenarioRules[usecase]
	if !ok {
		return nil, fmt.Errorf("unknown use-case %q", usecase)
	}

	/* ---------- 1. загрузка и грубый пред-фильтр ---------- */

	all, err := s.repo.GetComponents("", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to load components: %w", err)
	}

	bkt := map[string][]domain.Component{}
	for _, c := range all {
		switch strings.ToLower(c.Category) {
		case "cpu":
			if cpuMatches(c, rule) {
				bkt["cpu"] = append(bkt["cpu"], c)
			}
		case "motherboard":
			if mbMatches(c, rule) {
				bkt["motherboard"] = append(bkt["motherboard"], c)
			}
		case "ram":
			if ramMatches(c, rule) {
				bkt["ram"] = append(bkt["ram"], c)
			}
		case "psu":
			if psuMatches(c, rule) {
				bkt["psu"] = append(bkt["psu"], c)
			}
		case "case":
			if caseMatches(c, rule) {
				bkt["case"] = append(bkt["case"], c)
			}
		case "gpu":
			if gpuMatches(c, rule) {
				bkt["gpu"] = append(bkt["gpu"], c)
			}
		case "ssd":
			if ssdMatches(c, rule) {
				bkt["ssd"] = append(bkt["ssd"], c)
			}
		case "hdd":
			if hddMatches(c, rule) {
				bkt["hdd"] = append(bkt["hdd"], c)
			}
		}
	}

	/* ---------- 1-bis. формируем пулы с ограничением размера ---------- */

	w := scenarioWeights[usecase]
	gpus := capped(bkt["gpu"], w.GPU > 0, 4) // ≤ 4 вариантов
	ssds := capped(bkt["ssd"], w.SSD > 0, 3) // ≤ 3 вариантов
	hdds := capped(bkt["hdd"], w.HDD > 0, 2) // ≤ 2 вариантов

	cpus := capPrimary(bkt["cpu"], 5) // ≤ 5 CPU
	mbs := capPrimary(bkt["motherboard"], 5)
	rams := capPrimary(bkt["ram"], 5)
	psus := capPrimary(bkt["psu"], 4)
	cases := capPrimary(bkt["case"], 4)

	/* ---------- 2. комбинаторика с ранними continue ---------- */

	seen := map[string]struct{}{}
	ranked := map[int][]domain.Component{}
	closed := 0 // сколько разных рангов уже заполнено

	// подготовим соответствие «форма корпуса → набор плат, которые влезут»
	fits := make(map[ffPair]bool, len(formOK))
	for k, v := range formOK {
		fits[k] = v
	}

	for _, cpu := range cpus {
		scpu := specs(cpu)

		for _, mb := range mbs {
			if scpu["socket"] != specs(mb)["socket"] {
				continue
			}

			for _, ram := range rams {
				for _, psu := range psus {

					for _, cs := range cases {
						// внутри комбинаторного цикла:
						fc := strings.ToUpper(specs(cs)["form_factor"].(string))
						fm := strings.ToUpper(specs(mb)["form_factor"].(string))
						if !formOK[ffPair{fc, fm}] {
							continue
						}

						for _, gpu := range gpus {
							if w.GPU == 0 && gpu.ID != 0 {
								continue
							}
							for _, ssd := range ssds {
								if w.SSD == 0 && ssd.ID != 0 {
									continue
								}
								for _, hdd := range hdds {
									if w.HDD == 0 && hdd.ID != 0 {
										continue
									}

									combo := []domain.Component{cpu, mb, ram, psu, cs}
									if gpu.ID != 0 {
										combo = append(combo, gpu)
									}
									if ssd.ID != 0 {
										combo = append(combo, ssd)
									}
									if hdd.ID != 0 {
										combo = append(combo, hdd)
									}

									key := fastKey(combo)
									if _, ok := seen[key]; ok {
										continue
									}
									seen[key] = struct{}{}

									r := rankBuildByScenario(rule, combo, usecase)
									if _, ok := ranked[r]; !ok {
										ranked[r] = combo
										closed++
										if closed == limit {
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
	}
DONE:
	if len(ranked) == 0 {
		return nil, fmt.Errorf("no builds for %s", usecase)
	}

	/* ---------- 3. собираем ответ ---------- */

	out := make([]domain.NamedBuild, 0, len(buildLabels))
	for i := 0; i < len(buildLabels); i++ {
		if c, ok := ranked[i]; ok {
			out = append(out, domain.NamedBuild{buildLabels[i], c})
		}
	}
	return out, nil
}

/* ----------  init() для сидирования rand  ---------- */

func init() {
	rand.Seed(time.Now().UnixNano()) // ★ NEW: чтобы Shuffle был непредсказуемым
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

func hddMatches(c domain.Component, rule rules.ScenarioRule) bool {
	var specs struct {
		CapacityGB float64 `json:"capacity_gb"`
		Interface  string  `json:"interface"`
	}
	if err := json.Unmarshal(c.Specs, &specs); err != nil {
		return false
	}

	// 1) интерфейс ― любой SATA-вариант
	if !strings.HasPrefix(strings.ToUpper(specs.Interface), "SATA") {
		return false
	}

	// 2) диапазон ёмкости
	capGB := int(specs.CapacityGB)
	return capGB >= rule.MinHDDCapacity && capGB <= rule.MaxHDDCapacity
}

// веса одного сценария
type Weights struct {
	CPU, GPU, RAM, SSD, HDD, PSU float64
}

// глобальная карта «сценарий → веса»
var scenarioWeights = map[string]Weights{
	"office": {
		CPU: 1, GPU: 0.5, RAM: 1, SSD: 0.5, HDD: 0, PSU: 0.5, // офисному ПК хватит SSD
	},
	"htpc": {
		CPU: 2, GPU: 1, RAM: 1, SSD: 1, HDD: 1, PSU: 0.5, // медиаплееру нужен объём для фильмов
	},
	"gaming": {
		CPU: 3, GPU: 5, RAM: 2, SSD: 2, HDD: 0, PSU: 1, // игры на SSD
	},
	"streamer": {
		CPU: 4, GPU: 3, RAM: 3, SSD: 2, HDD: 0, PSU: 1.5, // стримы с SSD
	},
	"design": {
		CPU: 3, GPU: 4, RAM: 2, SSD: 1, HDD: 1, PSU: 1, // дизайн-файлы могут быть крупными
	},
	"video": {
		CPU: 4, GPU: 5, RAM: 4, SSD: 2, HDD: 2, PSU: 1.5, // видеопроекты занимают много места
	},
	"cad": {
		CPU: 4, GPU: 3, RAM: 4, SSD: 2, HDD: 0, PSU: 1, // CAD-модели требуют и скорость, и объём
	},
	"dev": {
		CPU: 3, GPU: 1, RAM: 3, SSD: 2, HDD: 0.5, PSU: 1, // виртуалки на SSD, но полезен резерв на HDD
	},
	"enthusiast": {
		CPU: 4, GPU: 5, RAM: 3, SSD: 1.5, HDD: 0, PSU: 1, // главное — производительность
	},
	"nas": {
		CPU: 1, GPU: 0.5, RAM: 2, SSD: 1, HDD: 1, PSU: 1, // основной сценарий для HDD
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

	// ---------- HDD ----------
	if w.HDD > 0 {
		if hdd := get("hdd"); hdd != nil {
			cap := toInt(parseSpecs(hdd.Specs)["capacity_gb"])
			s := normalize(cap, rule.MinHDDCapacity, rule.MaxHDDCapacity)
			score += s * w.HDD
			total += w.HDD
		}
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

	bucket := int(math.Round(avg * 4)) // 0..4
	if bucket > 4 {
		bucket = 4
	}
	return bucket // 0..1 делим на 5 интервалов
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
