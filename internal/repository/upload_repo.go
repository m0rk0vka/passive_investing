package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// UploadRepository handles upload operations
type UploadRepository struct {
	db *sql.DB
}

// NewUploadRepository creates a new upload repository
func NewUploadRepository(db *sql.DB) *UploadRepository {
	return &UploadRepository{db: db}
}

// CreateUpload creates a new upload record
func (r *UploadRepository) CreateUpload(ctx context.Context, upload *Upload) error {
	query := `
		INSERT INTO upload (
			tg_user_id, tg_chat_id, telegram_message_id, telegram_file_id,
			telegram_file_unique_id, original_filename, mime_type, file_size, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		upload.TgUserID, upload.TgChatID, upload.TelegramMessageID,
		upload.TelegramFileID, upload.TelegramFileUniqueID,
		upload.OriginalFilename, upload.MimeType, upload.FileSize, upload.Status,
	).Scan(&upload.ID, &upload.CreatedAt, &upload.UpdatedAt)

	return err
}

// UpdateUploadStatus updates the status of an upload
func (r *UploadRepository) UpdateUploadStatus(ctx context.Context, id int64, status UploadStatus, errorMsg *string) error {
	query := `
		UPDATE upload
		SET status = $1, error_message = $2, updated_at = now()
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, status, errorMsg, id)
	return err
}

// GetPendingUploads returns uploads with RECEIVED status
func (r *UploadRepository) GetPendingUploads(ctx context.Context, limit int) ([]Upload, error) {
	query := `
		SELECT id, tg_user_id, tg_chat_id, telegram_message_id, telegram_file_id,
		       telegram_file_unique_id, original_filename, mime_type, file_size,
		       status, error_message, created_at, updated_at
		FROM upload
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, UploadStatusReceived, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploads []Upload
	for rows.Next() {
		var u Upload
		err := rows.Scan(
			&u.ID, &u.TgUserID, &u.TgChatID, &u.TelegramMessageID,
			&u.TelegramFileID, &u.TelegramFileUniqueID,
			&u.OriginalFilename, &u.MimeType, &u.FileSize,
			&u.Status, &u.ErrorMessage, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		uploads = append(uploads, u)
	}

	return uploads, rows.Err()
}

// GetUploadByID returns an upload by ID
func (r *UploadRepository) GetUploadByID(ctx context.Context, id int64) (*Upload, error) {
	query := `
		SELECT id, tg_user_id, tg_chat_id, telegram_message_id, telegram_file_id,
		       telegram_file_unique_id, original_filename, mime_type, file_size,
		       status, error_message, created_at, updated_at
		FROM upload
		WHERE id = $1
	`

	var u Upload
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.TgUserID, &u.TgChatID, &u.TelegramMessageID,
		&u.TelegramFileID, &u.TelegramFileUniqueID,
		&u.OriginalFilename, &u.MimeType, &u.FileSize,
		&u.Status, &u.ErrorMessage, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// CreateRawFile creates a raw file record
func (r *UploadRepository) CreateRawFile(ctx context.Context, rawFile *RawFile) error {
	query := `
		INSERT INTO raw_file (upload_id, sha256, storage_kind, storage_key)
		VALUES ($1, $2, $3, $4)
		RETURNING id, stored_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		rawFile.UploadID, rawFile.SHA256, rawFile.StorageKind, rawFile.StorageKey,
	).Scan(&rawFile.ID, &rawFile.StoredAt)

	return err
}

// GetRawFileByUploadID returns raw file by upload ID
func (r *UploadRepository) GetRawFileByUploadID(ctx context.Context, uploadID int64) (*RawFile, error) {
	query := `
		SELECT id, upload_id, sha256, storage_kind, storage_key, stored_at
		FROM raw_file
		WHERE upload_id = $1
	`

	var rf RawFile
	err := r.db.QueryRowContext(ctx, query, uploadID).Scan(
		&rf.ID, &rf.UploadID, &rf.SHA256, &rf.StorageKind, &rf.StorageKey, &rf.StoredAt,
	)
	if err != nil {
		return nil, err
	}

	return &rf, nil
}

// EnsureTgUser creates or returns existing telegram user
func (r *UploadRepository) EnsureTgUser(ctx context.Context, telegramUserID int64) (int64, error) {
	// Try to get existing
	var id int64
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM tg_user WHERE telegram_user_id = $1
	`, telegramUserID).Scan(&id)

	if err == nil {
		return id, nil
	}

	if err != sql.ErrNoRows {
		return 0, err
	}

	// Create new
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO tg_user (telegram_user_id)
		VALUES ($1)
		RETURNING id
	`, telegramUserID).Scan(&id)

	return id, err
}

// EnsureTgChat creates or returns existing telegram chat
func (r *UploadRepository) EnsureTgChat(ctx context.Context, telegramChatID int64) (int64, error) {
	// Try to get existing
	var id int64
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM tg_chat WHERE telegram_chat_id = $1
	`, telegramChatID).Scan(&id)

	if err == nil {
		return id, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("query tg_chat: %w", err)
	}

	// Create new
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO tg_chat (telegram_chat_id)
		VALUES ($1)
		RETURNING id
	`, telegramChatID).Scan(&id)

	return id, err
}
