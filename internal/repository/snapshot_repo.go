package repository

import (
	"context"
	"database/sql"

	"github.com/shopspring/decimal"
)

// SnapshotRepository handles snapshot operations
type SnapshotRepository struct {
	db *sql.DB
}

// NewSnapshotRepository creates a new snapshot repository
func NewSnapshotRepository(db *sql.DB) *SnapshotRepository {
	return &SnapshotRepository{db: db}
}

// UpsertValuationSnapshot creates or updates a valuation snapshot
func (r *SnapshotRepository) UpsertValuationSnapshot(ctx context.Context, snap *ValuationSnapshot) error {
	query := `
		INSERT INTO valuation_snapshot (
			account_id, period, total_value, cash_balance, securities_value, currency
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (account_id, period)
		DO UPDATE SET
			total_value = EXCLUDED.total_value,
			cash_balance = EXCLUDED.cash_balance,
			securities_value = EXCLUDED.securities_value,
			currency = EXCLUDED.currency,
			updated_at = now()
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		snap.AccountID, snap.Period, snap.TotalValue, snap.CashBalance,
		snap.SecuritiesValue, snap.Currency,
	).Scan(&snap.ID, &snap.CreatedAt, &snap.UpdatedAt)

	return err
}

// UpsertPositionSnapshot creates or updates a position snapshot
func (r *SnapshotRepository) UpsertPositionSnapshot(ctx context.Context, snap *PositionSnapshot) error {
	query := `
		INSERT INTO position_snapshot (
			account_id, period, isin, security_name, quantity, price, market_value, currency
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (account_id, period, isin)
		DO UPDATE SET
			security_name = EXCLUDED.security_name,
			quantity = EXCLUDED.quantity,
			price = EXCLUDED.price,
			market_value = EXCLUDED.market_value,
			currency = EXCLUDED.currency,
			updated_at = now()
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		snap.AccountID, snap.Period, snap.ISIN, snap.SecurityName,
		snap.Quantity, snap.Price, snap.MarketValue, snap.Currency,
	).Scan(&snap.ID, &snap.CreatedAt, &snap.UpdatedAt)

	return err
}

// GetValuationSnapshot returns valuation snapshot for account and period
func (r *SnapshotRepository) GetValuationSnapshot(ctx context.Context, accountID int64, period string) (*ValuationSnapshot, error) {
	query := `
		SELECT id, account_id, period, total_value, cash_balance, securities_value,
		       currency, created_at, updated_at
		FROM valuation_snapshot
		WHERE account_id = $1 AND period = $2
	`

	var snap ValuationSnapshot
	err := r.db.QueryRowContext(ctx, query, accountID, period).Scan(
		&snap.ID, &snap.AccountID, &snap.Period, &snap.TotalValue,
		&snap.CashBalance, &snap.SecuritiesValue, &snap.Currency,
		&snap.CreatedAt, &snap.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &snap, nil
}

// ListPositionSnapshots returns all positions for account and period
func (r *SnapshotRepository) ListPositionSnapshots(ctx context.Context, accountID int64, period string) ([]PositionSnapshot, error) {
	query := `
		SELECT id, account_id, period, isin, security_name, quantity, price,
		       market_value, currency, created_at, updated_at
		FROM position_snapshot
		WHERE account_id = $1 AND period = $2
		ORDER BY market_value DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []PositionSnapshot
	for rows.Next() {
		var snap PositionSnapshot
		err := rows.Scan(
			&snap.ID, &snap.AccountID, &snap.Period, &snap.ISIN,
			&snap.SecurityName, &snap.Quantity, &snap.Price,
			&snap.MarketValue, &snap.Currency, &snap.CreatedAt, &snap.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snap)
	}

	return snapshots, rows.Err()
}

// ListPeriods returns all periods for an account, sorted ascending
func (r *SnapshotRepository) ListPeriods(ctx context.Context, accountID int64) ([]string, error) {
	query := `
		SELECT DISTINCT period
		FROM valuation_snapshot
		WHERE account_id = $1
		ORDER BY period ASC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var periods []string
	for rows.Next() {
		var period string
		if err := rows.Scan(&period); err != nil {
			return nil, err
		}
		periods = append(periods, period)
	}

	return periods, rows.Err()
}

// GetLastPeriod returns the most recent period for an account
func (r *SnapshotRepository) GetLastPeriod(ctx context.Context, accountID int64) (string, error) {
	query := `
		SELECT period
		FROM valuation_snapshot
		WHERE account_id = $1
		ORDER BY period DESC
		LIMIT 1
	`

	var period string
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(&period)
	return period, err
}

// GetNextPeriod returns the next period after the given one
func (r *SnapshotRepository) GetNextPeriod(ctx context.Context, accountID int64, currentPeriod string) (string, error) {
	query := `
		SELECT period
		FROM valuation_snapshot
		WHERE account_id = $1 AND period > $2
		ORDER BY period ASC
		LIMIT 1
	`

	var period string
	err := r.db.QueryRowContext(ctx, query, accountID, currentPeriod).Scan(&period)
	return period, err
}

// GetPrevPeriod returns the previous period before the given one
func (r *SnapshotRepository) GetPrevPeriod(ctx context.Context, accountID int64, currentPeriod string) (string, error) {
	query := `
		SELECT period
		FROM valuation_snapshot
		WHERE account_id = $1 AND period < $2
		ORDER BY period DESC
		LIMIT 1
	`

	var period string
	err := r.db.QueryRowContext(ctx, query, accountID, currentPeriod).Scan(&period)
	return period, err
}

// GetTotalPositionsValue returns sum of all positions for account and period
func (r *SnapshotRepository) GetTotalPositionsValue(ctx context.Context, accountID int64, period string) (decimal.Decimal, error) {
	query := `
		SELECT COALESCE(SUM(market_value), 0)
		FROM position_snapshot
		WHERE account_id = $1 AND period = $2
	`

	var total decimal.Decimal
	err := r.db.QueryRowContext(ctx, query, accountID, period).Scan(&total)
	return total, err
}
