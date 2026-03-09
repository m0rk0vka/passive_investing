package repository

import (
	"context"
	"database/sql"
)

// VirtualPortfolioRepository handles virtual portfolio operations
type VirtualPortfolioRepository struct {
	db *sql.DB
}

// NewVirtualPortfolioRepository creates a new virtual portfolio repository
func NewVirtualPortfolioRepository(db *sql.DB) *VirtualPortfolioRepository {
	return &VirtualPortfolioRepository{db: db}
}

// VirtualPortfolio represents a virtual portfolio
type VirtualPortfolio struct {
	ID        int64
	UserID    int64
	Name      string
	CreatedAt string
	UpdatedAt string
}

// CreateVirtualPortfolio creates a new virtual portfolio
func (r *VirtualPortfolioRepository) CreateVirtualPortfolio(ctx context.Context, userID int64, name string, accountIDs []int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Create virtual portfolio
	var vpID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO virtual_portfolio (id, user_id, name)
		VALUES (nextval('virtual_portfolio_id_seq'), $1, $2)
		RETURNING id
	`, userID, name).Scan(&vpID)
	if err != nil {
		return 0, err
	}

	// Add accounts to virtual portfolio
	for _, accountID := range accountIDs {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO virtual_portfolio_item (id, virtual_portfolio_id, account_id)
			VALUES (nextval('virtual_portfolio_item_id_seq'), $1, $2)
		`, vpID, accountID)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return vpID, nil
}

// GetVirtualPortfolio returns a virtual portfolio by ID
func (r *VirtualPortfolioRepository) GetVirtualPortfolio(ctx context.Context, id, userID int64) (*VirtualPortfolio, error) {
	var vp VirtualPortfolio
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, created_at, updated_at
		FROM virtual_portfolio
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&vp.ID, &vp.UserID, &vp.Name, &vp.CreatedAt, &vp.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &vp, nil
}

// ListVirtualPortfolios returns all virtual portfolios for a user
func (r *VirtualPortfolioRepository) ListVirtualPortfolios(ctx context.Context, userID int64) ([]VirtualPortfolio, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, created_at, updated_at
		FROM virtual_portfolio
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var portfolios []VirtualPortfolio
	for rows.Next() {
		var vp VirtualPortfolio
		if err := rows.Scan(&vp.ID, &vp.UserID, &vp.Name, &vp.CreatedAt, &vp.UpdatedAt); err != nil {
			return nil, err
		}
		portfolios = append(portfolios, vp)
	}

	return portfolios, rows.Err()
}

// GetVirtualPortfolioAccounts returns all account IDs in a virtual portfolio
func (r *VirtualPortfolioRepository) GetVirtualPortfolioAccounts(ctx context.Context, vpID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT account_id
		FROM virtual_portfolio_item
		WHERE virtual_portfolio_id = $1
		ORDER BY created_at ASC
	`, vpID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accountIDs []int64
	for rows.Next() {
		var accountID int64
		if err := rows.Scan(&accountID); err != nil {
			return nil, err
		}
		accountIDs = append(accountIDs, accountID)
	}

	return accountIDs, rows.Err()
}

// AddAccountToVirtualPortfolio adds an account to a virtual portfolio
func (r *VirtualPortfolioRepository) AddAccountToVirtualPortfolio(ctx context.Context, vpID, accountID int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO virtual_portfolio_item (id, virtual_portfolio_id, account_id)
		VALUES (nextval('virtual_portfolio_item_id_seq'), $1, $2)
		ON CONFLICT (virtual_portfolio_id, account_id) DO NOTHING
	`, vpID, accountID)
	return err
}

// RemoveAccountFromVirtualPortfolio removes an account from a virtual portfolio
func (r *VirtualPortfolioRepository) RemoveAccountFromVirtualPortfolio(ctx context.Context, vpID, accountID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM virtual_portfolio_item
		WHERE virtual_portfolio_id = $1 AND account_id = $2
	`, vpID, accountID)
	return err
}

// DeleteVirtualPortfolio deletes a virtual portfolio
func (r *VirtualPortfolioRepository) DeleteVirtualPortfolio(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM virtual_portfolio
		WHERE id = $1 AND user_id = $2
	`, id, userID)
	return err
}
