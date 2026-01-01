package models

import "time"

type MultipartUpload struct {
	ID            string
	FileID        string
	UploadID      string
	Bucket        string
	ObjectKey     string
	ChunkSize     int
	TotalParts    int
	UploadedParts []byte
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpiresAt     time.Time
}
