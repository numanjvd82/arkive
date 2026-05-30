package models

import "time"

type File struct {
	ID                  string
	UserID              string
	FolderID            *string
	EncryptedMetadata   []byte
	EncryptedFileKey    []byte
	EncryptedManifest   []byte
	EncryptionVersion   int16
	ChunkSize           int64
	ChunkCount          int
	PlaintextSize       int64
	ActualEncryptedSize int64
	EncryptedHash       []byte
	UploadStatus        string
	ThumbnailStatus     string
	ThumbnailSizeBytes  int64
	ThumbnailMime       string
	ThumbnailWidth      int
	ThumbnailHeight     int
	CompletedAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ExpiresAt           *time.Time
	SearchScore         int64
}
