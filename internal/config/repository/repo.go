package repository

import (
	"StartupPCConfigurator/internal/domain"
	"database/sql"
	_ "fmt"
)

// Repository интерфейс
type ConfigRepository interface {
	GetComponents(category, search string) ([]domain.Component, error)
	CreateConfiguration(userId, name string, components []domain.ComponentRef) (domain.Configuration, error)
	//...
}

// Реализация
type configRepository struct {
	db *sql.DB
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

	// Пока заглушка (stub), чтобы код компилировался
	// В будущем вы сделаете INSERT в таблицу "configurations"
	// + INSERT в "configuration_components" и т. д.
	//
	// _ = userId     // убирает warning "unused parameter"
	// _ = name
	// _ = components

	return domain.Configuration{}, nil
}

// И т.д. для CreateConfiguration, Update, DeleteConfiguration...
