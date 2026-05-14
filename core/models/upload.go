package models

import "time"

type UploadStartResponse struct {
	UploadID         string
	FileID           string
	VaultID          string
	Mode             string
	UploadURL        string
	UploadSessionID  string
	ProviderUploadID string
	FileChunkSize    int64
	TotalChunks      int
	UploadPartSize   int64
	UploadPartCount  int
}

type SingleStartResponse struct {
	FileID    string
	UploadURL string
	ExpiresAt time.Time
}

type UploadSession struct {
	ID               string
	FileID           string
	ProviderUploadID string
	UploadPartCount  int
	Status           string
	ExpiresAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UploadPart struct {
	ID              string
	UploadSessionID string
	PartNumber      int
	ETag            string
	EncryptedHash   []byte
	CreatedAt       time.Time
}
