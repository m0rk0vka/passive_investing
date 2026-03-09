package worker

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/m0rk0vka/passive_investing/internal/parsing"
	"github.com/m0rk0vka/passive_investing/internal/repository"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

const (
	pollInterval = 5 * time.Second
	batchSize    = 10
)

// ParserWorker processes pending uploads
type ParserWorker struct {
	db           *sql.DB
	logger       *zap.Logger
	uploadRepo   *repository.UploadRepository
	snapshotRepo *repository.SnapshotRepository
	accountRepo  *repository.AccountRepository
	cashFlowRepo *repository.CashFlowRepository
}

// NewParserWorker creates a new parser worker
func NewParserWorker(db *sql.DB, logger *zap.Logger) *ParserWorker {
	return &ParserWorker{
		db:           db,
		logger:       logger,
		uploadRepo:   repository.NewUploadRepository(db),
		snapshotRepo: repository.NewSnapshotRepository(db),
		accountRepo:  repository.NewAccountRepository(db),
		cashFlowRepo: repository.NewCashFlowRepository(db),
	}
}

// Start starts the worker loop
func (w *ParserWorker) Start(ctx context.Context) {
	w.logger.Info("Parser worker started")
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Parser worker stopped")
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.logger.Error("Error processing batch", zap.Error(err))
			}
		}
	}
}

// processBatch processes a batch of pending uploads
func (w *ParserWorker) processBatch(ctx context.Context) error {
	uploads, err := w.uploadRepo.GetPendingUploads(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("get pending uploads: %w", err)
	}

	if len(uploads) == 0 {
		return nil
	}

	w.logger.Info("Processing uploads", zap.Int("count", len(uploads)))

	for _, upload := range uploads {
		if err := w.processUpload(ctx, &upload); err != nil {
			w.logger.Error("Error processing upload",
				zap.Int64("upload_id", upload.ID),
				zap.Error(err),
			)
			// Mark as failed
			errMsg := err.Error()
			_ = w.uploadRepo.UpdateUploadStatus(ctx, upload.ID, repository.UploadStatusFailed, &errMsg)
		}
	}

	return nil
}

// processUpload processes a single upload
func (w *ParserWorker) processUpload(ctx context.Context, upload *repository.Upload) error {
	// Update status to PROCESSING
	if err := w.uploadRepo.UpdateUploadStatus(ctx, upload.ID, repository.UploadStatusProcessing, nil); err != nil {
		return fmt.Errorf("update status to processing: %w", err)
	}

	// Get raw file
	rawFile, err := w.uploadRepo.GetRawFileByUploadID(ctx, upload.ID)
	if err != nil {
		return fmt.Errorf("get raw file: %w", err)
	}

	// Parse XLSX file
	data, err := w.parseFile(rawFile.StorageKey)
	if err != nil {
		return fmt.Errorf("parse file: %w", err)
	}

	// Ensure account exists
	accountID, err := w.accountRepo.EnsureAccount(
		ctx,
		data.AccountNumber,
		"ВТБ", // institution
		"BROKER",
		"RUB",
	)
	if err != nil {
		return fmt.Errorf("ensure account: %w", err)
	}

	// Save snapshots
	if err := w.saveSnapshots(ctx, accountID, data); err != nil {
		return fmt.Errorf("save snapshots: %w", err)
	}

	// Save cash flow operations
	if err := w.saveCashFlowOperations(ctx, accountID, data); err != nil {
		return fmt.Errorf("save cash flow operations: %w", err)
	}

	// Mark as done
	if err := w.uploadRepo.UpdateUploadStatus(ctx, upload.ID, repository.UploadStatusDone, nil); err != nil {
		return fmt.Errorf("update status to done: %w", err)
	}

	w.logger.Info("Upload processed successfully",
		zap.Int64("upload_id", upload.ID),
		zap.Int64("account_id", accountID),
		zap.String("period", formatPeriod(data.PeriodEnd)),
	)

	return nil
}

// parseFile parses XLSX file and returns portfolio data
func (w *ParserWorker) parseFile(filePath string) (*parsing.PortfolioData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()

	// Get all rows from first sheet
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	// Parse using existing parser
	data := parsing.ParsePositions(rows)
	return data, nil
}

// saveSnapshots saves valuation and position snapshots
func (w *ParserWorker) saveSnapshots(ctx context.Context, accountID int64, data *parsing.PortfolioData) error {
	period := formatPeriod(data.PeriodEnd)

	// Calculate securities value
	securitiesValue := decimal.Zero
	for _, pos := range data.Positions {
		securitiesValue = securitiesValue.Add(pos.MarketValue)
	}

	// Save valuation snapshot
	valuationSnap := &repository.ValuationSnapshot{
		AccountID:       accountID,
		Period:          period,
		TotalValue:      data.TotalAssets,
		CashBalance:     data.CashBalance,
		SecuritiesValue: securitiesValue,
		Currency:        "RUB",
	}

	if err := w.snapshotRepo.UpsertValuationSnapshot(ctx, valuationSnap); err != nil {
		return fmt.Errorf("upsert valuation snapshot: %w", err)
	}

	// Save position snapshots
	for _, pos := range data.Positions {
		posSnap := &repository.PositionSnapshot{
			AccountID:    accountID,
			Period:       period,
			ISIN:         pos.ISIN,
			SecurityName: pos.Name,
			Quantity:     pos.Qty,
			Price:        pos.Price,
			MarketValue:  pos.MarketValue,
			Currency:     pos.Currency,
		}

		if err := w.snapshotRepo.UpsertPositionSnapshot(ctx, posSnap); err != nil {
			return fmt.Errorf("upsert position snapshot: %w", err)
		}
	}

	return nil
}

// formatPeriod formats time to YYYY-MM period string
func formatPeriod(t time.Time) string {
	return t.Format("2006-01")
}

// saveCashFlowOperations saves cash flow operations
func (w *ParserWorker) saveCashFlowOperations(ctx context.Context, accountID int64, data *parsing.PortfolioData) error {
	period := formatPeriod(data.PeriodEnd)

	// Фильтруем и конвертируем операции
	var operations []repository.CashFlowOperation
	for _, op := range data.CashFlow {

		operations = append(operations, repository.CashFlowOperation{
			AccountID:     accountID,
			Period:        period,
			OperationDate: op.Date,
			Amount:        op.Amount,
			Currency:      op.Currency,
			OperationType: convertCashFlowType(op.OperationType),
			Comment:       op.Comment,
		})
	}

	// Идемпотентное сохранение: удаляем старые и вставляем новые
	if err := w.cashFlowRepo.ReplaceOperationsForPeriod(ctx, accountID, period, operations); err != nil {
		return fmt.Errorf("replace cash flow operations: %w", err)
	}

	return nil
}

// convertCashFlowType конвертирует тип операции из parsing в repository
func convertCashFlowType(t parsing.CashFlowOperationType) repository.CashFlowOperationType {
	return repository.CashFlowOperationType(t)
}
