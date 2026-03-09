package repository

import (
	"context"
	"database/sql"
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
