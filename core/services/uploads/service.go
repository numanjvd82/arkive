package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/jackc/pgx/v5"

	"arkive/core/database"
	"arkive/core/models"
	filerepo "arkive/core/repositories/files"
	storagerepo "arkive/core/repositories/storage"
	uploadrepo "arkive/core/repositories/uploads"
	"arkive/pkg/safeptr"
	"arkive/pkg/storage"
	"arkive/pkg/storage/r2"
	"arkive/pkg/validation"
)

const (
	FileStatusPending   = "pending"
	FileStatusUploading = "uploading"
	FileStatusComplete  = "complete"
	FileStatusFailed    = "failed"
	FileStatusAborted   = "aborted"

	MultipartStatusInitiated = "initiated"
	MultipartStatusUploading = "uploading"
	MultipartStatusCompleted = "completed"
	MultipartStatusAborted   = "aborted"
	MultipartStatusFailed    = "failed"
)

const (
	MaxFileSizeBytes        int64 = 5 * 1024 * 1024 * 1024
	MultipartThresholdBytes int64 = 200 * 1024 * 1024
)

type Service struct {
	db                  database.PgPool
	storageRepo         *storagerepo.Repository
	fileRepo            *filerepo.Repository
	uploadRepo          *uploadrepo.Repository
	r2                  *r2.Client
	bucket              string
	uploadExpires       time.Duration
	downloadExpire      time.Duration
	shareDownloadExpire time.Duration
}

type Config struct {
	Bucket              string
	UploadExpires       time.Duration
	DownloadExpire      time.Duration
	ShareDownloadExpire time.Duration
}

func NewService(
	db database.PgPool,
	storageRepo *storagerepo.Repository,
	fileRepo *filerepo.Repository,
	uploadRepo *uploadrepo.Repository,
	r2Client *r2.Client,
	cfg Config,
) *Service {
	return &Service{
		db:                  db,
		storageRepo:         storageRepo,
		fileRepo:            fileRepo,
		uploadRepo:          uploadRepo,
		r2:                  r2Client,
		bucket:              cfg.Bucket,
		uploadExpires:       cfg.UploadExpires,
		downloadExpire:      cfg.DownloadExpire,
		shareDownloadExpire: cfg.ShareDownloadExpire,
	}
}

func validateStartInput(userID, filename string, sizeBytes int64, requiresMultipart bool) (validation.Errors, error) {
	validationErrors := validation.New()
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorized
	}
	if strings.TrimSpace(filename) == "" {
		validationErrors.Add("filename", ErrFilenameRequired.Error())
	}
	if sizeBytes <= 0 {
		validationErrors.Add("size", ErrFileSizeRequired.Error())
	}
	if sizeBytes > 0 {
		switch {
		case sizeBytes > MaxFileSizeBytes:
			validationErrors.Add("size", ErrFileTooLarge.Error())
		case requiresMultipart && sizeBytes <= MultipartThresholdBytes:
			validationErrors.Add("size", ErrFileTooSmall.Error())
		case !requiresMultipart && sizeBytes > MultipartThresholdBytes:
			validationErrors.Add("size", ErrMultipartRequired.Error())
		}
	}
	if validationErrors.HasAny() {
		return validationErrors, nil
	}
	return nil, nil
}

func (s *Service) beginUploadTx(ctx context.Context, userID, objectKey, filename, contentType string, sizeBytes int64) (pgx.Tx, models.File, validation.Errors, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, models.File{}, nil, err
	}

	reserved, err := s.storageRepo.ReserveStorage(ctx, tx, userID, sizeBytes)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, models.File{}, nil, err
	}
	if !reserved {
		_ = tx.Rollback(ctx)
		validationErrors := validation.New()
		validationErrors.Add("size", ErrQuotaExceeded.Error())
		return nil, models.File{}, validationErrors, nil
	}

	file, err := s.fileRepo.CreateFile(ctx, tx, models.File{
		UserID:      userID,
		Bucket:      s.bucket,
		ObjectKey:   objectKey,
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   sizeBytes,
		Status:      FileStatusUploading,
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, models.File{}, nil, err
	}

	return tx, file, nil, nil
}

func (s *Service) finalizeFileCompletion(ctx context.Context, tx pgx.Tx, userID string, file models.File, actualSize int64) error {
	if actualSize <= 0 {
		return ErrUploadFailed
	}
	if actualSize > MaxFileSizeBytes {
		return ErrFileTooLarge
	}

	if err := s.fileRepo.UpdateFileSize(ctx, tx, file.ID, actualSize); err != nil {
		return err
	}
	if err := s.fileRepo.UpdateFileStatus(ctx, tx, file.ID, FileStatusComplete); err != nil {
		return err
	}

	if actualSize > file.SizeBytes {
		extra := actualSize - file.SizeBytes
		reserved, err := s.storageRepo.ReserveStorage(ctx, tx, userID, extra)
		if err != nil {
			return err
		}
		if !reserved {
			return ErrQuotaExceeded
		}
	}

	committed, err := s.storageRepo.CommitStorage(ctx, tx, userID, actualSize)
	if err != nil {
		return err
	}
	if !committed {
		return ErrUploadFailed
	}

	if actualSize < file.SizeBytes {
		extra := file.SizeBytes - actualSize
		released, err := s.storageRepo.ReleaseReservedStorage(ctx, tx, userID, extra)
		if err != nil {
			return err
		}
		if !released {
			return ErrUploadFailed
		}
	}

	return nil
}

func (s *Service) markUploadFailed(ctx context.Context, userID, fileID string, sizeBytes int64) {
	updated, err := s.fileRepo.UpdateFileStatusIf(ctx, s.db, fileID, FileStatusFailed, []string{FileStatusPending, FileStatusUploading})
	if err != nil || !updated {
		return
	}
	_, _ = s.storageRepo.ReleaseReservedStorage(ctx, s.db, userID, sizeBytes)
}

func buildCompletedFromList(listed []types.Part, totalParts int) ([]types.CompletedPart, []models.StoredPart, []int32) {
	partsByNumber := make(map[int32]types.Part, len(listed))
	for _, part := range listed {
		number := safeptr.Int32(part.PartNumber)
		etag := safeptr.String(part.ETag)
		if number <= 0 || number > int32(totalParts) || strings.TrimSpace(etag) == "" {
			continue
		}
		partsByNumber[number] = part
	}

	missing := make([]int32, 0)
	for i := 1; i <= totalParts; i++ {
		if _, ok := partsByNumber[int32(i)]; !ok {
			missing = append(missing, int32(i))
		}
	}
	if len(missing) > 0 {
		return nil, nil, missing
	}

	completed := make([]types.CompletedPart, 0, totalParts)
	stored := make([]models.StoredPart, 0, totalParts)
	for i := 1; i <= totalParts; i++ {
		part := partsByNumber[int32(i)]
		number := int32(i)
		etag := strings.TrimSpace(safeptr.String(part.ETag))
		completed = append(completed, types.CompletedPart{
			PartNumber: &number,
			ETag:       &etag,
		})
		stored = append(stored, models.StoredPart{
			PartNumber: number,
			ETag:       etag,
			Size:       safeptr.Int64(part.Size),
		})
	}

	return completed, stored, nil
}

func (s *Service) StartMultipart(ctx context.Context, userID, filename string, sizeBytes int64, contentType string) (models.MultipartStartResponse, validation.Errors, error) {
	userID = strings.TrimSpace(userID)
	filename = strings.TrimSpace(filename)
	contentType = strings.TrimSpace(contentType)
	validationErrors, err := validateStartInput(userID, filename, sizeBytes, true)
	if err != nil || (validationErrors != nil && validationErrors.HasAny()) {
		return models.MultipartStartResponse{}, validationErrors, err
	}

	chunkSize := storage.ChooseChunkSize(sizeBytes, MaxFileSizeBytes)
	partCount := storage.TotalParts(sizeBytes, chunkSize)

	objectKey, err := storage.BuildObjectKey(userID)
	if err != nil {
		return models.MultipartStartResponse{}, nil, err
	}

	uploadID, err := s.r2.CreateMultipartUpload(ctx, objectKey, contentType)
	if err != nil {
		return models.MultipartStartResponse{}, nil, err
	}

	tx, file, validationErrors, err := s.beginUploadTx(ctx, userID, objectKey, filename, contentType, sizeBytes)
	if err != nil || (validationErrors != nil && validationErrors.HasAny()) {
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return models.MultipartStartResponse{}, validationErrors, err
	}

	multipart, err := s.uploadRepo.CreateMultipart(ctx, tx, models.MultipartUpload{
		FileID:        file.ID,
		UploadID:      uploadID,
		Bucket:        s.bucket,
		ObjectKey:     objectKey,
		ChunkSize:     chunkSize,
		TotalParts:    partCount,
		UploadedParts: []byte("[]"),
		Status:        MultipartStatusInitiated,
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return models.MultipartStartResponse{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		_ = s.r2.AbortMultipartUpload(ctx, objectKey, uploadID)
		return models.MultipartStartResponse{}, nil, err
	}

	return models.MultipartStartResponse{
		FileID:      file.ID,
		MultipartID: multipart.ID,
		ObjectKey:   objectKey,
		ChunkSize:   chunkSize,
		TotalParts:  partCount,
	}, nil, nil
}

func (s *Service) StartUpload(ctx context.Context, userID, filename string, sizeBytes int64, contentType string) (models.UploadStartResponse, validation.Errors, error) {
	if sizeBytes > MultipartThresholdBytes {
		resp, validationErrors, err := s.StartMultipart(ctx, userID, filename, sizeBytes, contentType)
		if err != nil || (validationErrors != nil && validationErrors.HasAny()) {
			return models.UploadStartResponse{}, validationErrors, err
		}
		return models.UploadStartResponse{
			UploadID:   resp.MultipartID,
			FileID:     resp.FileID,
			ObjectKey:  resp.ObjectKey,
			Mode:       "multipart",
			ChunkSize:  resp.ChunkSize,
			TotalParts: resp.TotalParts,
		}, nil, nil
	}
	resp, validationErrors, err := s.StartSingleUpload(ctx, userID, filename, sizeBytes, contentType)
	if err != nil || (validationErrors != nil && validationErrors.HasAny()) {
		return models.UploadStartResponse{}, validationErrors, err
	}
	return models.UploadStartResponse{
		UploadID:  resp.FileID,
		FileID:    resp.FileID,
		ObjectKey: resp.ObjectKey,
		Mode:      "single",
		UploadURL: resp.UploadURL,
	}, nil, nil
}

func (s *Service) StartSingleUpload(ctx context.Context, userID, filename string, sizeBytes int64, contentType string) (models.SingleStartResponse, validation.Errors, error) {
	userID = strings.TrimSpace(userID)
	filename = strings.TrimSpace(filename)
	contentType = strings.TrimSpace(contentType)
	validationErrors, err := validateStartInput(userID, filename, sizeBytes, false)
	if err != nil || (validationErrors != nil && validationErrors.HasAny()) {
		return models.SingleStartResponse{}, validationErrors, err
	}

	objectKey, err := storage.BuildObjectKey(userID)
	if err != nil {
		return models.SingleStartResponse{}, nil, err
	}

	uploadURL, err := s.r2.PresignUpload(ctx, objectKey, contentType, s.uploadExpires)
	if err != nil {
		return models.SingleStartResponse{}, nil, err
	}

	tx, file, validationErrors, err := s.beginUploadTx(ctx, userID, objectKey, filename, contentType, sizeBytes)
	if err != nil || (validationErrors != nil && validationErrors.HasAny()) {
		return models.SingleStartResponse{}, validationErrors, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.SingleStartResponse{}, nil, err
	}

	return models.SingleStartResponse{
		FileID:    file.ID,
		ObjectKey: objectKey,
		UploadURL: uploadURL,
		ExpiresAt: time.Now().Add(s.uploadExpires),
	}, nil, nil
}

func (s *Service) NextUpload(ctx context.Context, userID, uploadID string, uploadedParts []int32) (models.UploadNextResponse, error) {
	userID = strings.TrimSpace(userID)
	uploadID = strings.TrimSpace(uploadID)
	if userID == "" {
		return models.UploadNextResponse{}, ErrUnauthorized
	}
	if uploadID == "" {
		return models.UploadNextResponse{}, ErrInvalidInput
	}

	upload, err := s.uploadRepo.GetMultipartForUser(ctx, s.db, uploadID, userID)
	if err == nil {
		return s.nextMultipart(ctx, upload, uploadedParts)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return models.UploadNextResponse{}, err
	}

	upload, err = s.uploadRepo.GetMultipartForFile(ctx, s.db, uploadID, userID)
	if err == nil {
		return s.nextMultipart(ctx, upload, uploadedParts)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return models.UploadNextResponse{}, err
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, uploadID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UploadNextResponse{}, ErrNotFound
		}
		return models.UploadNextResponse{}, err
	}
	if file.Status == FileStatusComplete {
		return models.UploadNextResponse{}, ErrNoNextPart
	}
	if file.Status == FileStatusAborted || file.Status == FileStatusFailed {
		return models.UploadNextResponse{}, ErrUploadCancelled
	}

	uploadURL, err := s.r2.PresignUpload(ctx, file.ObjectKey, file.ContentType, s.uploadExpires)
	if err != nil {
		return models.UploadNextResponse{}, err
	}

	return models.UploadNextResponse{
		UploadID: uploadID,
		FileID:   file.ID,
		Mode:     "single",
		URL:      uploadURL,
	}, nil
}

func (s *Service) nextMultipart(ctx context.Context, upload models.MultipartUpload, uploadedParts []int32) (models.UploadNextResponse, error) {
	if upload.Status == MultipartStatusCompleted {
		return models.UploadNextResponse{}, ErrNoNextPart
	}
	if upload.Status == MultipartStatusAborted || upload.Status == MultipartStatusFailed {
		return models.UploadNextResponse{}, ErrUploadCancelled
	}

	partsByNumber := make(map[int32]bool, len(uploadedParts))
	for _, number := range uploadedParts {
		if number <= 0 || number > int32(upload.TotalParts) {
			continue
		}
		partsByNumber[number] = true
	}

	resumeParts := make([]models.ResumePart, 0)
	listed, listErr := s.r2.ListParts(ctx, upload.ObjectKey, upload.UploadID)
	if listErr == nil {
		partsByNumber = make(map[int32]bool, len(listed))
		resumeParts = make([]models.ResumePart, 0, len(listed))
		for _, part := range listed {
			number := safeptr.Int32(part.PartNumber)
			etag := strings.TrimSpace(safeptr.String(part.ETag))
			if number <= 0 || number > int32(upload.TotalParts) || etag == "" {
				continue
			}
			partsByNumber[number] = true
			resumeParts = append(resumeParts, models.ResumePart{
				PartNumber: number,
				ETag:       etag,
				Size:       safeptr.Int64(part.Size),
			})
		}
	}

	var nextPart int32
	for i := 1; i <= upload.TotalParts; i++ {
		partNumber := int32(i)
		if !partsByNumber[partNumber] {
			nextPart = partNumber
			break
		}
	}
	if nextPart == 0 {
		return models.UploadNextResponse{}, ErrNoNextPart
	}

	_, err := s.uploadRepo.UpdateMultipartStatusIf(ctx, s.db, upload.ID, MultipartStatusUploading, []string{MultipartStatusInitiated})
	if err != nil {
		return models.UploadNextResponse{}, err
	}

	url, err := s.r2.PresignUploadPart(ctx, upload.ObjectKey, upload.UploadID, nextPart, s.uploadExpires)
	if err != nil {
		return models.UploadNextResponse{}, err
	}

	return models.UploadNextResponse{
		UploadID:      upload.ID,
		FileID:        upload.FileID,
		Mode:          "multipart",
		NextPart:      nextPart,
		URL:           url,
		ChunkSize:     upload.ChunkSize,
		TotalParts:    upload.TotalParts,
		UploadedParts: resumeParts,
	}, nil
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
	if upload.Status == MultipartStatusCompleted || upload.Status == MultipartStatusAborted {
		return "", ErrInvalidInput
	}
	if partNumber > int32(upload.TotalParts) {
		return "", ErrInvalidInput
	}

	if err := s.uploadRepo.TouchMultipart(ctx, s.db, multipartID, MultipartStatusUploading); err != nil {
		return "", err
	}

	return s.r2.PresignUploadPart(ctx, upload.ObjectKey, upload.UploadID, partNumber, s.uploadExpires)
}

func (s *Service) CompleteUpload(ctx context.Context, userID, uploadID string, parts []models.CompletedPartInput) error {
	userID = strings.TrimSpace(userID)
	uploadID = strings.TrimSpace(uploadID)
	if userID == "" {
		return ErrUnauthorized
	}
	if uploadID == "" {
		return ErrInvalidInput
	}

	upload, err := s.uploadRepo.GetMultipartForUser(ctx, s.db, uploadID, userID)
	if err == nil {
		return s.CompleteMultipart(ctx, userID, upload.ID, parts)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	upload, err = s.uploadRepo.GetMultipartForFile(ctx, s.db, uploadID, userID)
	if err == nil {
		return s.CompleteMultipart(ctx, userID, upload.ID, parts)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	return s.CompleteSingleUpload(ctx, userID, uploadID)
}

func (s *Service) CancelUpload(ctx context.Context, userID, uploadID string) error {
	userID = strings.TrimSpace(userID)
	uploadID = strings.TrimSpace(uploadID)
	if userID == "" {
		return ErrUnauthorized
	}
	if uploadID == "" {
		return ErrInvalidInput
	}

	upload, err := s.uploadRepo.GetMultipartForUser(ctx, s.db, uploadID, userID)
	if err == nil {
		return s.AbortMultipart(ctx, userID, upload.ID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	upload, err = s.uploadRepo.GetMultipartForFile(ctx, s.db, uploadID, userID)
	if err == nil {
		return s.AbortMultipart(ctx, userID, upload.ID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	return s.AbortSingleUpload(ctx, userID, uploadID)
}

func (s *Service) CompleteMultipart(ctx context.Context, userID, multipartID string, parts []models.CompletedPartInput) error {
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
	if upload.Status == MultipartStatusCompleted {
		return nil
	}
	if upload.Status == MultipartStatusAborted || upload.Status == MultipartStatusFailed {
		return ErrUploadCancelled
	}

	seen := make(map[int32]bool, len(parts))
	completed := make([]types.CompletedPart, 0, len(parts))
	stored := make([]models.StoredPart, 0, len(parts))
	needsList := len(parts) != upload.TotalParts
	for _, part := range parts {
		if part.PartNumber <= 0 || strings.TrimSpace(part.ETag) == "" {
			return ErrInvalidInput
		}
		if part.PartNumber > int32(upload.TotalParts) {
			return ErrInvalidInput
		}
		if seen[part.PartNumber] {
			needsList = true
			continue
		}
		seen[part.PartNumber] = true
		etag := strings.TrimSpace(part.ETag)
		completed = append(completed, types.CompletedPart{
			PartNumber: &part.PartNumber,
			ETag:       &etag,
		})
		stored = append(stored, models.StoredPart{
			PartNumber: part.PartNumber,
			ETag:       etag,
			Size:       part.Size,
		})
	}

	if len(seen) != upload.TotalParts {
		needsList = true
	}

	if needsList {
		listed, err := s.r2.ListParts(ctx, upload.ObjectKey, upload.UploadID)
		if err != nil {
			return err
		}
		var missing []int32
		completed, stored, missing = buildCompletedFromList(listed, upload.TotalParts)
		if len(missing) > 0 {
			return MissingPartsError{Missing: missing}
		}
	}

	sort.Slice(completed, func(i, j int) bool {
		if completed[i].PartNumber == nil || completed[j].PartNumber == nil {
			return false
		}
		return *completed[i].PartNumber < *completed[j].PartNumber
	})

	if err := s.r2.CompleteMultipartUpload(ctx, upload.ObjectKey, upload.UploadID, completed); err != nil {
		_, _ = s.uploadRepo.UpdateMultipartStatusIf(ctx, s.db, multipartID, MultipartStatusFailed, []string{MultipartStatusInitiated, MultipartStatusUploading})
		file, fileErr := s.fileRepo.GetFileByID(ctx, s.db, upload.FileID)
		if fileErr == nil {
			s.markUploadFailed(ctx, userID, file.ID, file.SizeBytes)
		}
		return err
	}

	storedJSON, err := json.Marshal(stored)
	if err != nil {
		return err
	}

	actualSize, err := s.r2.HeadObjectSize(ctx, upload.ObjectKey)
	if err != nil {
		_, _ = s.uploadRepo.UpdateMultipartStatusIf(ctx, s.db, multipartID, MultipartStatusFailed, []string{MultipartStatusInitiated, MultipartStatusUploading})
		file, fileErr := s.fileRepo.GetFileByID(ctx, s.db, upload.FileID)
		if fileErr == nil {
			s.markUploadFailed(ctx, userID, file.ID, file.SizeBytes)
		}
		_ = s.r2.DeleteObject(ctx, upload.ObjectKey)
		return err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	updated, err := s.uploadRepo.UpdateMultipartIf(ctx, tx, multipartID, MultipartStatusCompleted, storedJSON, []string{MultipartStatusInitiated, MultipartStatusUploading})
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !updated {
		_ = tx.Rollback(ctx)
		current, fetchErr := s.uploadRepo.GetMultipartForUser(ctx, s.db, multipartID, userID)
		if fetchErr == nil && current.Status == MultipartStatusCompleted {
			return nil
		}
		if fetchErr == nil && (current.Status == MultipartStatusAborted || current.Status == MultipartStatusFailed) {
			return ErrUploadCancelled
		}
		return ErrInvalidInput
	}
	file, err := s.fileRepo.GetFileByID(ctx, tx, upload.FileID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := s.finalizeFileCompletion(ctx, tx, userID, file, actualSize); err != nil {
		_ = tx.Rollback(ctx)
		_, _ = s.uploadRepo.UpdateMultipartStatusIf(ctx, s.db, multipartID, MultipartStatusFailed, []string{MultipartStatusInitiated, MultipartStatusUploading})
		s.markUploadFailed(ctx, userID, file.ID, file.SizeBytes)
		_ = s.r2.DeleteObject(ctx, upload.ObjectKey)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	s.scheduleVideoMetadata(file)
	return nil
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
	if upload.Status == MultipartStatusCompleted || upload.Status == MultipartStatusAborted || upload.Status == MultipartStatusFailed {
		return nil
	}

	_ = s.r2.AbortMultipartUpload(ctx, upload.ObjectKey, upload.UploadID)
	_ = s.r2.DeleteObject(ctx, upload.ObjectKey)

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	file, err := s.fileRepo.GetFileByID(ctx, tx, upload.FileID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if file.Status == FileStatusComplete || file.Status == FileStatusAborted || file.Status == FileStatusFailed {
		_ = tx.Rollback(ctx)
		return nil
	}
	updatedMultipart, err := s.uploadRepo.UpdateMultipartIf(ctx, tx, multipartID, MultipartStatusAborted, upload.UploadedParts, []string{MultipartStatusInitiated, MultipartStatusUploading})
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !updatedMultipart {
		_ = tx.Rollback(ctx)
		return nil
	}
	updatedFile, err := s.fileRepo.UpdateFileStatusIf(ctx, tx, upload.FileID, FileStatusAborted, []string{FileStatusPending, FileStatusUploading})
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !updatedFile {
		_ = tx.Rollback(ctx)
		return nil
	}
	released, err := s.storageRepo.ReleaseReservedStorage(ctx, tx, userID, file.SizeBytes)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !released {
		_ = tx.Rollback(ctx)
		return ErrUploadFailed
	}

	return tx.Commit(ctx)
}

func (s *Service) CompleteSingleUpload(ctx context.Context, userID, fileID string) error {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return ErrUnauthorized
	}
	if fileID == "" {
		return ErrInvalidInput
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if file.Status == FileStatusComplete {
		return nil
	}
	if file.Status == FileStatusAborted || file.Status == FileStatusFailed {
		return ErrUploadCancelled
	}

	actualSize, err := s.r2.HeadObjectSize(ctx, file.ObjectKey)
	if err != nil {
		s.markUploadFailed(ctx, userID, file.ID, file.SizeBytes)
		_ = s.r2.DeleteObject(ctx, file.ObjectKey)
		return err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	if err := s.finalizeFileCompletion(ctx, tx, userID, file, actualSize); err != nil {
		_ = tx.Rollback(ctx)
		s.markUploadFailed(ctx, userID, file.ID, file.SizeBytes)
		_ = s.r2.DeleteObject(ctx, file.ObjectKey)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	s.scheduleVideoMetadata(file)
	return nil
}

func (s *Service) AbortSingleUpload(ctx context.Context, userID, fileID string) error {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return ErrUnauthorized
	}
	if fileID == "" {
		return ErrInvalidInput
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	file, err := s.fileRepo.GetFileForUser(ctx, tx, fileID, userID)
	if err != nil {
		_ = tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if file.Status == FileStatusComplete || file.Status == FileStatusAborted || file.Status == FileStatusFailed {
		_ = tx.Rollback(ctx)
		return nil
	}

	updated, err := s.fileRepo.UpdateFileStatusIf(ctx, tx, fileID, FileStatusAborted, []string{FileStatusPending, FileStatusUploading})
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !updated {
		_ = tx.Rollback(ctx)
		return nil
	}

	released, err := s.storageRepo.ReleaseReservedStorage(ctx, tx, userID, file.SizeBytes)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !released {
		_ = tx.Rollback(ctx)
		return ErrUploadFailed
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	_ = s.r2.DeleteObject(ctx, file.ObjectKey)
	return nil
}

func (s *Service) AbortUploadByFile(ctx context.Context, userID, fileID string) error {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return ErrUnauthorized
	}
	if fileID == "" {
		return ErrInvalidInput
	}

	_, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	upload, err := s.uploadRepo.GetMultipartForFile(ctx, s.db, fileID, userID)
	if err == nil {
		return s.AbortMultipart(ctx, userID, upload.ID)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return s.AbortSingleUpload(ctx, userID, fileID)
	}
	return err
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
	if file.Status != FileStatusComplete {
		if file.Status == FileStatusFailed || file.Status == FileStatusAborted {
			return "", ErrUploadCancelled
		}
		return "", ErrNotFound
	}

	return s.r2.PresignDownload(ctx, file.ObjectKey, file.Filename, "attachment", s.downloadExpire)
}

func (s *Service) DeleteFile(ctx context.Context, userID, fileID string) error {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return ErrUnauthorized
	}
	if fileID == "" {
		return ErrInvalidInput
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	file, err := s.fileRepo.GetFileForUser(ctx, tx, fileID, userID)
	if err != nil {
		_ = tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	switch file.Status {
	case FileStatusComplete:
		if err := s.storageRepo.DecreaseUsedStorage(ctx, tx, userID, file.SizeBytes); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	case FileStatusPending, FileStatusUploading:
		_ = tx.Rollback(ctx)
		return ErrUploadCancelled
	}

	deleted, err := s.fileRepo.DeleteFileForUser(ctx, tx, fileID, userID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if !deleted {
		_ = tx.Rollback(ctx)
		return ErrNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	_ = s.r2.DeleteObject(ctx, file.ObjectKey)
	return nil
}

func (s *Service) ListPendingUploads(ctx context.Context, userID string) ([]models.File, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUnauthorized
	}
	return s.fileRepo.ListPendingForUser(ctx, s.db, userID)
}

func (s *Service) ListCompletedUploads(ctx context.Context, userID string) ([]models.File, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUnauthorized
	}
	return s.fileRepo.ListCompletedForUser(ctx, s.db, userID)
}

func (s *Service) ResumeMultipart(ctx context.Context, userID, fileID string) (models.MultipartResumeResponse, error) {
	userID = strings.TrimSpace(userID)
	fileID = strings.TrimSpace(fileID)
	if userID == "" {
		return models.MultipartResumeResponse{}, ErrUnauthorized
	}
	if fileID == "" {
		return models.MultipartResumeResponse{}, ErrInvalidInput
	}

	upload, err := s.uploadRepo.GetMultipartForFile(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.MultipartResumeResponse{}, ErrNotFound
		}
		return models.MultipartResumeResponse{}, err
	}
	if upload.Status == MultipartStatusCompleted || upload.Status == MultipartStatusAborted {
		return models.MultipartResumeResponse{}, ErrInvalidInput
	}

	file, err := s.fileRepo.GetFileForUser(ctx, s.db, fileID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.MultipartResumeResponse{}, ErrNotFound
		}
		return models.MultipartResumeResponse{}, err
	}

	if err := s.uploadRepo.TouchMultipart(ctx, s.db, upload.ID, MultipartStatusUploading); err != nil {
		return models.MultipartResumeResponse{}, err
	}

	listed, err := s.r2.ListParts(ctx, upload.ObjectKey, upload.UploadID)
	if err != nil {
		return models.MultipartResumeResponse{}, err
	}

	resumeParts := make([]models.ResumePart, 0, len(listed))
	for _, part := range listed {
		number := safeptr.Int32(part.PartNumber)
		etag := strings.TrimSpace(safeptr.String(part.ETag))
		if number <= 0 || number > int32(upload.TotalParts) || etag == "" {
			continue
		}
		resumeParts = append(resumeParts, models.ResumePart{
			PartNumber: number,
			ETag:       etag,
			Size:       safeptr.Int64(part.Size),
		})
	}

	return models.MultipartResumeResponse{
		FileID:        upload.FileID,
		MultipartID:   upload.ID,
		Filename:      file.Filename,
		SizeBytes:     file.SizeBytes,
		ChunkSize:     upload.ChunkSize,
		TotalParts:    upload.TotalParts,
		UploadedParts: resumeParts,
	}, nil
}
