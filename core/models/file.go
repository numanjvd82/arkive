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
	ThrottleMs           int
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ExpiresAt            time.Time
}
