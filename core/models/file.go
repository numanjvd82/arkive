package models

import "time"

type File struct {
	ID                   string
	UserID               string
	Bucket               string
	ObjectKey            string
	Filename             string
	ContentType          string
	SizeBytes            int64
	VideoWidth           int
	VideoHeight          int
	VideoDurationSeconds int64
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ExpiresAt            *time.Time

	EncryptedMetadata []byte
	EncryptedFileKey  []byte
	EncryptedManifest []byte
	EncryptionVersion int16
	ChunkSize         int64
	ChunkCount        int
	PlaintextSize     int64
	EncryptedSize     *int64
	EncryptedHash     []byte
	UploadStatus      string
	StorageBackend    string
	CompletedAt       *time.Time
}

type FileChunk struct {
	ID            string
	FileID        string
	ChunkIndex    int
	StorageKey    string
	PlaintextSize int64
	EncryptedSize int64
	EncryptedHash []byte
	UploadStatus  string
	UploadedAt    *time.Time
	CreatedAt     time.Time
}
