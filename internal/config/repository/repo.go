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
	GetComponents(category, search string) ([]domain.Component, error)
	GetCompatibleComponents(filter domain.CompatibilityFilter) ([]domain.Component, error)
	CreateConfiguration(userId uuid.UUID, name string, components []domain.Component) (domain.Configuration, error)
	UpdateConfiguration(userId uuid.UUID, configId, name string, comps []domain.Component) (domain.Configuration, error)
	GetUserConfigurations(userId uuid.UUID) ([]domain.Configuration, error)
	DeleteConfiguration(userId uuid.UUID, configId string) error
	GetComponentByID(category, id string) (domain.Component, error)
	GetComponentByName(category, name string) (domain.Component, error)
}

// Реализация
type configRepository struct {
	db *sql.DB
}

func (r *configRepository) GetCompatibleComponents(filter domain.CompatibilityFilter) ([]domain.Component, error) {
	query := `
		SELECT id, name, category, brand, specs, created_at, updated_at
		FROM components
		WHERE LOWER(category) = LOWER($1)
	`
	args := []interface{}{filter.Category}
	index := 2

	if filter.CPUSocket != "" {
		query += fmt.Sprintf(" AND LOWER(specs->>'socket') = LOWER($%d)", index)
		args = append(args, filter.CPUSocket)
		index++
	}

	if filter.RAMType != "" {
		query += fmt.Sprintf(" AND LOWER(specs->>'ram_type') = LOWER($%d)", index)
		args = append(args, filter.RAMType)
		index++
	}

	if filter.FormFactor != "" {
		query += fmt.Sprintf(" AND LOWER(specs->>'form_factor') = LOWER($%d)", index)
		args = append(args, filter.FormFactor)
		index++
	}

	if filter.GPULengthMM > 0 {
		query += fmt.Sprintf(" AND (specs->>'gpu_max_length')::float >= $%d", index)
		args = append(args, filter.GPULengthMM)
		index++
	}

	if filter.CoolerHeightMM > 0 {
		query += fmt.Sprintf(" AND (specs->>'cooler_max_height')::float >= $%d", index)
		args = append(args, filter.CoolerHeightMM)
		index++
	}

	if filter.PowerRequired > 0 {
		query += fmt.Sprintf(" AND (specs->>'power')::float >= $%d", index)
		args = append(args, filter.PowerRequired)
		index++
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Component
	for rows.Next() {
		var c domain.Component
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Category,
			&c.Brand,
			&c.Specs,
			&c.CreatedAt,
			&c.UpdatedAt,
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
func (r *configRepository) GetComponents(category, search string) ([]domain.Component, error) {
	// SQL-запрос; учтите, что если это PostgreSQL — придётся использовать $1, $2 вместо ?
	query := `
    SELECT id, name, category, brand, specs, created_at, updated_at
    FROM components
    WHERE 1=1
`
	var (
		args  []interface{}
		index int = 1 // счётчик placeholder'ов
	)

	if category != "" {
		category = strings.ToLower(category)
		query += fmt.Sprintf(" AND LOWER(category) = $%d", index)
		args = append(args, category)
		index++
	}

	if search != "" {
		search = strings.ToLower(search)
		query += fmt.Sprintf(" AND (LOWER(name) LIKE $%d OR LOWER(brand) LIKE $%d)", index, index+1)
		searchLike := "%" + search + "%"
		args = append(args, searchLike, searchLike)
		index += 2
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Component
	for rows.Next() {
		var c domain.Component
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Category,
			&c.Brand,
			&c.Specs,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
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
