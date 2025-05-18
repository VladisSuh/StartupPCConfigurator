package repository

import (
	"StartupPCConfigurator/internal/domain"
	"database/sql"
	"fmt"
	_ "fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Repository интерфейс
type ConfigRepository interface {
	GetComponents(category, search, brand string) ([]domain.Component, error)
	GetCompatibleComponents(filter domain.CompatibilityFilter) ([]domain.Component, error)
	CreateConfiguration(userId uuid.UUID, name string, components []domain.Component) (domain.Configuration, error)
	UpdateConfiguration(userId uuid.UUID, configId, name string, comps []domain.Component) (domain.Configuration, error)
	GetUserConfigurations(userId uuid.UUID) ([]domain.Configuration, error)
	DeleteConfiguration(userId uuid.UUID, configId string) error
	GetComponentByID(category, id string) (domain.Component, error)
	GetComponentByName(category, name string) (domain.Component, error)
	GetUseCases() ([]domain.UseCase, error)
	GetBrandsByCategory(category string) ([]string, error)
}

// Реализация
type configRepository struct {
	db *sql.DB
}

func (r *configRepository) GetCompatibleComponents(filter domain.CompatibilityFilter) ([]domain.Component, error) {
	// 1) Список «разрешённых» полей для каждой целевой категории
	allowedByCategory := map[string][]string{
		"motherboard": {"socket", "ram_type", "form_factor", "max_memory_gb", "pcie_version", "m2_slots", "sata_ports"},
		"case":        {"form_factor", "gpu_max_length", "cooler_max_height", "max_psu_length", "psu_form_factor"},
		"psu":         {"power", "power_required", "modular", "efficiency", "form_factor"},
		"gpu":         {"interface", "power_draw", "length_mm", "height_mm"},
		"ram":         {"ram_type", "capacity", "frequency", "modules", "voltage"},
		"ssd":         {"interface", "form_factor", "capacity_gb", "m2_key", "max_throughput"},
		"hdd":         {"interface", "form_factor", "capacity_gb", "rpm"},
	}

	allowedKeys := allowedByCategory[strings.ToLower(filter.Category)]
	allowedSet := make(map[string]bool, len(allowedKeys))
	for _, k := range allowedKeys {
		allowedSet[k] = true
	}

	// 2) Базовый запрос по целевой категории
	query := `
        SELECT id, name, category, brand, specs, created_at, updated_at
          FROM components
         WHERE LOWER(category) = LOWER($1)
	`

	args := []interface{}{filter.Category}
	idx := 2

	// 3) Добавляем условия только по полям, которые есть в allowedSet
	for key, val := range filter.Specs {
		if !allowedSet[key] {
			continue
		}
		switch v := val.(type) {
		case string:
			query += fmt.Sprintf(" AND LOWER(specs->>'%s') = LOWER($%d)", key, idx)
			args = append(args, v)
			idx++
		case float64:
			query += fmt.Sprintf(" AND (specs->>'%s')::float >= $%d", key, idx)
			args = append(args, v)
			idx++
		}
	}

	// 4) Debug (можно убрать в production)
	fmt.Printf("DEBUG GetCompatibleComponents:\nSQL:\n%s\nARGS: %v\n\n", query, args)

	// 5) Выполняем запрос
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 6) Сканируем результаты в непустой слайс
	result := make([]domain.Component, 0)
	for rows.Next() {
		var c domain.Component
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Category, &c.Brand,
			&c.Specs, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, c)
	}

	return result, nil
}

func NewConfigRepository(db *sql.DB) ConfigRepository {
	return &configRepository{db: db}
}

// Пример метода GetComponents
// Реализация метода GetComponents
// старый метод замените на:
func (r *configRepository) GetComponents(category, search, brand string) ([]domain.Component, error) {
	var (
		args  []interface{}
		where []string
		idx   = 1
	)
	if category != "" {
		where = append(where, fmt.Sprintf("category = $%d", idx))
		args = append(args, category)
		idx++
	}
	if search != "" {
		where = append(where, fmt.Sprintf("LOWER(name) ILIKE LOWER($%d)", idx))
		args = append(args, "%"+search+"%")
		idx++
	}
	if brand != "" {
		where = append(where, fmt.Sprintf("LOWER(brand) = LOWER($%d)", idx))
		args = append(args, brand)
		idx++
	}

	q := "SELECT id, name, category, brand, specs FROM components"
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Component
	for rows.Next() {
		var c domain.Component
		if err := rows.Scan(&c.ID, &c.Name, &c.Category, &c.Brand, &c.Specs); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *configRepository) CreateConfiguration(
	userId uuid.UUID, name string,
	components []domain.Component,
) (domain.Configuration, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return domain.Configuration{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	var configID int
	var createdAt, updatedAt time.Time
	insertConfig := `
		INSERT INTO configurations (user_id, name, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(insertConfig, userId, name).Scan(&configID, &createdAt, &updatedAt)
	if err != nil {
		return domain.Configuration{}, err
	}

	insertComp := `
		INSERT INTO configuration_components (config_id, component_id, category, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	for _, comp := range components {
		_, err = tx.Exec(insertComp, configID, comp.ID, comp.Category)
		if err != nil {
			return domain.Configuration{}, err
		}
	}

	refs := make([]domain.ComponentRef, 0, len(components))
	for _, c := range components {
		refs = append(refs, domain.ComponentRef{
			Category: c.Category,
			Name:     c.Name,
		})
	}

	return domain.Configuration{
		ID:         configID,
		Name:       name,
		OwnerID:    userId,
		Components: refs,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func (r *configRepository) UpdateConfiguration(
	userId uuid.UUID,
	configId, name string,
	comps []domain.Component,
) (domain.Configuration, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return domain.Configuration{}, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var existing domain.Configuration
	queryCheck := `
		SELECT id, user_id, name, created_at, updated_at
		FROM configurations
		WHERE id = $1
	`
	err = tx.QueryRow(queryCheck, configId).Scan(
		&existing.ID,
		&existing.OwnerID,
		&existing.Name,
		&existing.CreatedAt,
		&existing.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Configuration{}, domain.ErrConfigNotFound
	} else if err != nil {
		return domain.Configuration{}, err
	}

	if existing.OwnerID != userId {
		return domain.Configuration{}, domain.ErrForbidden
	}

	queryUpdate := `
		UPDATE configurations
		SET name = $1,
		    updated_at = NOW()
		WHERE id = $2
	`
	_, err = tx.Exec(queryUpdate, name, configId)
	if err != nil {
		return domain.Configuration{}, err
	}

	queryDelComps := `
		DELETE FROM configuration_components
		WHERE config_id = $1
	`
	_, err = tx.Exec(queryDelComps, configId)
	if err != nil {
		return domain.Configuration{}, err
	}

	queryInsertComp := `
		INSERT INTO configuration_components
			(config_id, component_id, category, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	for _, c := range comps {
		_, err = tx.Exec(queryInsertComp, configId, c.ID, c.Category)
		if err != nil {
			return domain.Configuration{}, err
		}
	}

	// обновим время конфигурации
	err = tx.QueryRow(queryCheck, configId).Scan(
		&existing.ID,
		&existing.OwnerID,
		&existing.Name,
		&existing.CreatedAt,
		&existing.UpdatedAt,
	)
	if err != nil {
		return domain.Configuration{}, err
	}

	refs := make([]domain.ComponentRef, 0, len(comps))
	for _, c := range comps {
		refs = append(refs, domain.ComponentRef{
			Category: c.Category,
			Name:     c.Name,
		})
	}

	updatedConfig := domain.Configuration{
		ID:         existing.ID,
		OwnerID:    existing.OwnerID,
		Name:       existing.Name,
		CreatedAt:  existing.CreatedAt,
		UpdatedAt:  existing.UpdatedAt,
		Components: refs,
	}

	return updatedConfig, nil
}

// И т.д. для CreateConfiguration, Update, DeleteConfiguration...
func (r *configRepository) GetUserConfigurations(userId uuid.UUID) ([]domain.Configuration, error) {
	queryConfigs := `
		SELECT id, name, created_at, updated_at
		FROM configurations
		WHERE user_id = $1
	`
	rows, err := r.db.Query(queryConfigs, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []domain.Configuration
	for rows.Next() {
		var cfg domain.Configuration
		cfg.OwnerID = userId
		if err := rows.Scan(&cfg.ID, &cfg.Name, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			return nil, err
		}

		compQuery := `
		SELECT c.name, c.category
		FROM configuration_components cc
		JOIN components c ON cc.component_id = c.id
		WHERE cc.config_id = $1
	`
		compRows, err := r.db.Query(compQuery, cfg.ID)
		if err != nil {
			return nil, err
		}
		for compRows.Next() {
			var ref domain.ComponentRef
			if err := compRows.Scan(&ref.Name, &ref.Category); err != nil {
				compRows.Close()
				return nil, err
			}
			cfg.Components = append(cfg.Components, ref)
		}
		compRows.Close()

		configs = append(configs, cfg)
	}

	return configs, nil
}

func (r *configRepository) DeleteConfiguration(userId uuid.UUID, configId string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Проверяем, что конфигурация существует и принадлежит пользователю
	var owner uuid.UUID
	checkQuery := `SELECT user_id FROM configurations WHERE id = $1`
	err = tx.QueryRow(checkQuery, configId).Scan(&owner)
	if err == sql.ErrNoRows {
		return domain.ErrConfigNotFound
	} else if err != nil {
		return err
	}
	if owner != userId {
		return domain.ErrForbidden
	}

	// Удаляем компоненты
	delComponents := `DELETE FROM configuration_components WHERE config_id = $1`
	if _, err = tx.Exec(delComponents, configId); err != nil {
		return err
	}

	// Удаляем саму конфигурацию
	delConfig := `DELETE FROM configurations WHERE id = $1`
	if _, err = tx.Exec(delConfig, configId); err != nil {
		return err
	}

	return nil
}

// GetComponentByID извлекает компонент по категории и ID
func (r *configRepository) GetComponentByID(category, id string) (domain.Component, error) {
	query := `
		SELECT id, name, category, brand, specs, created_at, updated_at
		FROM components
		WHERE id = $1 AND category = $2
	`

	// предполагаем, что id — это int (если uuid — адаптировать)
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.Component{}, fmt.Errorf("invalid component ID: %s", id)
	}

	var c domain.Component
	err = r.db.QueryRow(query, idInt, category).Scan(
		&c.ID,
		&c.Name,
		&c.Category,
		&c.Brand,
		&c.Specs,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return domain.Component{}, err
	}

	return c, nil
}

func (r *configRepository) GetComponentByName(category, name string) (domain.Component, error) {
	query := `
		SELECT id, name, category, brand, specs, created_at, updated_at
		FROM components
		WHERE LOWER(category) = LOWER($1) AND LOWER(name) = LOWER($2)
	`

	var c domain.Component
	err := r.db.QueryRow(query, category, name).Scan(
		&c.ID,
		&c.Name,
		&c.Category,
		&c.Brand,
		&c.Specs,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return domain.Component{}, fmt.Errorf("component not found: %s / %s", category, name)
	}

	return c, nil
}

// Реализуем GetUseCases
func (r *configRepository) GetUseCases() ([]domain.UseCase, error) {
	rows, err := r.db.Query(`SELECT id, name, description FROM usecases`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.UseCase
	for rows.Next() {
		var uc domain.UseCase
		if err := rows.Scan(&uc.ID, &uc.Name, &uc.Description); err != nil {
			return nil, err
		}
		out = append(out, uc)
	}
	return out, nil
}

// internal/config/repository/pg/repository.go
func (r *configRepository) GetBrandsByCategory(cat string) ([]string, error) {
	const q = `SELECT DISTINCT brand
               FROM components
               WHERE category = $1
               ORDER BY brand`
	rows, err := r.db.Query(q, cat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []string
	for rows.Next() {
		var b string
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}
		brands = append(brands, b)
	}
	return brands, nil
}
