package repository

import (
	"StartupPCConfigurator/internal/aggregator/usecase"
	"context"
	"database/sql"

	_ "github.com/lib/pq" // драйвер PostgreSQL

	"StartupPCConfigurator/internal/domain"
)

// repoImpl — один единственный репозиторий, реализующий оба набора методов
type repoImpl struct {
	db *sql.DB
}

// NewRepository открывает соединение и возвращает *repoImpl
func NewRepository(connStr string) (*repoImpl, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &repoImpl{db: db}, nil
}

// ----------------------
// Методы для HTTP‑usecase
// ----------------------

// FetchOffers реализует usecase.OffersRepository и возвращает список офферов
func (r *repoImpl) FetchOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error) {
	const q = `
SELECT
    o.component_id,
    o.shop_id,
    s.code   AS shop_code,
    s.name   AS shop_name,
    o.price,
    o.currency,
    o.availability,
    o.url,
    o.fetched_at
FROM offers o
JOIN shops s ON s.id = o.shop_id
WHERE o.component_id = $1
`
	// Добавляем сортировку
	query := q
	switch filter.Sort {
	case "priceAsc":
		query += " ORDER BY o.price ASC"
	case "priceDesc":
		query += " ORDER BY o.price DESC"
	default:
		query += " ORDER BY s.name"
	}

	rows, err := r.db.QueryContext(ctx, query, filter.ComponentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Offer
	for rows.Next() {
		var of domain.Offer
		of.ComponentID = filter.ComponentID
		if err := rows.Scan(
			&of.ComponentID,
			&of.ShopID,
			&of.ShopCode,
			&of.ShopName,
			&of.Price,
			&of.Currency,
			&of.Availability,
			&of.URL,
			&of.FetchedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, of)
	}
	return out, rows.Err()
}

// ---------------------------
// Методы для Update‑UseCase
// ---------------------------

// ShopComponent — пара component_id + URL страницы
// нужен для usecase.Repository
type ShopComponent struct {
	ComponentID string
	URL         string
}

// ListShopComponents возвращает все componentID+URL для shopID
func (r *repoImpl) ListShopComponents(ctx context.Context, shopID int64) ([]usecase.ShopComponent, error) {
	const q = `
SELECT component_id, url
FROM shop_components
WHERE shop_id = $1
`
	rows, err := r.db.QueryContext(ctx, q, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []usecase.ShopComponent
	for rows.Next() {
		var sc usecase.ShopComponent
		if err := rows.Scan(&sc.ComponentID, &sc.URL); err != nil {
			return nil, err
		}
		res = append(res, sc)
	}
	return res, rows.Err()
}

// UpsertOffer вставляет или обновляет запись в offers
func (r *repoImpl) UpsertOffer(ctx context.Context,
	compID string, shopID int64,
	price float64, availability, url string,
) error {
	const q = `
INSERT INTO offers
  (component_id, shop_id, price, currency, availability, url, fetched_at)
VALUES ($1,$2,$3,$4,$5,$6,NOW())
ON CONFLICT (component_id, shop_id) DO UPDATE
  SET price        = EXCLUDED.price,
      currency     = EXCLUDED.currency,
      availability = EXCLUDED.availability,
      url          = EXCLUDED.url,
      fetched_at   = EXCLUDED.fetched_at
`
	_, err := r.db.ExecContext(ctx, q,
		compID, shopID, price, "USD", availability, url,
	)
	return err
}

// InsertPriceHistory пишет запись в price_history
func (r *repoImpl) InsertPriceHistory(ctx context.Context,
	compID string, shopID int64, price float64,
) error {
	const q = `
INSERT INTO price_history
  (component_id, shop_id, price, currency, captured_at)
VALUES ($1,$2,$3,$4,NOW())
`
	_, err := r.db.ExecContext(ctx, q, compID, shopID, price, "USD")
	return err
}

// UpdateJobStatus обновляет статус задачи в update_jobs
func (r *repoImpl) UpdateJobStatus(ctx context.Context,
	jobID int64, status string, message interface{},
) error {
	var q string
	var args []interface{}

	if status == "running" {
		q = `
UPDATE update_jobs
   SET status     = $2,
       started_at = NOW()
 WHERE id = $1
`
		args = []interface{}{jobID, status}
	} else {
		q = `
UPDATE update_jobs
   SET status      = $2,
       finished_at = NOW(),
       message     = $3
 WHERE id = $1
`
		args = []interface{}{jobID, status, message}
	}

	_, err := r.db.ExecContext(ctx, q, args...)
	return err
}

// record — структура, совпадающая с тем, что передаёт usecase
func (r *repoImpl) BulkUpsertOffers(
	ctx context.Context,
	recs []usecase.ImportRecord, // надо тип record вынести в usecase
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO offers
  (component_id, shop_id, price, currency, availability, url, fetched_at)
VALUES
  ($1, (SELECT id FROM shops WHERE code = $2), $3, $4, $5, $6, NOW())
ON CONFLICT (component_id, shop_id) DO UPDATE
  SET price = EXCLUDED.price,
      currency = EXCLUDED.currency,
      availability = EXCLUDED.availability,
      url = EXCLUDED.url,
      fetched_at = EXCLUDED.fetched_at
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range recs {
		if _, err := stmt.ExecContext(
			ctx,
			r.ComponentID,
			r.ShopCode,
			r.Price,
			r.Currency,
			r.Availability,
			r.URL,
		); err != nil {
			return err
		}
		// можно тут же писать и в price_history, если нужно
	}

	return tx.Commit()
}
