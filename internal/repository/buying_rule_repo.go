package repository

import (
	"context"
	"database/sql"
)

type BuyingRuleRepository struct {
	db *sql.DB
}

func NewBuyingRuleRepository(db *sql.DB) *BuyingRuleRepository {
	return &BuyingRuleRepository{db: db}
}

func (r *BuyingRuleRepository) ListByPortfolio(ctx context.Context, portfolioID string) ([]BuyingRule, error) {
	query := `
		SELECT id, portfolio_id, isin, ticker, security_name, target_pct, created_at, updated_at
		FROM buying_rule
		WHERE portfolio_id = $1
		ORDER BY target_pct DESC
	`
	rows, err := r.db.QueryContext(ctx, query, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []BuyingRule
	for rows.Next() {
		var rule BuyingRule
		if err := rows.Scan(
			&rule.ID, &rule.PortfolioID, &rule.ISIN, &rule.Ticker,
			&rule.SecurityName, &rule.TargetPct, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *BuyingRuleRepository) Upsert(ctx context.Context, rule BuyingRule) error {
	query := `
		INSERT INTO buying_rule (portfolio_id, isin, ticker, security_name, target_pct)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (portfolio_id, isin) DO UPDATE
		  SET ticker = EXCLUDED.ticker,
		      security_name = EXCLUDED.security_name,
		      target_pct = EXCLUDED.target_pct,
		      updated_at = now()
	`
	_, err := r.db.ExecContext(ctx, query,
		rule.PortfolioID, rule.ISIN, rule.Ticker, rule.SecurityName, rule.TargetPct)
	return err
}
