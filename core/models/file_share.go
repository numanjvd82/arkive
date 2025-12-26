package models

import "time"

type FileShare struct {
	ID        string
	FileID    string
	TokenHash []byte
	ExpiresAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
