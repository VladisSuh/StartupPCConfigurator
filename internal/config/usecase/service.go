package usecase

import (
	"StartupPCConfigurator/internal/config/repository"
	"StartupPCConfigurator/internal/config/usecase/rules"
	"StartupPCConfigurator/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
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
		if mbSpec, ok := specsByCat["motherboard"]; ok {
			if ff, ok2 := mbSpec["form_factor"]; ok2 {
				specC["form_factor"] = ff
			}
		}
		if cpuSpec, ok := specsByCat["cpu"]; ok {
			if h, ok2 := cpuSpec["cooler_height"]; ok2 {
				specC["cooler_max_height"] = h
			}
		}
		if gpuSpec, ok := specsByCat["gpu"]; ok {
			if l, ok2 := gpuSpec["length_mm"]; ok2 {
				specC["gpu_max_length"] = l
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
func (s *configService) GetUseCaseBuild(usecaseName string, limit int) ([]domain.NamedBuild, error) {
	rule, ok := rules.ScenarioRules[usecaseName]
	if !ok {
		return nil, fmt.Errorf("unknown use case %q", usecaseName)
	}
	w := scenarioWeights[usecaseName]
	ctx := context.Background()
	start := time.Now()

	var (
		cpusRaw, mbsRaw, ramsRaw, psusRaw, casesRaw,
		gpusRaw, ssdsRaw, hddsRaw, coolersRaw []domain.Component
	)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		cpusRaw, err = s.repo.GetComponentsFiltered(ctx, repository.ComponentFilter{
			Categories:      []string{"cpu"},
			SocketWhitelist: rule.CPUSocketWhitelist,
			MinTDP:          rule.MinCPUTDP,
			MaxTDP:          rule.MaxCPUTDP,
		})
		return err
	})
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
	log.Printf("GetUseCaseBuild: fetch all components took %v", time.Since(start))

	// 2) Разбираем specs один раз и считаем score
	type cpuInfo struct {
		comp      domain.Component
		socket    string
		tdp       int
		perfScore float64
	}
	type mbInfo struct {
		comp       domain.Component
		socket     string
		formFactor string
	}
	type ramInfo struct {
		comp  domain.Component
		score float64
	}
	type psuInfo struct {
		comp  domain.Component
		score float64
	}
	type caseInfo struct {
		comp            domain.Component
		formFactor      string
		coolerMaxHeight int
		score           float64
	}
	type gpuInfo struct {
		comp  domain.Component
		score float64
	}
	type ssdInfo struct {
		comp  domain.Component
		score float64
	}
	type hddInfo struct {
		comp  domain.Component
		score float64
	}
	type coolerInfo struct {
		comp   domain.Component
		socket string
		height int
		score  float64
	}

	var (
		cpus    = make([]cpuInfo, 0, len(cpusRaw))
		mbs     = make([]mbInfo, 0, len(mbsRaw))
		rams    = make([]ramInfo, 0, len(ramsRaw))
		gpus    = make([]gpuInfo, 0, len(gpusRaw))
		psus    = make([]psuInfo, 0, len(psusRaw))
		cases   = make([]caseInfo, 0, len(casesRaw))
		ssds    = make([]ssdInfo, 0, len(ssdsRaw))
		hdds    = make([]hddInfo, 0, len(hddsRaw))
		coolers = make([]coolerInfo, 0, len(coolersRaw))
	)

	var wg sync.WaitGroup
	wg.Add(8)

	// 1) CPU
	go func() {
		defer wg.Done()
		for _, c := range cpusRaw {
			spec := parseSpecs(c.Specs)
			perf := toInt(spec["cores"])*2 + toInt(spec["threads"])
			score := normalize(perf, 8, 48) * w.CPU
			cpus = append(cpus, cpuInfo{comp: c, socket: spec["socket"].(string),
				tdp: toInt(spec["tdp"]), perfScore: score})
		}
	}()

	// 2) MB
	go func() {
		defer wg.Done()
		for _, m := range mbsRaw {
			spec := parseSpecs(m.Specs)
			mbs = append(mbs, mbInfo{comp: m, socket: spec["socket"].(string),
				formFactor: spec["form_factor"].(string)})
		}
	}()
	// 3) RAM
	go func() {
		defer wg.Done()
		for _, r := range ramsRaw {
			spec := parseSpecs(r.Specs)
			cap := toInt(spec["capacity"])
			score := normalize(cap, rule.MinRAM, rule.MaxRAM) * w.RAM
			rams = append(rams, ramInfo{comp: r, score: score})
		}
	}()
	// 5) PSU
	go func() {
		defer wg.Done()
		for _, p := range psusRaw {
			spec := parseSpecs(p.Specs)
			power := toInt(spec["power"])
			score := normalize(power, rule.MinPSUPower, rule.MaxPSUPower) * w.PSU
			psus = append(psus, psuInfo{comp: p, score: score})
		}
	}()
	// 6) Case
	go func() {
		defer wg.Done()
		for _, c := range casesRaw {
			spec := parseSpecs(c.Specs)
			bays := toInt(spec["drive_bays_2_5"]) + toInt(spec["drive_bays_3_5"])
			score := float64(bays) * w.SSD
			cases = append(cases, caseInfo{comp: c, formFactor: spec["form_factor"].(string),
				coolerMaxHeight: toInt(spec["cooler_max_height"]), score: score})
		}
	}()

	// 4) GPU
	go func() {
		defer wg.Done()
		for _, g := range gpusRaw {
			spec := parseSpecs(g.Specs)
			mem := toInt(spec["memory_gb"])
			score := normalize(mem, rule.MinGPUMemory, rule.MaxGPUMemory) * w.GPU
			gpus = append(gpus, gpuInfo{comp: g, score: score})
		}
	}()

	// 7) SSD & HDD
	go func() {
		defer wg.Done()
		for _, s := range ssdsRaw {
			spec := parseSpecs(s.Specs)
			tp := toInt(spec["max_throughput"])
			score := normalize(tp, rule.MinSSDThroughput, 8000) * w.SSD
			ssds = append(ssds, ssdInfo{comp: s, score: score})
		}
		for _, h := range hddsRaw {
			spec := parseSpecs(h.Specs)
			cap := toInt(spec["capacity_gb"])
			score := normalize(cap, rule.MinHDDCapacity, rule.MaxHDDCapacity) * w.HDD
			hdds = append(hdds, hddInfo{comp: h, score: score})
		}
	}()
	// 8) Cooler
	go func() {
		defer wg.Done()
		for _, c := range coolersRaw {
			spec := parseSpecs(c.Specs)
			h := toInt(spec["height_mm"])
			score := normalize(200-h, 0, 200) * w.PSU
			coolers = append(coolers, coolerInfo{comp: c, socket: spec["socket"].(string),
				height: h, score: score})
		}
	}()

	wg.Wait()

	duration := time.Since(start)
	log.Printf("GetUseCaseBuild: name=%s took %s", usecaseName, duration)

	// 3) Собираем быстрые мапы совместимости и пуллы
	mbsBySocket := make(map[string][]mbInfo)
	for _, mb := range mbs {
		mbsBySocket[mb.socket] = append(mbsBySocket[mb.socket], mb)
	}
	casesByForm := make(map[string][]caseInfo)
	for _, cs := range cases {
		casesByForm[cs.formFactor] = append(casesByForm[cs.formFactor], cs)
	}
	coolerBySocket := make(map[string][]coolerInfo)
	for _, cl := range coolers {
		coolerBySocket[cl.socket] = append(coolerBySocket[cl.socket], cl)
	}

	// ────────────────────────────────────────────────────────────────
	// 4) Строим мапы скор- и TDP-значений один раз для rankCached
	// ────────────────────────────────────────────────────────────────
	cpuScoreMap := make(map[int]float64, len(cpus))
	cpuTDPMap := make(map[int]int, len(cpus))
	for _, ci := range cpus {
		cpuScoreMap[ci.comp.ID] = ci.perfScore
		cpuTDPMap[ci.comp.ID] = ci.tdp
	}

	gpuScoreMap := make(map[int]float64, len(gpus))
	gpuTDPMap := make(map[int]int, len(gpus))
	for _, gi := range gpus {
		gpuScoreMap[gi.comp.ID] = gi.score
		// если тебе нужен именно power_draw — получай его из gi.comp.Specs
		gpuTDPMap[gi.comp.ID] = toInt(parseSpecs(gi.comp.Specs)["power_draw"])
	}

	ramScoreMap := make(map[int]float64, len(rams))
	for _, ri := range rams {
		ramScoreMap[ri.comp.ID] = ri.score
	}

	ssdScoreMap := make(map[int]float64, len(ssds))
	for _, si := range ssds {
		ssdScoreMap[si.comp.ID] = si.score
	}

	hddScoreMap := make(map[int]float64, len(hdds))
	for _, hi := range hdds {
		hddScoreMap[hi.comp.ID] = hi.score
	}

	psuScoreMap := make(map[int]float64, len(psus))
	for _, pi := range psus {
		psuScoreMap[pi.comp.ID] = pi.score
	}
	// ────────────────────────────────────────────────────────────────

	// Пулы GPU/SSD/HDD (с опциональным «пустым» элементом)
	var gpuPool []domain.Component
	for _, gi := range gpus {
		gpuPool = append(gpuPool, gi.comp)
	}
	if rule.MinGPUMemory == 0 {
		gpuPool = append([]domain.Component{{}}, gpuPool...)
	}
	var ssdPool []domain.Component
	for _, si := range ssds {
		ssdPool = append(ssdPool, si.comp)
	}
	var hddPool []domain.Component
	for _, hi := range hdds {
		hddPool = append(hddPool, hi.comp)
	}
	if rule.MinHDDCapacity == 0 {
		hddPool = append([]domain.Component{{}}, hddPool...)
	}

	type hardCombo struct {
		cpu domain.Component
		mb  domain.Component
		ram domain.Component
		psu domain.Component
	}
	const (
		maxCPU  = 12
		maxMB   = 12
		maxRAM  = 12
		maxPSU  = 8
		maxCase = 8
	)

	numCPU := runtime.NumCPU()
	jobs := make(chan hardCombo, 1024) // буферный канал для производительности
	hardList := make([]hardCombo, 0, maxCPU*maxMB*maxRAM*maxPSU)

	go func() {
		for _, ci := range cpus {
			for _, mb := range mbsBySocket[ci.socket] {
				for _, ri := range rams {
					for _, pi := range psus {
						jobs <- hardCombo{
							cpu: ci.comp,
							mb:  mb.comp,
							ram: ri.comp,
							psu: pi.comp,
						}
					}
				}
			}
		}
		close(jobs)
	}()

	var hlMu sync.Mutex
	wg.Add(numCPU)

	for i := 0; i < numCPU; i++ {
		go func() {
			defer wg.Done()
			for hc := range jobs {
				hlMu.Lock()
				hardList = append(hardList, hc)
				hlMu.Unlock()
			}
		}()
	}

	wg.Wait()

	type midCombo struct {
		hard   hardCombo
		cs     domain.Component
		cooler domain.Component
	}
	// 2) Для каждого «жёсткого» приклеиваем допустимые Case + Cooler
	midList := make([]midCombo, 0, len(hardList)*4) // емпирически
	for _, hc := range hardList {
		socket := parseSpecs(hc.cpu.Specs)["socket"].(string)
		for _, cs := range casesByForm[parseSpecs(hc.mb.Specs)["form_factor"].(string)] {
			for _, cl := range coolerBySocket[socket] {
				if cl.height > cs.coolerMaxHeight {
					continue
				}
				midList = append(midList, midCombo{
					hard:   hc,
					cs:     cs.comp,
					cooler: cl.comp,
				})
			}
		}
	}

	duration2 := time.Since(start)
	log.Printf("GetUseCaseBuild: name=%s took %s", usecaseName, duration2)

	// 3) Параллельный ранкинг
	type comboResult struct {
		rank  int
		combo []domain.Component
	}

	numWorkers := runtime.NumCPU()
	midLen := len(midList)
	chunkSize := (midLen + numWorkers - 1) / numWorkers

	results := make(chan comboResult, limit)
	// защищённые общие структуры
	var (
		mu     sync.Mutex
		seen   = make(map[string]struct{})
		ranked = make(map[int][]domain.Component)
		need   = int32(limit)
	)

	worker := func(start, end int) {
		defer wg.Done()
		for i := start; i < end && atomic.LoadInt32(&need) > 0; i++ {
			mc := midList[i]
			base := []domain.Component{
				mc.hard.cpu, mc.hard.mb, mc.hard.ram,
				mc.hard.psu, mc.cs, mc.cooler,
			}

			for _, ss := range ssdPool {
				if atomic.LoadInt32(&need) <= 0 {
					return
				}
				for _, hd := range hddPool {
					if atomic.LoadInt32(&need) <= 0 {
						return
					}
					for _, gp := range gpuPool {
						if atomic.LoadInt32(&need) <= 0 {
							return
						}

						combo := make([]domain.Component, len(base), len(base)+2)
						copy(combo, base)
						combo = append(combo, ss)
						if hd.ID != 0 {
							combo = append(combo, hd)
						}
						if gp.ID != 0 {
							combo = append(combo, gp)
						}

						key := hashCombo(combo)
						mu.Lock()
						if _, dup := seen[key]; dup {
							mu.Unlock()
							continue
						}
						seen[key] = struct{}{}
						mu.Unlock()

						rank := rankCached(usecaseName, combo,
							cpuScoreMap, gpuScoreMap, ramScoreMap,
							ssdScoreMap, hddScoreMap, psuScoreMap,
							cpuTDPMap, gpuTDPMap,
						)
						if rank >= len(buildLabels) {
							rank = len(buildLabels) - 1
						}

						mu.Lock()
						if _, exists := ranked[rank]; !exists && atomic.LoadInt32(&need) > 0 {
							ranked[rank] = combo
							atomic.AddInt32(&need, -1)
							results <- comboResult{rank: rank, combo: combo}
						}
						mu.Unlock()
					}
				}
			}
		}
	}

	for w := 0; w < numWorkers; w++ {
		start := w * chunkSize
		end := (w + 1) * chunkSize
		if end > midLen {
			end = midLen
		}
		if start < end {
			wg.Add(1)
			go worker(start, end)
		}
	}

	// закрываем канал, когда все воркеры отработают
	go func() {
		wg.Wait()
		close(results)
	}()

	// собираем первые limit сборок
	var out []domain.NamedBuild
	for res := range results {
		out = append(out, domain.NamedBuild{
			Name:       buildLabels[res.rank],
			Components: res.combo,
		})
		if len(out) >= limit {
			break
		}
	}

	if len(out) < limit {
		return nil, fmt.Errorf("не получилось набрать %d сборок, нашлось только %d", limit, len(out))
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
