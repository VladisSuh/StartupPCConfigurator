package usecase

import (
	"StartupPCConfigurator/internal/config/repository"
	"StartupPCConfigurator/internal/config/usecase/rules"
	"StartupPCConfigurator/internal/domain"
	"context"
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

	// --- (1) формируем фильтр для репозитория -----------------------
	f := repository.ComponentFilter{
		NameILike: search,
		BrandEQ:   brand,
	}
	if category != "" {
		f.Categories = []string{category}
	}

	// --- (2) получаем записи напрямую от БД -------------------------
	comps, err := s.repo.GetComponentsFiltered(context.Background(), f) // <<< PATCH
	if err != nil {
		return nil, err
	}

	// --- (3) если нет сценария — сразу возвращаем -------------------
	if usecase == "" {
		return comps, nil
	}
	rule, ok := rules.ScenarioRules[usecase]
	if !ok {
		return nil, fmt.Errorf("unknown usecase %q", usecase)
	}

	// --- (4) сценарием отфильтровываем ровно по категории ------------
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
		case "hdd":
			if hddMatches(c, rule) {
				out = append(out, c)
			}
		default:
			out = append(out, c)
		}
	}
	return out, nil
}

// ВЕРСИЯ ПОЛНОСТЬЮ РАБОЧАЯ
func (s *configService) FetchCompatibleComponentsMulti(
	category string,
	bases []domain.ComponentRef,
	brand *string,
	usecase *string,
) ([]domain.Component, error) {
	// 1) Получаем пул кандидатов
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

	// 2) Сценарная фильтрация
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
			case "hdd":
				if hddMatches(comp, rule) {
					filtered = append(filtered, comp)
				}
			default:
				filtered = append(filtered, comp)
			}
		}
		candidates = filtered
	}

	// 2.1) Спецлогика для кулеров
	if strings.EqualFold(category, "cooler") {
		var refs []domain.ComponentRef
		for _, ref := range bases {
			if strings.EqualFold(ref.Category, "cpu") ||
				strings.EqualFold(ref.Category, "case") {
				refs = append(refs, ref)
			}
		}
		bases = refs
	}

	// 3) Собираем «сырые» спецификации по категориям + totalDraw для PSU
	specsByCat := make(map[string]map[string]interface{})
	var totalDraw float64
	for _, ref := range bases {
		comp, err := s.repo.GetComponentByName(ref.Category, ref.Name)
		if err != nil {
			return nil, fmt.Errorf("component not found: %s/%s", ref.Category, ref.Name)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(comp.Specs, &m); err != nil {
			return nil, fmt.Errorf("invalid specs for %s/%s: %w", ref.Category, ref.Name, err)
		}
		specsByCat[strings.ToLower(ref.Category)] = m

		if strings.EqualFold(category, "psu") &&
			(strings.EqualFold(ref.Category, "cpu") || strings.EqualFold(ref.Category, "gpu")) {
			if d, ok := m["power_draw"].(float64); ok {
				totalDraw += d
			}
		}
	}

	// 4) Ручная фильтрация для HDD — сразу возвращаем
	if strings.EqualFold(category, "hdd") {
		var out []domain.Component
		mbSpec, hasMB := specsByCat["motherboard"]
		caseSpec, hasCase := specsByCat["case"]

		for _, comp := range candidates {
			// распарсим specs диска
			var m map[string]interface{}
			json.Unmarshal(comp.Specs, &m)

			// 4.1) только SATA
			iface, _ := m["interface"].(string)
			if !strings.HasPrefix(strings.ToUpper(iface), "SATA") {
				continue
			}
			// 4.2) проверяем порты на MB
			if hasMB {
				if ports, ok := mbSpec["sata_ports"].(float64); !ok || ports < 1 {
					continue
				}
			}
			// 4.3) проверяем 3.5″ bay на Case
			if hasCase {
				if bays, ok := caseSpec["drive_bays_3_5"].(float64); ok && bays < 1 {
					continue
				}
			}
			out = append(out, comp)
		}
		return out, nil
	}

	// 5) Собираем merged-спецификацию для всех прочих категорий
	merged := make(map[string]interface{})
	switch strings.ToLower(category) {
	case "case":
		specC := make(map[string]interface{})

		// 1. Совместимость с материнкой:
		//    корпус хранит max_motherboard_form_factors []string
		if mbSpec, ok := specsByCat["motherboard"]; ok {
			if ff, ok2 := mbSpec["form_factor"]; ok2 {
				specC["max_motherboard_form_factors"] = []interface{}{ff}
			}
		}

		// 2. Высота кулера:
		//    корпус.Specs["cooler_max_height"]
		if cpuSpec, ok := specsByCat["cpu"]; ok {
			if h, ok2 := cpuSpec["cooler_height"]; ok2 {
				specC["cooler_max_height"] = h
			}
		}

		// 3. Длина видеокарты:
		if gpuSpec, ok := specsByCat["gpu"]; ok {
			if l, ok2 := gpuSpec["length_mm"]; ok2 {
				specC["gpu_max_length"] = l
			}
		}

		// 4. Длина и форм-фактор блока питания:
		if psuSpec, ok := specsByCat["psu"]; ok {
			if l, ok2 := psuSpec["length_mm"]; ok2 {
				specC["max_psu_length"] = l
			}
			if pf, ok2 := psuSpec["form_factor"]; ok2 {
				specC["psu_form_factor"] = pf
			}
		}

		merged = specC

	case "gpu":
		if mbSpec, ok := specsByCat["motherboard"]; ok {
			if v, ok2 := mbSpec["pcie_version"]; ok2 {
				merged["interface"] = v
			}
		}

	case "psu":
		merged["power"] = totalDraw + 150

	case "ssd":
		if mbSpec, ok := specsByCat["motherboard"]; ok {
			if v, ok2 := mbSpec["pcie_version"]; ok2 {
				merged["interface"] = v
			} else if ports, ok3 := mbSpec["sata_ports"]; ok3 {
				merged["interface"] = "SATA III"
				merged["sata_ports"] = ports
			}
		}
		if ssdSpec, ok := specsByCat["ssd"]; ok {
			if slots, ok2 := ssdSpec["m2_slots"].(float64); ok2 && slots >= 1 {
				merged["form_factor"] = "M.2"
			} else {
				merged["form_factor"] = "2.5"
			}
		}

	case "cooler":
		if cpuSpec, ok := specsByCat["cpu"]; ok {
			if h, ok2 := cpuSpec["cooler_height"]; ok2 {
				merged["height_mm"] = h
			}
		}
		if mbSpec, ok := specsByCat["motherboard"]; ok {
			if sock, ok2 := mbSpec["socket"]; ok2 {
				merged["socket"] = sock
			}
		}
	case "motherboard":
		if cpuSpec, ok := specsByCat["cpu"]; ok {
			if sock, ok2 := cpuSpec["socket"]; ok2 {
				merged["socket"] = sock
			}
		}
	}

	// 6) Общий FilterPoolByCompatibility для всех остальных
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

// ВЕРСИЯ ГДЕ ЗАПРОС УПАЛ ДО 8 СЕКУНД С REPO + попытка score сделать + попытка убрать parse
/*func (s *configService) GetUseCaseBuild(usecaseName string, limit int) ([]domain.NamedBuild, error) {
	// ───────────────────────────────── rule / weight ──────────────────────────
	rule, ok := rules.ScenarioRules[usecaseName]
	if !ok {
		return nil, fmt.Errorf("unknown use case %q", usecaseName)
	}
	w := scenarioWeights[usecaseName]
	type cpuInfo struct {
		comp      domain.Component
		socket    string
		tdp       int
		perfScore float64
	}

	// ────────────────────────────── фиксированный CPU ─────────────────────────
	predef := PredefinedCPUs[usecaseName] // [][]string
	cpuRankMap := make(map[int]int)       // cpuID → ранг 0..4
	cpus := make([]cpuInfo, 0, 5)

	for rank, names := range predef {
		for _, name := range names {
			c, err := s.repo.GetComponentByName("cpu", name)
			if err != nil {
				continue // CPU отсутствует в БД
			}
			spec := parseSpecs(c.Specs)
			cpus = append(cpus, cpuInfo{
				comp:      c,
				socket:    spec["socket"].(string),
				tdp:       toInt(spec["tdp"]),
				perfScore: float64(rank), // ранг как метрика «качества»
			})
			cpuRankMap[c.ID] = rank
		}
	}
	if len(cpus) == 0 {
		return nil, fmt.Errorf("no predefined CPUs found in DB for %q", usecaseName)
	}

	ctx := context.Background()
	start := time.Now()

	// ────────────────────────────── пул остальных категорий ───────────────────
	var (
		mbsRaw, ramsRaw, psusRaw, casesRaw,
		gpusRaw, ssdsRaw, hddsRaw, coolersRaw []domain.Component
	)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		mbsRaw, err = s.repo.GetComponentsFiltered(ctx, repository.ComponentFilter{
			Categories:      []string{"motherboard"},
			SocketWhitelist: rule.CPUSocketWhitelist,
			RAMType:         rule.RAMType,
		})
		return err
	})
	g.Go(func() error {
		var err error
		ramsRaw, err = s.repo.GetComponentsFiltered(ctx, repository.ComponentFilter{
			Categories: []string{"ram"},
			RAMType:    rule.RAMType,
			MinRAM:     rule.MinRAM,
			MaxRAM:     rule.MaxRAM,
		})
		return err
	})
	g.Go(func() error {
		var err error
		psusRaw, err = s.repo.GetComponentsFiltered(ctx, repository.ComponentFilter{
			Categories:  []string{"psu"},
			MinPSUPower: rule.MinPSUPower,
			MaxPSUPower: rule.MaxPSUPower,
		})
		return err
	})
	g.Go(func() error {
		var err error
		casesRaw, err = s.repo.GetComponentsFiltered(ctx, repository.ComponentFilter{
			Categories:      []string{"case"},
			CaseFormFactors: rule.CaseFormFactors,
		})
		return err
	})
	g.Go(func() error {
		var err error
		gpusRaw, err = s.repo.GetComponentsFiltered(ctx, repository.ComponentFilter{
			Categories: []string{"gpu"},
			MinGPUMem:  rule.MinGPUMemory,
			MaxGPUMem:  rule.MaxGPUMemory,
		})
		return err
	})
	g.Go(func() error {
		var err error
		ssdsRaw, err = s.repo.GetComponentsFiltered(ctx, repository.ComponentFilter{
			Categories: []string{"ssd"},
			MinSSDTP:   rule.MinSSDThroughput,
		})
		return err
	})
	g.Go(func() error {
		var err error
		hddsRaw, err = s.repo.GetComponentsFiltered(ctx, repository.ComponentFilter{
			Categories: []string{"hdd"},
			MinHDDCap:  rule.MinHDDCapacity,
			MaxHDDCap:  rule.MaxHDDCapacity,
		})
		return err
	})
	g.Go(func() error {
		var err error
		coolersRaw, err = s.repo.GetComponentsFiltered(ctx, repository.ComponentFilter{
			Categories:      []string{"cooler"},
			SocketWhitelist: rule.CPUSocketWhitelist,
		})
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	log.Printf("GetUseCaseBuild: fetch non-CPU pools took %v", time.Since(start))

	// ────────────────────── парс specs остальных пулов (параллельно) ──────────
	type (
		mbInfo struct {
			comp               domain.Component
			socket, formFactor string
		}
		ramInfo struct {
			comp  domain.Component
			score float64
		}
		psuInfo struct {
			comp  domain.Component
			score float64
		}
		caseInfo struct {
			comp       domain.Component
			formFactor string
			coolerMax  int
		}
		gpuInfo struct {
			comp  domain.Component
			score float64
		}
		ssdInfo struct {
			comp  domain.Component
			score float64
		}
		hddInfo struct {
			comp  domain.Component
			score float64
		}
		coolerInfo struct {
			comp   domain.Component
			socket string
			height int
		}
	)
	var (
		mbs     []mbInfo
		rams    []ramInfo
		psus    []psuInfo
		cases   []caseInfo
		gpus    []gpuInfo
		ssds    []ssdInfo
		hdds    []hddInfo
		coolers []coolerInfo
	)

	var wg sync.WaitGroup
	wg.Add(7)

	go func() { // MB
		defer wg.Done()
		for _, m := range mbsRaw {
			s := parseSpecs(m.Specs)
			mbs = append(mbs, mbInfo{comp: m, socket: s["socket"].(string), formFactor: s["form_factor"].(string)})
		}
	}()
	go func() { // RAM
		defer wg.Done()
		for _, r := range ramsRaw {
			s := parseSpecs(r.Specs)
			c := toInt(s["capacity"])
			score := normalize(c, rule.MinRAM, rule.MaxRAM) * w.RAM
			rams = append(rams, ramInfo{comp: r, score: score})
		}
	}()
	go func() { // PSU
		defer wg.Done()
		for _, p := range psusRaw {
			s := parseSpecs(p.Specs)
			pow := toInt(s["power"])
			score := normalize(pow, rule.MinPSUPower, rule.MaxPSUPower) * w.PSU
			psus = append(psus, psuInfo{comp: p, score: score})
		}
	}()
	go func() { // Case
		defer wg.Done()
		for _, c := range casesRaw {
			s := parseSpecs(c.Specs)
			cases = append(cases, caseInfo{
				comp:       c,
				formFactor: s["form_factor"].(string),
				coolerMax:  toInt(s["cooler_max_height"]),
			})
		}
	}()
	go func() { // GPU
		defer wg.Done()
		for _, g := range gpusRaw {
			s := parseSpecs(g.Specs)
			mem := toInt(s["memory_gb"])
			score := normalize(mem, rule.MinGPUMemory, rule.MaxGPUMemory) * w.GPU
			gpus = append(gpus, gpuInfo{comp: g, score: score})
		}
	}()
	go func() { // SSD
		defer wg.Done()
		for _, ssd := range ssdsRaw {
			s := parseSpecs(ssd.Specs)
			tp := toInt(s["max_throughput"])
			score := normalize(tp, rule.MinSSDThroughput, 8000) * w.SSD
			ssds = append(ssds, ssdInfo{comp: ssd, score: score})
		}
	}()
	go func() { // HDD + Cooler
		defer wg.Done()
		for _, h := range hddsRaw {
			s := parseSpecs(h.Specs)
			c := toInt(s["capacity_gb"])
			score := normalize(c, rule.MinHDDCapacity, rule.MaxHDDCapacity) * w.HDD
			hdds = append(hdds, hddInfo{comp: h, score: score})
		}
		for _, cl := range coolersRaw {
			s := parseSpecs(cl.Specs)
			coolers = append(coolers, coolerInfo{
				comp:   cl,
				socket: s["socket"].(string),
				height: toInt(s["height_mm"]),
			})
		}
	}()
	wg.Wait()

	// ────────────────────── быстрые мапы совместимости ───────────────────────
	mbBySocket := map[string][]mbInfo{}
	for _, mb := range mbs {
		mbBySocket[mb.socket] = append(mbBySocket[mb.socket], mb)
	}
	coolerBySocket := map[string][]coolerInfo{}
	for _, cl := range coolers {
		coolerBySocket[cl.socket] = append(coolerBySocket[cl.socket], cl)
	}
	caseByForm := map[string][]caseInfo{}
	for _, cs := range cases {
		caseByForm[cs.formFactor] = append(caseByForm[cs.formFactor], cs)
	}

	// ────────────────────────── генерация «жёстких» combo ---------------------
	type hardCombo struct{ cpu, mb, ram, psu domain.Component }
	var hardList []hardCombo
	for _, ci := range cpus {
		for _, mb := range mbBySocket[ci.socket] {
			for _, r := range rams {
				for _, p := range psus {
					hardList = append(hardList, hardCombo{
						cpu: ci.comp, mb: mb.comp, ram: r.comp, psu: p.comp,
					})
				}
			}
		}
	}

	// ─────────────────────── case+cooler, затем SSD/HDD/GPU ───────────────────
	type midCombo struct {
		h   hardCombo
		cs  domain.Component
		clr domain.Component
	}
	var midList []midCombo
	for _, hc := range hardList {
		socket := parseSpecs(hc.cpu.Specs)["socket"].(string)
		form := parseSpecs(hc.mb.Specs)["form_factor"].(string)
		for _, cs := range caseByForm[form] {
			for _, cl := range coolerBySocket[socket] {
				if cl.height > cs.coolerMax {
					continue
				}
				midList = append(midList, midCombo{h: hc, cs: cs.comp, clr: cl.comp})
			}
		}
	}

	// пулы с «пустыми» элементами
	var gpuPool []domain.Component
	for _, g := range gpus {
		gpuPool = append(gpuPool, g.comp)
	}
	if rule.MinGPUMemory == 0 {
		gpuPool = append([]domain.Component{{}}, gpuPool...)
	}

	var hddPool []domain.Component
	for _, h := range hdds {
		hddPool = append(hddPool, h.comp)
	}
	if rule.MinHDDCapacity == 0 {
		hddPool = append([]domain.Component{{}}, hddPool...)
	}

	var ssdPool []domain.Component
	for _, s2 := range ssds {
		ssdPool = append(ssdPool, s2.comp)
	}

	// полный перебор до limit
	var combos [][]domain.Component
outer:
	for _, mc := range midList {
		base := []domain.Component{mc.h.cpu, mc.h.mb, mc.h.ram, mc.h.psu, mc.cs, mc.clr}
		for _, ssd := range ssdPool {
			for _, hdd := range hddPool {
				for _, gpu := range gpuPool {
					c := append(append([]domain.Component{}, base...), ssd)
					if hdd.ID != 0 {
						c = append(c, hdd)
					}
					if gpu.ID != 0 {
						c = append(c, gpu)
					}
					if !ValidateCombo(c, rule) {
						continue
					}
					combos = append(combos, c)
					if len(combos) >= limit {
						break outer
					}
				}
			}
		}
	}
	if len(combos) == 0 {
		return nil, fmt.Errorf("no valid builds produced")
	}

	// ─────────────────────── группировка по рангу CPU ─────────────────────────
	rankBuckets := make(map[int][][]domain.Component) // корректная инициализация
	for _, c := range combos {
		cpu := findCPU(c)
		if cpu == nil {
			continue
		}
		rank, ok := cpuRankMap[cpu.ID] // ищём ранг в мапе
		if !ok {
			// CPU нет в PredefinedCPUs → пропускаем
			continue
		}
		rankBuckets[rank] = append(rankBuckets[rank], c)
	}

	// ────────────────── выбираем лучшую сборку в каждом ранге ────────────────
	out := make([]domain.NamedBuild, 0, 5)
	for rank := 0; rank < 5; rank++ {
		bucket := rankBuckets[rank]
		if len(bucket) == 0 {
			continue
		}
		best := bucket[0]
		bestScore := rankCached(usecaseName, best, nil, nil, nil, nil, nil, nil, nil, nil)
		for _, c := range bucket[1:] {
			sc := rankCached(usecaseName, c, nil, nil, nil, nil, nil, nil, nil, nil)
			if sc > bestScore {
				best, bestScore = c, sc
			}
		}
		out = append(out, domain.NamedBuild{
			Name:       buildLabels[rank],
			Components: best,
		})
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}*/

// NamedComponent представляет компонент с его категорией
type NamedComponent struct {
	Name     string
	Category string
}

// PredefinedBuilds содержит предопределённые сборки по сценариям использования
var PredefinedBuilds = map[string][][]NamedComponent{
	"office": {
		{
			{Name: "AMD Athlon 3000G", Category: "cpu"},
			{Name: "ASUS PRIME A320M-K", Category: "motherboard"},
			{Name: "Kingston ValueRAM 8GB DDR4-2666", Category: "ram"},
			{Name: "ADATA SU650 240GB", Category: "ssd"},
			{Name: "Seasonic SSP-300ET", Category: "psu"},
			{Name: "DeepCool Matrexx 30", Category: "case"},
		},
		{
			{Name: "Intel Pentium Gold G7400T", Category: "cpu"},
			{Name: "Gigabyte H610M K", Category: "motherboard"},
			{Name: "Kingston ValueRAM 8GB DDR4-2666", Category: "ram"},
			{Name: "Kingston A400 480GB", Category: "ssd"},
			{Name: "Seasonic SSP-300ET", Category: "psu"},
			{Name: "DeepCool Matrexx 30", Category: "case"},
		},
		{
			{Name: "Intel Core i3-12100T", Category: "cpu"},
			{Name: "Gigabyte H610M K", Category: "motherboard"},
			{Name: "Corsair Vengeance LPX 16GB DDR4-3200", Category: "ram"}, // если нет — можно 2x8GB 2666
			{Name: "Kingston A400 480GB", Category: "ssd"},
			{Name: "Seasonic SSP-300ET", Category: "psu"},
			{Name: "DeepCool Matrexx 30", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 5 5600GE", Category: "cpu"},
			{Name: "ASUS PRIME A320M-K", Category: "motherboard"},
			{Name: "Corsair Vengeance LPX 16GB DDR4-3200", Category: "ram"},
			{Name: "Kingston A400 480GB", Category: "ssd"},
			{Name: "Seasonic SSP-300ET", Category: "psu"},
			{Name: "DeepCool Matrexx 30", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 5 5600X", Category: "cpu"},
			{Name: "MSI B550 Tomahawk", Category: "motherboard"},
			{Name: "Corsair Vengeance LPX 16GB DDR4-3200", Category: "ram"},
			{Name: "Kingston A400 480GB", Category: "ssd"},
			{Name: "Seasonic SSP-300ET", Category: "psu"},
			{Name: "DeepCool Matrexx 30", Category: "case"},
		},
	},

	"gaming": {
		{
			{Name: "Intel Core i5-13400", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Elite AX", Category: "motherboard"},
			{Name: "Corsair Vengeance 32GB DDR5-6000", Category: "ram"},
			{Name: "AMD Radeon RX 6700 XT", Category: "gpu"},
			{Name: "Crucial P3 1TB", Category: "ssd"},
			{Name: "Corsair RM750", Category: "psu"},
			{Name: "Fractal Design North", Category: "case"},
		},
		{
			{Name: "Intel Core i5-14400", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Elite AX", Category: "motherboard"},
			{Name: "Corsair Vengeance 32GB DDR5-6000", Category: "ram"},
			{Name: "AMD Radeon RX 7800 XT", Category: "gpu"},
			{Name: "WD Black SN850X 1TB", Category: "ssd"},
			{Name: "ARCTIC Freezer 34 eSports DUO", Category: "cooler"},
			{Name: "Corsair RM850x SHIFT", Category: "psu"},
			{Name: "Phanteks Eclipse G500A", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 5 7600", Category: "cpu"},
			{Name: "ASUS TUF Gaming B650-PLUS WIFI", Category: "motherboard"},
			{Name: "Corsair Vengeance 32GB DDR5-6000", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4060 8GB", Category: "gpu"},
			{Name: "WD Blue SN570 1TB", Category: "ssd"},
			{Name: "Seasonic Focus GX-650", Category: "psu"},
			{Name: "Phanteks Eclipse P400A", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 7 7700X", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "G.Skill Trident Z5 RGB 32GB DDR5-6000", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4070 Ti SUPER 16GB", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "Seagate IronWolf 4TB", Category: "hdd"},
			{Name: "Thermalright Peerless Assassin 120 SE", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
		{
			{Name: "Intel Core i7-14700K", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Master", Category: "motherboard"},
			{Name: "G.Skill Trident Z5 128GB DDR5-6000", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4080 SUPER 16GB", Category: "gpu"},
			{Name: "Crucial T700 2TB", Category: "ssd"},
			{Name: "Toshiba N300 6TB", Category: "hdd"},
			{Name: "Noctua NH-U12A chromax.black", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
	},
	"htpc": {
		{
			{Name: "AMD Athlon 3000G", Category: "cpu"},
			{Name: "ASRock B550 Phantom Gaming-ITX/ax", Category: "motherboard"},
			{Name: "Kingston ValueRAM 8GB DDR4-2666", Category: "ram"},
			{Name: "ADATA SU650 240GB", Category: "ssd"},
			{Name: "Seagate Barracuda 2TB", Category: "hdd"},
			{Name: "SilverStone ST30SF V2.0", Category: "psu"},
			{Name: "SilverStone SG13", Category: "case"},
		},
		{
			{Name: "Intel Pentium Gold G7400T", Category: "cpu"},
			{Name: "ASUS ROG Strix B660-I Gaming WiFi", Category: "motherboard"},
			{Name: "Kingston ValueRAM 8GB DDR4-2666", Category: "ram"},
			{Name: "Kingston A400 480GB", Category: "ssd"},
			{Name: "Seagate Barracuda 2TB", Category: "hdd"},
			{Name: "SilverStone SX450-G", Category: "psu"}, // чуть выше 350 Вт, но компактный и тихий
			{Name: "SilverStone SG13", Category: "case"},
		},
		{
			{Name: "Intel Core i3-12100T", Category: "cpu"},
			{Name: "ASUS ROG Strix B660-I Gaming WiFi", Category: "motherboard"},
			{Name: "Corsair Vengeance LPX 16GB DDR4-3200", Category: "ram"},
			{Name: "Kingston NV2 500GB", Category: "ssd"},
			{Name: "WD Red Plus 4TB", Category: "hdd"},
			{Name: "Corsair SF500", Category: "psu"}, // чуть выше лимита, но тихий и SFX
			{Name: "Fractal Design Node 202", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 5 5600GE", Category: "cpu"},
			{Name: "ASRock B550 Phantom Gaming-ITX/ax", Category: "motherboard"},
			{Name: "Corsair Vengeance LPX 16GB DDR4-3200", Category: "ram"},
			{Name: "Kingston A2000 500GB", Category: "ssd"},
			{Name: "Toshiba N300 6TB", Category: "hdd"},
			{Name: "SilverStone SX500-G", Category: "psu"},
			{Name: "Fractal Design Node 202", Category: "case"},
		},
		{
			{Name: "Intel Core i7-13700T", Category: "cpu"},
			{Name: "ASUS ROG Strix B660-I Gaming WiFi", Category: "motherboard"},
			{Name: "Corsair Vengeance LPX 16GB DDR4-3200", Category: "ram"},
			{Name: "Samsung 980 Pro 1TB", Category: "ssd"},
			{Name: "Toshiba N300 6TB", Category: "hdd"},
			{Name: "Corsair SF500", Category: "psu"},
			{Name: "Fractal Design Node 202", Category: "case"},
		},
	},
	"streamer": {
		{
			{Name: "AMD Ryzen 7 7700X", Category: "cpu"},
			{Name: "ASUS TUF Gaming B650-PLUS WIFI", Category: "motherboard"},
			{Name: "Corsair Vengeance 32GB DDR5-6000", Category: "ram"},
			{Name: "AMD Radeon RX 6700 10GB", Category: "gpu"},
			{Name: "WD Black SN770 1TB", Category: "ssd"},
			{Name: "Seagate Barracuda 2TB", Category: "hdd"},
			{Name: "ARCTIC Freezer 34 eSports DUO", Category: "cooler"},
			{Name: "Corsair RM750e", Category: "psu"},
			{Name: "Phanteks Eclipse P400A", Category: "case"},
		},
		{
			{Name: "Intel Core i5-14400", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Elite AX", Category: "motherboard"},
			{Name: "Corsair Vengeance 32GB DDR5-6000", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 3080 10GB", Category: "gpu"},
			{Name: "Crucial P3 1TB", Category: "ssd"},
			{Name: "WD Red Plus 4TB", Category: "hdd"},
			{Name: "ARCTIC Freezer 34 eSports DUO", Category: "cooler"},
			{Name: "Corsair RM850x SHIFT", Category: "psu"},
			{Name: "Fractal Design North", Category: "case"},
		},
		{
			{Name: "Intel Core i7-13700", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Master", Category: "motherboard"},
			{Name: "Kingston Fury Renegade 64GB DDR5-6400", Category: "ram"},
			{Name: "AMD Radeon RX 7800 XT", Category: "gpu"},
			{Name: "Samsung 980 Pro 1TB", Category: "ssd"},
			{Name: "Toshiba N300 6TB", Category: "hdd"},
			{Name: "Noctua NH-U12A chromax.black", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Phanteks Eclipse G500A", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 7 7800X3D", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "Kingston Fury Renegade 64GB DDR5-6400", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4070 SUPER", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "Toshiba N300 6TB", Category: "hdd"},
			{Name: "Thermalright Peerless Assassin 120 SE", Category: "cooler"},
			{Name: "Seasonic PRIME TX-1000", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
		{
			{Name: "Intel Core i9-14900K", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Master", Category: "motherboard"},
			{Name: "G.Skill Trident Z5 128GB DDR5-6000", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4080 SUPER 16GB", Category: "gpu"},
			{Name: "Crucial T700 2TB", Category: "ssd"},
			{Name: "Toshiba N300 6TB", Category: "hdd"},
			{Name: "Noctua NH-D15", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
	},
	"design": {
		{
			{Name: "AMD Ryzen 5 7600", Category: "cpu"},
			{Name: "ASUS TUF Gaming B650-PLUS WIFI", Category: "motherboard"},
			{Name: "Corsair Vengeance 32GB DDR5-6000", Category: "ram"},
			{Name: "AMD Radeon RX 6700 10GB", Category: "gpu"},
			{Name: "Kingston A2000 500GB", Category: "ssd"},
			{Name: "Seagate Barracuda 2TB", Category: "hdd"},
			{Name: "Thermalright Peerless Assassin 120 SE", Category: "cooler"},
			{Name: "Seasonic Focus GX-650", Category: "psu"},
			{Name: "Phanteks Eclipse P400A", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 7 7700X", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "Corsair Vengeance RGB Pro 64GB DDR4-3600", Category: "ram"}, // если не хочешь DDR4 — замени
			{Name: "AMD Radeon RX 7800 XT", Category: "gpu"},
			{Name: "WD Black SN850X 1TB", Category: "ssd"},
			{Name: "WD Red Plus 4TB", Category: "hdd"},
			{Name: "Scythe Fuma 2 Rev.B", Category: "cooler"},
			{Name: "Corsair RM850x SHIFT", Category: "psu"},
			{Name: "Phanteks Eclipse G500A", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 7 7800X3D", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "Kingston Fury Renegade 64GB DDR5-6400", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4070 SUPER", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "Seagate IronWolf 4TB", Category: "hdd"},
			{Name: "DeepCool AK620 Digital", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Fractal Design North", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 9 7950X", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "G.Skill Trident Z5 RGB 32GB DDR5-6000", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4080 SUPER 16GB", Category: "gpu"},
			{Name: "Crucial T700 2TB", Category: "ssd"},
			{Name: "Seagate IronWolf 4TB", Category: "hdd"},

			{Name: "Noctua NH-U12A chromax.black", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 9 7950X3D", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "G.Skill Trident Z5 128GB DDR5-6000", Category: "ram"},
			{Name: "AMD Radeon RX 7900 XTX 24GB", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "Seagate IronWolf 4TB", Category: "hdd"},
			{Name: "Noctua NH-D15", Category: "cooler"},
			{Name: "Corsair HX1500i", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
	},
	"video": {
		{
			{Name: "AMD Ryzen 7 7700X", Category: "cpu"},
			{Name: "ASUS TUF Gaming B650-PLUS WIFI", Category: "motherboard"},
			{Name: "Kingston Fury Renegade 64GB DDR5-6400", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4070 SUPER", Category: "gpu"},
			{Name: "WD Black SN850X 1TB", Category: "ssd"},
			{Name: "WD Red Plus 4TB", Category: "hdd"},
			{Name: "ARCTIC Freezer 34 eSports DUO", Category: "cooler"},
			{Name: "Corsair RM850x SHIFT", Category: "psu"},
			{Name: "Phanteks Eclipse P400A", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 7 7800X3D", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "Kingston Fury Renegade 64GB DDR5-6400", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4070 Ti SUPER 16GB", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "Seagate IronWolf 4TB", Category: "hdd"},
			{Name: "Scythe Fuma 2 Rev.B", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Phanteks Eclipse G500A", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 9 7950X", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "G.Skill Trident Z5 RGB 32GB DDR5-6000", Category: "ram"}, // x2 → 64GB
			{Name: "AMD Radeon RX 7900 XT 20GB", Category: "gpu"},
			{Name: "Crucial T700 2TB", Category: "ssd"},
			{Name: "Toshiba N300 6TB", Category: "hdd"},
			{Name: "DeepCool AK620 Digital", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Fractal Design North", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 9 7950X", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "G.Skill Trident Z5 128GB DDR5-6000", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4080 SUPER 16GB", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "Toshiba N300 6TB", Category: "hdd"},
			{Name: "Noctua NH-U12A chromax.black", Category: "cooler"},
			{Name: "Corsair HX1500i", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 9 7950X3D", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "G.Skill Trident Z5 128GB DDR5-6000", Category: "ram"},
			{Name: "AMD Radeon RX 7900 XTX 24GB", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "Toshiba N300 6TB", Category: "hdd"},
			{Name: "Noctua NH-D15", Category: "cooler"},
			{Name: "Corsair HX1500i", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
	},
	"cad": {
		{
			{Name: "AMD Ryzen 7 7700X", Category: "cpu"},
			{Name: "ASUS TUF Gaming B650-PLUS WIFI", Category: "motherboard"},
			{Name: "Corsair Vengeance 32GB DDR5-6000", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 3060", Category: "gpu"},
			{Name: "WD Black SN850X 1TB", Category: "ssd"},
			{Name: "ARCTIC Freezer 34 eSports DUO", Category: "cooler"},
			{Name: "Seasonic Focus GX-650", Category: "psu"},
			{Name: "Phanteks Eclipse P400A", Category: "case"},
		},
		{
			{Name: "Intel Core i5-13400F", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Elite AX", Category: "motherboard"},
			{Name: "Corsair Vengeance 32GB DDR5-6000", Category: "ram"},
			{Name: "AMD Radeon RX 6700 XT", Category: "gpu"},
			{Name: "Crucial T700 2TB", Category: "ssd"},
			{Name: "ARCTIC Freezer 34 eSports DUO", Category: "cooler"},
			{Name: "Corsair RM750", Category: "psu"},
			{Name: "Fractal Design North", Category: "case"},
		},
		{
			{Name: "Intel Core i5-14400", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Elite AX", Category: "motherboard"},
			{Name: "Kingston Fury Renegade 64GB DDR5-6400", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4060 8GB", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "Scythe Fuma 2 Rev.B", Category: "cooler"},
			{Name: "Corsair RM850x SHIFT", Category: "psu"},
			{Name: "Phanteks Eclipse G500A", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 7 7800X3D", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "Kingston Fury Renegade 64GB DDR5-6400", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4070 SUPER", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "DeepCool AK620 Digital", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
		{
			{Name: "Intel Core i7-14700K", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Master", Category: "motherboard"},
			{Name: "G.Skill Trident Z5 128GB DDR5-6000", Category: "ram"}, // можно заменить на 64GB при необходимости
			{Name: "NVIDIA GeForce RTX 4080 SUPER 16GB", Category: "gpu"}, // превышает 12GB — если критично, скажи
			{Name: "Crucial T700 2TB", Category: "ssd"},
			{Name: "Noctua NH-U12A chromax.black", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
	},
	"dev": {
		{
			{Name: "Intel Core i3-12100F", Category: "cpu"},
			{Name: "Gigabyte H610M K", Category: "motherboard"},
			{Name: "Kingston Fury 32GB DDR4-3600", Category: "ram"},
			{Name: "AMD Radeon RX 6400 4GB", Category: "gpu"},
			{Name: "WD Blue SN550 500GB", Category: "ssd"},
			{Name: "WD Blue 1TB", Category: "hdd"},
			{Name: "ARCTIC Freezer 34 eSports DUO", Category: "cooler"},
			{Name: "Corsair RM750e", Category: "psu"},
			{Name: "Phanteks Eclipse P400A", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 5 5600X", Category: "cpu"},
			{Name: "MSI B550 Tomahawk", Category: "motherboard"},
			{Name: "Corsair Vengeance LPX 16GB DDR4-3200", Category: "ram"},
			{Name: "AMD Radeon RX 6400 4GB", Category: "gpu"},
			{Name: "WD Black SN770 1TB", Category: "ssd"},
			{Name: "WD Blue 1TB", Category: "hdd"},
			{Name: "BeQuiet Shadow Rock 3", Category: "cooler"},
			{Name: "Corsair RM750e", Category: "psu"},
			{Name: "Phanteks Eclipse P400A", Category: "case"},
		},
		{
			{Name: "Intel Core i5-12400", Category: "cpu"},
			{Name: "ASUS PRIME B660-PLUS", Category: "motherboard"},
			{Name: "Kingston Fury 32GB DDR4-3600", Category: "ram"},
			{Name: "NVIDIA GeForce GTX 1660 6GB", Category: "gpu"},
			{Name: "Crucial P3 1TB", Category: "ssd"},
			{Name: "WD Blue 1TB", Category: "hdd"},
			{Name: "ARCTIC Freezer 34 eSports DUO", Category: "cooler"},
			{Name: "Corsair RM750e", Category: "psu"},
			{Name: "Phanteks Eclipse P400A", Category: "case"},
		},
		{
			{Name: "Intel Core i5-13400F", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Elite AX", Category: "motherboard"},
			{Name: "Kingston Fury 32GB DDR4-3600", Category: "ram"},
			{Name: "AMD Radeon RX 6400 4GB", Category: "gpu"},
			{Name: "Crucial P5 500GB", Category: "ssd"},
			{Name: "WD Blue 1TB", Category: "hdd"},
			{Name: "Scythe Fuma 2 Rev.B", Category: "cooler"},
			{Name: "Corsair RM750e", Category: "psu"},
			{Name: "Fractal Design North", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 5 5600X", Category: "cpu"},
			{Name: "MSI B550 Tomahawk", Category: "motherboard"},
			{Name: "Corsair Vengeance LPX 16GB DDR4-3200", Category: "ram"},
			{Name: "NVIDIA GeForce GTX 1660 6GB", Category: "gpu"},
			{Name: "WD Black SN850X 1TB", Category: "ssd"},
			{Name: "WD Blue 1TB", Category: "hdd"},
			{Name: "ARCTIC Freezer 34 eSports DUO", Category: "cooler"},
			{Name: "Corsair RM750e", Category: "psu"},
			{Name: "Fractal Design North", Category: "case"},
		},
	},
	"enthusiast": {
		{
			{Name: "Intel Core i9-14900K", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Master", Category: "motherboard"},
			{Name: "G.Skill Trident Z5 128GB DDR5-6000", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4090", Category: "gpu"},
			{Name: "Crucial T700 2TB", Category: "ssd"},
			{Name: "DeepCool AK620 Digital", Category: "cooler"},
			{Name: "Corsair HX1500i", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 7 7800X3D", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "Kingston Fury Renegade 64GB DDR5-6400", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4070 Ti SUPER 16GB", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "Thermalright Peerless Assassin 120 SE", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 9 7950X", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "Kingston Fury Renegade 64GB DDR5-6400", Category: "ram"},
			{Name: "AMD Radeon RX 7900 XT 20GB", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "Noctua NH-U12A chromax.black", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Phanteks Eclipse G500A", Category: "case"},
		},
		{
			{Name: "Intel Core i7-14700K", Category: "cpu"},
			{Name: "Gigabyte Z790 AORUS Master", Category: "motherboard"},
			{Name: "G.Skill Trident Z5 128GB DDR5-6000", Category: "ram"},
			{Name: "NVIDIA GeForce RTX 4080 SUPER 16GB", Category: "gpu"},
			{Name: "Crucial T700 2TB", Category: "ssd"},
			{Name: "Noctua NH-U12A chromax.black", Category: "cooler"},
			{Name: "Corsair HX1500i", Category: "psu"},
			{Name: "Lian Li O11D EVO", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 9 7950X3D", Category: "cpu"},
			{Name: "ASUS ROG Strix X670E-E Gaming WiFi", Category: "motherboard"},
			{Name: "Kingston Fury Renegade 64GB DDR5-6400", Category: "ram"},
			{Name: "AMD Radeon RX 7900 XTX 24GB", Category: "gpu"},
			{Name: "Samsung 990 Pro 4TB", Category: "ssd"},
			{Name: "Noctua NH-D15", Category: "cooler"},
			{Name: "Corsair HX1000i 1000W", Category: "psu"},
			{Name: "Phanteks Eclipse G500A", Category: "case"},
		},
	},
	"nas": {
		{
			{Name: "Intel Pentium Gold G7400T", Category: "cpu"},
			{Name: "Gigabyte H610M K", Category: "motherboard"},
			{Name: "Kingston ValueRAM 8GB DDR4-2666", Category: "ram"},
			{Name: "Kingston A400 480GB", Category: "ssd"},
			{Name: "Seagate IronWolf 4TB", Category: "hdd"},
			{Name: "Seasonic SSP-300ET", Category: "psu"},
			{Name: "DeepCool Matrexx 30", Category: "case"},
		},
		{
			{Name: "AMD Athlon 3000G", Category: "cpu"},
			{Name: "ASUS PRIME A320M-K", Category: "motherboard"},
			{Name: "Kingston ValueRAM 8GB DDR4-2666", Category: "ram"},
			{Name: "ADATA SU650 240GB", Category: "ssd"},
			{Name: "Toshiba N300 6TB", Category: "hdd"},
			{Name: "Seasonic SSP-300ET", Category: "psu"},
			{Name: "DeepCool Matrexx 30", Category: "case"},
		},
		{
			{Name: "Intel Core i3-12100T", Category: "cpu"},
			{Name: "Gigabyte H610M K", Category: "motherboard"},
			{Name: "Kingston Fury 32GB DDR4-3600", Category: "ram"},
			{Name: "Crucial P3 1TB", Category: "ssd"},
			{Name: "WD Red Plus 4TB", Category: "hdd"},
			{Name: "SilverStone ST30SF V2.0", Category: "psu"},
			{Name: "SilverStone SG13", Category: "case"},
		},
		{
			{Name: "AMD Ryzen 5 5600GE", Category: "cpu"},
			{Name: "ASUS PRIME A320M-K", Category: "motherboard"},
			{Name: "Corsair Vengeance LPX 16GB DDR4-3200", Category: "ram"},
			{Name: "Kingston A400 480GB", Category: "ssd"},
			{Name: "WD Red Plus 4TB", Category: "hdd"},
			{Name: "SilverStone ST30SF V2.0", Category: "psu"},
			{Name: "Fractal Design Node 202", Category: "case"},
		},
		{
			{Name: "Intel Core i5-13600T", Category: "cpu"},
			{Name: "Gigabyte H610M K", Category: "motherboard"},
			{Name: "Kingston Fury 32GB DDR4-3600", Category: "ram"},
			{Name: "Kingston NV2 500GB", Category: "ssd"},
			{Name: "Seagate IronWolf 4TB", Category: "hdd"},
			{Name: "SilverStone ST30SF V2.0", Category: "psu"},
			{Name: "Fractal Design Node 202", Category: "case"},
		},
	},
}

// GetUseCaseBuild возвращает список сборок по сценарию использования
func (s *configService) GetUseCaseBuild(usecase string, limit int) ([]domain.NamedBuild, error) {
	builds, ok := PredefinedBuilds[usecase]
	if !ok {
		return nil, fmt.Errorf("unknown use-case %q", usecase)
	}

	out := make([]domain.NamedBuild, 0, limit)

	for rank, items := range builds {
		if len(out) >= limit {
			break
		}

		var comps []domain.Component
		for _, item := range items {
			c, err := s.repo.GetComponentByName(item.Category, item.Name)
			if err != nil {
				return nil, fmt.Errorf("component %q not found in DB: %w", item.Name, err)
			}
			comps = append(comps, c)
		}

		out = append(out, domain.NamedBuild{
			Name:       buildLabels[rank],
			Components: comps,
		})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no builds available for %s", usecase)
	}
	return out, nil
}

// helper — найти CPU в комбо
func findCPU(combo []domain.Component) *domain.Component {
	for i := range combo {
		if strings.EqualFold(combo[i].Category, "cpu") {
			return &combo[i]
		}
	}
	return nil
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

func coolerMatches(c domain.Component, rule rules.ScenarioRule) bool {
	specs := parseSpecs(c.Specs)

	// 1) Socket: кулер должен поддерживать один из допустимых сокетов сценария
	sock, _ := specs["socket"].(string)
	if !contains(rule.CPUSocketWhitelist, sock) {
		return false
	}

	// 2) (Опционально) Можно здесь же проверять высоту кулера,
	//     если вы захотите заложить это в правила:
	// if hRaw, ok := specs["height_mm"].(float64); ok {
	//     height := int(hRaw)
	//     if rule.MaxCoolerHeight > 0 && height > rule.MaxCoolerHeight {
	//         return false
	//     }
	// }

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

	// 1) Форм-фактор корпуса
	if cfRaw, ok := specs["form_factor"].(string); ok && cfRaw != "" && len(rule.CaseFormFactors) > 0 {
		if !contains(rule.CaseFormFactors, cfRaw) {
			return false
		}
	}

	// 2) Выясняем, требует ли сценарий только M.2-накопители
	onlyM2 := len(rule.SSDFormFactors) > 0
	for _, ff := range rule.SSDFormFactors {
		if strings.ToUpper(ff) != "M.2" {
			onlyM2 = false
			break
		}
	}
	// Если только M.2 — слотами можно не проверять
	if onlyM2 {
		return true
	}

	// 3) Для каждого не-M.2 SSD проверяем физические слоты
	for _, ff := range rule.SSDFormFactors {
		switch strings.ToLower(ff) {
		case "2.5", `2.5"`:
			v, ok := specs["drive_bays_2_5"].(float64)
			if !ok || int(v) < 1 {
				return false
			}
		case "3.5", `3.5"`:
			v, ok := specs["drive_bays_3_5"].(float64)
			if !ok || int(v) < 1 {
				return false
			}
			// любые другие форм-факторы (например “M.2”) уже пропущены выше
		}
	}

	// 4) Если сценарий требует HDD (по MinHDDCapacity) — требует хотя бы один 3.5″ слот
	if rule.MinHDDCapacity > 0 {
		v, ok := specs["drive_bays_3_5"].(float64)
		if !ok || int(v) < 1 {
			return false
		}
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
	// разметка полей под JSON в Specs
	var specs struct {
		CapacityGB float64 `json:"capacity_gb"`
		Interface  string  `json:"interface"`
	}

	if err := json.Unmarshal(c.Specs, &specs); err != nil {
		return false
	}

	// 1) интерфейс должен быть SATA
	if !strings.EqualFold(specs.Interface, "SATA") {
		return false
	}

	// 2) проверяем, что ёмкость внутри заданного сценарием диапазона
	capGB := int(specs.CapacityGB)
	if capGB < rule.MinHDDCapacity || capGB > rule.MaxHDDCapacity {
		return false
	}

	return true
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

// веса одного сценария
type Weights struct {
	CPU, GPU, RAM, SSD, HDD, PSU float64
}

// глобальная карта «сценарий → веса»

var scenarioWeights = map[string]Weights{
	"office": {
		CPU: 0.8, GPU: 0.2, RAM: 1, SSD: 0.5, HDD: 0.2, PSU: 0.5,
	},
	"htpc": {
		CPU: 1.5, GPU: 0.5, RAM: 1, SSD: 1, HDD: 1, PSU: 0.5,
	},
	"gaming": {
		CPU: 2.5, GPU: 5, RAM: 2, SSD: 2, HDD: 0.2, PSU: 1.2,
	},
	"streamer": {
		CPU: 3.5, GPU: 3.5, RAM: 3, SSD: 2, HDD: 0.2, PSU: 1.5,
	},
	"design": {
		CPU: 2.5, GPU: 4, RAM: 2, SSD: 1.5, HDD: 1, PSU: 1,
	},
	"video": {
		CPU: 3.5, GPU: 4.5, RAM: 4, SSD: 2.5, HDD: 2, PSU: 1.5,
	},
	"cad": {
		CPU: 4, GPU: 3, RAM: 4, SSD: 2, HDD: 0.2, PSU: 1,
	},
	"dev": {
		CPU: 3, GPU: 0.8, RAM: 3, SSD: 2, HDD: 0.5, PSU: 1,
	},
	"enthusiast": {
		CPU: 4, GPU: 5, RAM: 3, SSD: 1.5, HDD: 0.5, PSU: 1.2,
	},
	"nas": {
		CPU: 1, GPU: 0.2, RAM: 2, SSD: 0.8, HDD: 2, PSU: 1,
	},
}

// rankCached не парсит JSON! Работает по мэпам скор-значений.
func rankCached(
	scen string,
	combo []domain.Component,

	cpuScoreMap, gpuScoreMap,
	ramScoreMap,
	ssdScoreMap,
	hddScoreMap,
	psuScoreMap map[int]float64,

	cpuTDPMap, gpuTDPMap map[int]int,
) int {
	w := scenarioWeights[scen]

	// Найти компоненты по категории
	find := func(cat string) *domain.Component {
		for i := range combo {
			if strings.EqualFold(combo[i].Category, cat) {
				return &combo[i]
			}
		}
		return nil
	}

	var score, total float64

	// CPU
	if cpu := find("cpu"); cpu != nil {
		if sc, ok := cpuScoreMap[cpu.ID]; ok {
			score += sc * w.CPU
			total += w.CPU
		}
	}

	// GPU (опционально)
	if gpu := find("gpu"); gpu != nil && gpu.ID != 0 {
		if sc, ok := gpuScoreMap[gpu.ID]; ok {
			score += sc * w.GPU
			total += w.GPU
		}
	}

	// RAM
	if ram := find("ram"); ram != nil {
		if sc, ok := ramScoreMap[ram.ID]; ok {
			score += sc * w.RAM
			total += w.RAM
		}
	}

	// SSD
	if ssd := find("ssd"); ssd != nil {
		if sc, ok := ssdScoreMap[ssd.ID]; ok {
			score += sc * w.SSD
			total += w.SSD
		}
	}

	// HDD
	if w.HDD > 0 {
		if hdd := find("hdd"); hdd != nil && hdd.ID != 0 {
			if sc, ok := hddScoreMap[hdd.ID]; ok {
				score += sc * w.HDD
				total += w.HDD
			}
		}
	}

	// PSU — здесь score уже равен headroom-скор
	if psu := find("psu"); psu != nil {
		if sc, ok := psuScoreMap[psu.ID]; ok {
			score += sc * w.PSU
			total += w.PSU
		}
	}

	if total == 0 {
		return 0
	}
	avg := score / total

	// 0..1 → 0..4
	return int(avg * 5)
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

func ValidateCombo(combo []domain.Component, rule rules.ScenarioRule) bool {
	var (
		cpu, mb, ram, psu, cs, cooler, ssd, hdd, gpu *domain.Component
	)
	for i := range combo {
		switch strings.ToLower(combo[i].Category) {
		case "cpu":
			cpu = &combo[i]
		case "motherboard":
			mb = &combo[i]
		case "ram":
			ram = &combo[i]
		case "psu":
			psu = &combo[i]
		case "case":
			cs = &combo[i]
		case "cooler":
			cooler = &combo[i]
		case "ssd":
			ssd = &combo[i]
		case "hdd":
			hdd = &combo[i]
		case "gpu":
			gpu = &combo[i]
		}
	}
	if cpu != nil && !cpuMatches(*cpu, rule) {
		return false
	}
	if mb != nil && !mbMatches(*mb, rule) {
		return false
	}
	if ram != nil && !ramMatches(*ram, rule) {
		return false
	}
	if psu != nil && !psuMatches(*psu, rule) {
		return false
	}
	if cs != nil && !caseMatches(*cs, rule) {
		return false
	}
	if cooler != nil && !coolerMatches(*cooler, rule) {
		return false
	}
	if ssd != nil && !ssdMatches(*ssd, rule) {
		return false
	}
	if hdd != nil && hdd.ID != 0 && !hddMatches(*hdd, rule) {
		return false
	}
	if gpu != nil && gpu.ID != 0 && !gpuMatches(*gpu, rule) {
		return false
	}
	return true
}
