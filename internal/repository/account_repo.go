package repository

import (
	"context"
	"database/sql"
)

// AccountRepository handles account operations
type AccountRepository struct {
	db *sql.DB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// EnsureAccount creates or returns existing account
func (r *AccountRepository) EnsureAccount(ctx context.Context, name, institution, accountType, currency string) (int64, error) {
	// Try to get existing by name and institution
	var id int64
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM account 
		WHERE name = $1 AND institution = $2
	`, name, institution).Scan(&id)

	if err == nil {
		return id, nil
	}

	if err != sql.ErrNoRows {
		return 0, err
	}

	// Create new
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO account (name, institution, type, currency)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, name, institution, accountType, currency).Scan(&id)

	return id, err
}

// GetAccountByID returns account by ID
func (r *AccountRepository) GetAccountByID(ctx context.Context, id int64) (*Account, error) {
	query := `
		SELECT id, name, institution, type, currency, created_at
		FROM account
		WHERE id = $1
	`

	var acc Account
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&acc.ID, &acc.Name, &acc.Institution, &acc.Type, &acc.Currency, &acc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &acc, nil
}

// ListAccounts returns all accounts
func (r *AccountRepository) ListAccounts(ctx context.Context) ([]Account, error) {
	query := `
		SELECT id, name, institution, type, currency, created_at
		FROM account
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		err := rows.Scan(
			&acc.ID, &acc.Name, &acc.Institution, &acc.Type, &acc.Currency, &acc.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return accounts, rows.Err()
}
