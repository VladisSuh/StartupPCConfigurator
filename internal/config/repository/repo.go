package repository

import (
	"StartupPCConfigurator/internal/domain"
	"database/sql"
	"fmt"
	_ "fmt"
	"strconv"
	"time"
)

// Repository интерфейс
type ConfigRepository interface {
	GetComponents(category, search string) ([]domain.Component, error)
	GetCompatibleComponents(category, cpuSocket, memoryType string) ([]domain.Component, error)
	CreateConfiguration(userId, name string, components []domain.ComponentRef) (domain.Configuration, error)
	UpdateConfiguration(userId, configId, name string, comps []domain.ComponentRef) (domain.Configuration, error)
	GetUserConfigurations(userId string) ([]domain.Configuration, error)
	DeleteConfiguration(userId, configId string) error
}

// Реализация
type configRepository struct {
	db *sql.DB
}

func (r *configRepository) GetCompatibleComponents(category, cpuSocket, memoryType string) ([]domain.Component, error) {
	// Начинаем с основного SQL-запроса для выборки компонентов по категории.
	query := `
        SELECT id, name, category, brand, specs, created_at, updated_at
        FROM components
        WHERE category = $1
    `
	args := []interface{}{category}

	// Если задан cpuSocket — добавляем условие для материнских плат,
	// предполагаем, что в поле specs хранится информация в JSON (например, {"socket": "LGA1200", ...})
	// или отдельное поле, зависящее от реализации.
	if cpuSocket != "" {
		query += ` AND specs->>'socket' = $2`
		args = append(args, cpuSocket)
	}

	// Если задан memoryType — добавляем условие для оперативной памяти или материнских плат,
	// предполагая, что, например, specs->>'memoryType' хранит этот тип
	if memoryType != "" {
		if cpuSocket != "" {
			// $3 если cpuSocket уже задан
			query += ` AND specs->>'memoryType' = $3`
		} else {
			query += ` AND specs->>'memoryType' = $2`
		}
		args = append(args, memoryType)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Component
	for rows.Next() {
		var comp domain.Component
		if err := rows.Scan(
			&comp.ID,
			&comp.Name,
			&comp.Category,
			&comp.Brand,
			&comp.Specs,
			&comp.CreatedAt,
			&comp.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, comp)
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
	var args []interface{}

	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}
	if search != "" {
		query += ` AND (LOWER(name) LIKE ? OR LOWER(brand) LIKE ?)`
		searchLike := "%" + search + "%"
		args = append(args, searchLike, searchLike)
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

// Пример метода CreateConfiguration
func (r *configRepository) CreateConfiguration(
	userId, name string,
	components []domain.ComponentRef,
) (domain.Configuration, error) {

	// Начинаем транзакцию, чтобы при ошибке можно было откатить
	tx, err := r.db.Begin()
	if err != nil {
		return domain.Configuration{}, err
	}

	// Если в ходе метода произойдёт ошибка, откатим транзакцию
	// Иначе — зафиксируем (commit).
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// 1) Вставляем в "configurations"
	var configID int
	var createdAt, updatedAt time.Time

	// Для PostgreSQL используем $1, $2 и т.п.
	insertConfig := `
        INSERT INTO configurations (user_id, name, created_at, updated_at)
        VALUES ($1, $2, NOW(), NOW())
        RETURNING id, created_at, updated_at
    `
	err = tx.QueryRow(insertConfig, userId, name).
		Scan(&configID, &createdAt, &updatedAt)
	if err != nil {
		return domain.Configuration{}, err
	}

	// 2) Для каждого элемента в components делаем INSERT в "configuration_components"
	insertComp := `
        INSERT INTO configuration_components (config_id, component_id, category, created_at)
        VALUES ($1, $2, $3, NOW())
    `
	for _, comp := range components {
		// Допустим, componentId должен быть числом (int) в БД
		// Парсим из строки:
		componentInt, convErr := strconv.Atoi(comp.ComponentID)
		if convErr != nil {
			return domain.Configuration{}, fmt.Errorf("invalid componentId: %s", comp.ComponentID)
		}

		_, err = tx.Exec(insertComp, configID, componentInt, comp.Category)
		if err != nil {
			return domain.Configuration{}, err
		}
	}

	// 3) Формируем результат, заполняем структуру domain.Configuration
	// ID мы возвращаем как строку (с учётом, что в структуре field ID — строка)
	cfg := domain.Configuration{
		ID:         configID,
		Name:       name,
		OwnerID:    userId,     // userId из аргумента
		Components: components, // вернём те же компоненты
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	return cfg, nil
}

func (r *configRepository) UpdateConfiguration(
	userId, configId, name string,
	comps []domain.ComponentRef,
) (domain.Configuration, error) {

	// Начинаем транзакцию, чтобы всё обновление было атомарным
	tx, err := r.db.Begin()
	if err != nil {
		return domain.Configuration{}, err
	}
	// Чтобы при ошибке сделать rollback, а при успехе — commit
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 1. Проверяем, что конфигурация существует и принадлежит userId
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
		// Нет такой конфигурации
		return domain.Configuration{}, domain.ErrConfigNotFound
	} else if err != nil {
		return domain.Configuration{}, err
	}
	// Проверяем владельца
	if existing.OwnerID != userId {
		return domain.Configuration{}, domain.ErrForbidden
	}

	// 2. Обновляем саму конфигурацию (меняем name, ставим updated_at)
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

	// 3. Удаляем старые записи из связующей таблицы configuration_components
	//    (предполагаем, что логика "полностью заменить список компонентов")
	queryDelComps := `
        DELETE FROM configuration_components
        WHERE config_id = $1
    `
	_, err = tx.Exec(queryDelComps, configId)
	if err != nil {
		return domain.Configuration{}, err
	}

	// 4. Вставляем новые компоненты в configuration_components
	queryInsertComp := `
        INSERT INTO configuration_components
            (config_id, component_id, category, created_at)
        VALUES ($1, $2, $3, NOW())
    `
	for _, c := range comps {
		_, err = tx.Exec(queryInsertComp, configId, c.ComponentID, c.Category)
		if err != nil {
			return domain.Configuration{}, err
		}
	}

	// 5. Соберём обновлённую структуру, включая Components
	//    Сначала перечитаем саму конфигурацию (чтобы узнать обновлённый updated_at)
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

	// Далее считаем новые компоненты
	queryGetComps := `
        SELECT component_id, category
        FROM configuration_components
        WHERE config_id = $1
    `
	rows, err := tx.Query(queryGetComps, configId)
	if err != nil {
		return domain.Configuration{}, err
	}
	defer rows.Close()

	var updatedComps []domain.ComponentRef
	for rows.Next() {
		var ref domain.ComponentRef
		if err = rows.Scan(&ref.ComponentID, &ref.Category); err != nil {
			return domain.Configuration{}, err
		}
		updatedComps = append(updatedComps, ref)
	}

	// Формируем итоговую конфигурацию
	updatedConfig := domain.Configuration{
		ID:         existing.ID,
		OwnerID:    existing.OwnerID,
		Name:       existing.Name,
		CreatedAt:  existing.CreatedAt,
		UpdatedAt:  existing.UpdatedAt,
		Components: updatedComps,
	}

	return updatedConfig, nil
}

// И т.д. для CreateConfiguration, Update, DeleteConfiguration...
func (r *configRepository) GetUserConfigurations(userId string) ([]domain.Configuration, error) {
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

		// Подтягиваем компоненты для этой конфигурации
		compQuery := `
            SELECT component_id, category
            FROM configuration_components
            WHERE config_id = $1
        `
		compRows, err := r.db.Query(compQuery, cfg.ID)
		if err != nil {
			return nil, err
		}

		for compRows.Next() {
			var ref domain.ComponentRef
			if err := compRows.Scan(&ref.ComponentID, &ref.Category); err != nil {
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

func (r *configRepository) DeleteConfiguration(userId, configId string) error {
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
	var owner string
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
