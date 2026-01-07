package models

import "time"

type Share struct {
	ID           string
	FileID       string
	OwnerUserID  string
	Token        string
	PasswordHash *string
	ExpiresAt    *time.Time
	Status       string
	RevokedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ShareWithFile struct {
	Share
	FileName        string
	FileContentType string
	FileSizeBytes   int64
	FileUpdatedAt   time.Time
}
