package uploads

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/jackc/pgx/v5"

	"arkive/core/database"
	"arkive/core/models"
	filerepo "arkive/core/repositories/files"
	uploadrepo "arkive/core/repositories/uploads"
	"arkive/pkg/storage/r2"
	"arkive/pkg/tokens"
	"arkive/pkg/validation"
)

const (
	fileStatusPending   = "pending"
	fileStatusUploading = "uploading"
	fileStatusComplete  = "complete"
	fileStatusFailed    = "failed"
	fileStatusAborted   = "aborted"

	multipartStatusInitiated = "initiated"
	multipartStatusUploading = "uploading"
	multipartStatusCompleted = "completed"
	multipartStatusAborted   = "aborted"
	multipartStatusFailed    = "failed"
)

type Service struct {
	db             database.PgPool
	fileRepo       *filerepo.Repository
	uploadRepo     *uploadrepo.Repository
	r2             *r2.Client
	bucket         string
	uploadExpires  time.Duration
	downloadExpire time.Duration
}

type Config struct {
	Bucket         string
	UploadExpires  time.Duration
	DownloadExpire time.Duration
}

type StartResponse struct {
	FileID      string
	MultipartID string
	ObjectKey   string
	ChunkSize   int
	TotalParts  int
}

type CompletedPartInput struct {
	PartNumber int32  `json:"partNumber"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size,omitempty"`
}

type storedPart struct {
	PartNumber int32  `json:"part"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
}

func NewService(
	db database.PgPool,
	fileRepo *filerepo.Repository,
	uploadRepo *uploadrepo.Repository,
	r2Client *r2.Client,
	cfg Config,
) *Service {
	return &Service{
		db:             db,
		fileRepo:       fileRepo,
		uploadRepo:     uploadRepo,
		r2:             r2Client,
		bucket:         cfg.Bucket,
		uploadExpires:  cfg.UploadExpires,
		downloadExpire: cfg.DownloadExpire,
	}
}

func (s *Service) StartMultipart(ctx context.Context, userID, filename string, sizeBytes int64, contentType string) (StartResponse, validation.Errors, error) {
	userID = strings.TrimSpace(userID)
	filename = strings.TrimSpace(filename)
	contentType = strings.TrimSpace(contentType)
	validationErrors := validation.New()
	if userID == "" {
		return StartResponse{}, nil, ErrUnauthorized
	}
	if filename == "" {
		validationErrors.Add("filename", ErrFilenameRequired.Error())
	}
	if sizeBytes <= 0 {
		validationErrors.Add("size", ErrFileSizeRequired.Error())
	}
	if validationErrors.HasAny() {
		return StartResponse{}, validationErrors, nil
	}

	chunkSize := chooseChunkSize(sizeBytes)
	partCount := totalParts(sizeBytes, chunkSize)

	objectKey, err := buildObjectKey(userID)
	if err != nil {
		return StartResponse{}, nil, err
	}

	uploadID, err := s.r2.CreateMultipartUpload(ctx, objectKey, contentType)
	if err != nil {
		return StartResponse{}, nil, err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return StartResponse{}, nil, err
	}

	file, err := s.fileRepo.CreateFile(ctx, tx, models.File{
		UserID:      userID,
		Bucket:      s.bucket,
		ObjectKey:   objectKey,
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   sizeBytes,
		Status:      fileStatusUploading,
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return StartResponse{}, nil, err
	}

	multipart, err := s.uploadRepo.CreateMultipart(ctx, tx, models.MultipartUpload{
		FileID:        file.ID,
		UploadID:      uploadID,
		Bucket:        s.bucket,
		ObjectKey:     objectKey,
		ChunkSize:     chunkSize,
		TotalParts:    partCount,
		UploadedParts: []byte("[]"),
		Status:        multipartStatusInitiated,
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return StartResponse{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return StartResponse{}, nil, err
	}

	return StartResponse{
		FileID:      file.ID,
		MultipartID: multipart.ID,
		ObjectKey:   objectKey,
		ChunkSize:   chunkSize,
		TotalParts:  partCount,
	}, nil, nil
}

func (s *Service) PresignPart(ctx context.Context, userID, multipartID string, partNumber int32) (string, error) {
	userID = strings.TrimSpace(userID)
	multipartID = strings.TrimSpace(multipartID)
	if userID == "" {
		return "", ErrUnauthorized
	}
	if multipartID == "" || partNumber <= 0 {
		return "", ErrInvalidInput
	}

	upload, err := s.uploadRepo.GetMultipartForUser(ctx, s.db, multipartID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if upload.Status == multipartStatusCompleted || upload.Status == multipartStatusAborted {
		return "", ErrInvalidInput
	}
	if partNumber > int32(upload.TotalParts) {
		return "", ErrInvalidInput
	}

	if err := s.uploadRepo.TouchMultipart(ctx, s.db, multipartID, multipartStatusUploading); err != nil {
		return "", err
	}

	return s.r2.PresignUploadPart(ctx, upload.ObjectKey, upload.UploadID, partNumber, s.uploadExpires)
}

func (s *Service) CompleteMultipart(ctx context.Context, userID, multipartID string, parts []CompletedPartInput) error {
	userID = strings.TrimSpace(userID)
	multipartID = strings.TrimSpace(multipartID)
	if userID == "" {
		return ErrUnauthorized
	}
	if multipartID == "" {
		return ErrInvalidInput
	}
	if len(parts) == 0 {
		return ErrPartsRequired
	}

	upload, err := s.uploadRepo.GetMultipartForUser(ctx, s.db, multipartID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if upload.Status == multipartStatusCompleted || upload.Status == multipartStatusAborted {
		return ErrInvalidInput
	}

	type dedupeKey struct {
		PartNumber int32
	}
	seen := map[dedupeKey]bool{}
	completed := make([]types.CompletedPart, 0, len(parts))
	stored := make([]storedPart, 0, len(parts))
	for _, part := range parts {
		if part.PartNumber <= 0 || strings.TrimSpace(part.ETag) == "" {
			return ErrInvalidInput
		}
		if part.PartNumber > int32(upload.TotalParts) {
			return ErrInvalidInput
		}
		key := dedupeKey{PartNumber: part.PartNumber}
		if seen[key] {
			return ErrInvalidInput
		}
		seen[key] = true
		etag := strings.TrimSpace(part.ETag)
		completed = append(completed, types.CompletedPart{
			PartNumber: &part.PartNumber,
			ETag:       &etag,
		})
		stored = append(stored, storedPart{
			PartNumber: part.PartNumber,
			ETag:       etag,
			Size:       part.Size,
		})
	}

	sort.Slice(completed, func(i, j int) bool {
		if completed[i].PartNumber == nil || completed[j].PartNumber == nil {
			return false
		}
		return *completed[i].PartNumber < *completed[j].PartNumber
	})

	if err := s.r2.CompleteMultipartUpload(ctx, upload.ObjectKey, upload.UploadID, completed); err != nil {
		_ = s.uploadRepo.TouchMultipart(ctx, s.db, multipartID, multipartStatusFailed)
		_ = s.fileRepo.UpdateFileStatus(ctx, s.db, upload.FileID, fileStatusFailed)
		return err
	}

	storedJSON, err := json.Marshal(stored)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	if err := s.uploadRepo.UpdateMultipart(ctx, tx, multipartID, multipartStatusCompleted, storedJSON); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := s.fileRepo.UpdateFileStatus(ctx, tx, upload.FileID, fileStatusComplete); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) AbortMultipart(ctx context.Context, userID, multipartID string) error {
	userID = strings.TrimSpace(userID)
	multipartID = strings.TrimSpace(multipartID)
	if userID == "" {
		return ErrUnauthorized
	}
	if multipartID == "" {
		return ErrInvalidInput
	}

	upload, err := s.uploadRepo.GetMultipartForUser(ctx, s.db, multipartID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if err := s.r2.AbortMultipartUpload(ctx, upload.ObjectKey, upload.UploadID); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if err := s.uploadRepo.UpdateMultipart(ctx, tx, multipartID, multipartStatusAborted, upload.UploadedParts); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := s.fileRepo.UpdateFileStatus(ctx, tx, upload.FileID, fileStatusAborted); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) PresignDownload(ctx context.Context, userID, fileID string) (string, error) {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return "", ErrUnauthorized
	}
	if fileID == "" {
		return "", ErrInvalidInput
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}

	return s.r2.PresignDownload(ctx, file.ObjectKey, s.downloadExpire)
}

func (s *Service) CreateShare(ctx context.Context, userID, fileID string, expiresAt *time.Time) (string, error) {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return "", ErrUnauthorized
	}
	if fileID == "" {
		return "", ErrInvalidInput
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}

	token, hash, err := tokens.Generate()
	if err != nil {
		return "", err
	}

	_, err = s.fileRepo.CreateShare(ctx, s.db, models.FileShare{
		FileID:    file.ID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) PresignShareDownload(ctx context.Context, token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrInvalidInput
	}

	hash := sha256.Sum256([]byte(token))
	share, err := s.fileRepo.GetShareByTokenHash(ctx, s.db, hash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return "", ErrNotFound
	}

	file, err := s.fileRepo.GetFileByID(ctx, s.db, share.FileID)
	if err != nil {
		return "", err
	}

	return s.r2.PresignDownload(ctx, file.ObjectKey, s.downloadExpire)
}

func (s *Service) CleanupStaleMultipart(ctx context.Context, maxAge time.Duration) (int, error) {
	if maxAge <= 0 {
		return 0, ErrInvalidInput
	}
	olderThan := time.Now().Add(-maxAge)
	uploads, err := s.uploadRepo.ListStaleMultipart(ctx, s.db, olderThan)
	if err != nil {
		return 0, err
	}

	cleaned := 0
	for _, upload := range uploads {
		if err := s.r2.AbortMultipartUpload(ctx, upload.ObjectKey, upload.UploadID); err != nil {
			continue
		}

		tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			continue
		}

		if err := s.uploadRepo.UpdateMultipart(ctx, tx, upload.ID, multipartStatusAborted, upload.UploadedParts); err != nil {
			_ = tx.Rollback(ctx)
			continue
		}
		if err := s.fileRepo.UpdateFileStatus(ctx, tx, upload.FileID, fileStatusAborted); err != nil {
			_ = tx.Rollback(ctx)
			continue
		}

		if err := tx.Commit(ctx); err != nil {
			continue
		}
		cleaned++
	}

	return cleaned, nil
}

func buildObjectKey(userID string) (string, error) {
	token, _, err := tokens.Generate()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("u/%s/%s", userID, token), nil
}

func chooseChunkSize(sizeBytes int64) int {
	chunk := 10 * 1024 * 1024
	if sizeBytes > 2*1024*1024*1024 {
		chunk = 25 * 1024 * 1024
	}
	return chunk
}

func totalParts(sizeBytes int64, chunkSize int) int {
	return int((sizeBytes + int64(chunkSize) - 1) / int64(chunkSize))
}
