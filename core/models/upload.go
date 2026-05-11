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
	PartSize         int64
	TotalParts       int
}

type SingleStartResponse struct {
	FileID    string
	UploadURL string
	ExpiresAt time.Time
}

type UploadSession struct {
	ID               string
	FileID           string
	StorageKey       string
	ProviderUploadID string
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
