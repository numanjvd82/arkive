package storage

import (
	"context"
	"time"
)

type Provider interface {
	PresignUpload(ctx context.Context, key, contentType string, expires time.Duration) (string, error)
	PresignDownload(ctx context.Context, key, filename, disposition string, expires time.Duration) (string, error)
	DeleteObject(ctx context.Context, key string) error
}
