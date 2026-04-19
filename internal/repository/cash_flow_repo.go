package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// CashFlowRepository handles cash flow operations
type CashFlowRepository struct {
	db *sql.DB
}

// NewCashFlowRepository creates a new cash flow repository
func NewCashFlowRepository(db *sql.DB) *CashFlowRepository {
	return &CashFlowRepository{db: db}
}

// ReplaceOperationsForPeriod удаляет все операции для account_id и period, затем вставляет новые
// Это обеспечивает идемпотентность при повторном парсинге файлов
func (r *CashFlowRepository) ReplaceOperationsForPeriod(ctx context.Context, accountID int64, period string, operations []CashFlowOperation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Удаляем все существующие операции для этого периода
	deleteQuery := `DELETE FROM cash_flow_operation WHERE account_id = $1 AND period = $2`
	if _, err := tx.ExecContext(ctx, deleteQuery, accountID, period); err != nil {
		return err
	}

	// Вставляем новые операции
	insertQuery := `
		INSERT INTO cash_flow_operation (
			account_id, period, operation_date, amount, currency, operation_type, comment
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	for i := range operations {
		err := tx.QueryRowContext(
			ctx, insertQuery,
			operations[i].AccountID, operations[i].Period, operations[i].OperationDate,
			operations[i].Amount, operations[i].Currency, operations[i].OperationType,
			operations[i].Comment,
		).Scan(&operations[i].ID, &operations[i].CreatedAt, &operations[i].UpdatedAt)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ListCashFlowOperations returns all cash flow operations for account and period
func (r *CashFlowRepository) ListCashFlowOperations(ctx context.Context, accountID int64, period string) ([]CashFlowOperation, error) {
	query := `
		SELECT id, account_id, period, operation_date, amount, currency,
		       operation_type, comment, created_at, updated_at
		FROM cash_flow_operation
		WHERE account_id = $1 AND period = $2
		ORDER BY operation_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var operations []CashFlowOperation
	for rows.Next() {
		var op CashFlowOperation
		err := rows.Scan(
			&op.ID, &op.AccountID, &op.Period, &op.OperationDate,
			&op.Amount, &op.Currency, &op.OperationType, &op.Comment,
			&op.CreatedAt, &op.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		operations = append(operations, op)
	}

	return operations, rows.Err()
}

// GetDepositsInfo returns aggregated information about deposits for given accounts up to and including the specified period
// Period format: "2025-10" (YYYY-MM)
func (r *CashFlowRepository) GetDepositsInfo(ctx context.Context, accountIDs []int64, upToPeriod string) (*DepositsInfo, error) {
	if len(accountIDs) == 0 {
		return &DepositsInfo{
			TotalAmount: decimal.Zero,
			Currency:    "RUB",
		}, nil
	}

	query := `
		SELECT COALESCE(SUM(amount), 0) as total, COALESCE(MAX(currency), 'RUB') as currency
		FROM cash_flow_operation
		WHERE account_id = ANY($1)
		  AND operation_type = 'DEPOSIT'
		  AND period <= $2
	`

	var info DepositsInfo
	err := r.db.QueryRowContext(ctx, query, accountIDs, upToPeriod).Scan(&info.TotalAmount, &info.Currency)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// ListAllDeposits returns all deposit operations for the given accounts, ordered by date descending.
func (r *CashFlowRepository) ListAllDeposits(ctx context.Context, accountIDs []int64) ([]DepositsInfo, error) {
	if len(accountIDs) == 0 {
		return []DepositsInfo{}, nil
	}

	query := `
		SELECT amount, currency, operation_date
		FROM cash_flow_operation
		WHERE account_id = ANY($1)
		  AND operation_type = 'DEPOSIT'
		ORDER BY operation_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query all deposits: %w", err)
	}
	defer rows.Close()

	var infos []DepositsInfo
	for rows.Next() {
		var info DepositsInfo
		if err := rows.Scan(&info.TotalAmount, &info.Currency, &info.OperationDate); err != nil {
			return nil, fmt.Errorf("failed to scan deposit: %w", err)
		}
		infos = append(infos, info)
	}
	return infos, rows.Err()
}

// GetDepositsInfo returns aggregated information about deposits for given accounts up to and including the specified period
// Period format: "2025-10" (YYYY-MM)
func (r *CashFlowRepository) GetDepositsInfos(
	ctx context.Context, accountIDs []int64, upToPeriod string) ([]DepositsInfo, error) {
	if len(accountIDs) == 0 {
		return []DepositsInfo{}, nil
	}

	query := `
		SELECT amount, currency, operation_date
		FROM cash_flow_operation
		WHERE account_id = ANY($1)
		  AND operation_type = 'DEPOSIT'
		  AND period <= $2
		ORDER BY operation_date
	`

	rows, err := r.db.QueryContext(ctx, query, accountIDs, upToPeriod)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []DepositsInfo{}, nil
		}

		return nil, fmt.Errorf("failed to query deposit info: %w", err)
	}
	defer rows.Close()

	var infos []DepositsInfo

	for rows.Next() {
		var info DepositsInfo
		err := rows.Scan(&info.TotalAmount, &info.Currency, &info.OperationDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deposit info: %w", err)
		}
		infos = append(infos, info)
	}

	return infos, nil
}
