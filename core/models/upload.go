package models

import "time"

type UploadStartResponse struct {
	UploadID  string
	FileID    string
	ObjectKey string
	Mode      string
	UploadURL string
}

type SingleStartResponse struct {
	FileID    string
	ObjectKey string
	UploadURL string
	ExpiresAt time.Time
}
