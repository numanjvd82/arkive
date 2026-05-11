package storage

import (
	"context"
	"time"
)

type CompletedPart struct {
	PartNumber int32
	ETag       string
}

type Provider interface {
	ActiveProvider(ctx context.Context) (string, error)
	PresignUpload(ctx context.Context, key, contentType string, expires time.Duration) (string, error)
	PresignDownload(ctx context.Context, key, filename, disposition string, expires time.Duration) (string, error)
	ObjectSize(ctx context.Context, key string) (int64, error)
	CreateMultipartUpload(ctx context.Context, key, contentType string) (string, error)
	PresignUploadPart(ctx context.Context, key, uploadID string, partNumber int32, expires time.Duration) (string, error)
	CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []CompletedPart) error
	AbortMultipartUpload(ctx context.Context, key, uploadID string) error
	DeleteObject(ctx context.Context, key string) error
}
