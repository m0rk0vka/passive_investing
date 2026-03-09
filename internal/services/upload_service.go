package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/m0rk0vka/passive_investing/internal/repository"
	"go.uber.org/zap"
)

// UploadService handles upload business logic
type UploadService struct {
	uploadRepo *repository.UploadRepository
	logger     *zap.Logger
}

// NewUploadService creates a new upload service
func NewUploadService(db *sql.DB, logger *zap.Logger) *UploadService {
	return &UploadService{
		uploadRepo: repository.NewUploadRepository(db),
		logger:     logger,
	}
}

// SaveUpload saves upload metadata to database
func (s *UploadService) SaveUpload(ctx context.Context, telegramUserID, telegramChatID int64, messageID *int64, fileID, fileUniqueID, filename, mimeType *string, fileSize *int64) (int64, error) {
	// Ensure telegram user exists
	tgUserID, err := s.uploadRepo.EnsureTgUser(ctx, telegramUserID)
	if err != nil {
		return 0, fmt.Errorf("ensure tg_user: %w", err)
	}

	// Ensure telegram chat exists
	tgChatID, err := s.uploadRepo.EnsureTgChat(ctx, telegramChatID)
	if err != nil {
		return 0, fmt.Errorf("ensure tg_chat: %w", err)
	}

	// Create upload record
	upload := &repository.Upload{
		TgUserID:             tgUserID,
		TgChatID:             tgChatID,
		TelegramMessageID:    messageID,
		TelegramFileID:       fileID,
		TelegramFileUniqueID: fileUniqueID,
		OriginalFilename:     filename,
		MimeType:             mimeType,
		FileSize:             fileSize,
		Status:               repository.UploadStatusReceived,
	}

	if err := s.uploadRepo.CreateUpload(ctx, upload); err != nil {
		return 0, fmt.Errorf("create upload: %w", err)
	}

	s.logger.Info("upload saved",
		zap.Int64("upload_id", upload.ID),
		zap.Int64("telegram_user_id", telegramUserID),
		zap.String("filename", stringOrEmpty(filename)),
	)

	return upload.ID, nil
}

// SaveRawFile saves raw file metadata
func (s *UploadService) SaveRawFile(ctx context.Context, uploadID int64, sha256, storageKind, storageKey string) error {
	rawFile := &repository.RawFile{
		UploadID:    uploadID,
		SHA256:      sha256,
		StorageKind: storageKind,
		StorageKey:  storageKey,
	}

	if err := s.uploadRepo.CreateRawFile(ctx, rawFile); err != nil {
		return fmt.Errorf("create raw_file: %w", err)
	}

	s.logger.Info("raw file saved",
		zap.Int64("upload_id", uploadID),
		zap.String("sha256", sha256),
	)

	return nil
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
