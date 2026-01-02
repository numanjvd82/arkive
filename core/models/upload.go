package models

import "time"

type MultipartStartResponse struct {
	FileID      string
	MultipartID string
	ObjectKey   string
	ChunkSize   int
	TotalParts  int
}

type UploadStartResponse struct {
	UploadID   string
	FileID     string
	ObjectKey  string
	Mode       string
	ChunkSize  int
	TotalParts int
	UploadURL  string
}

type UploadNextResponse struct {
	UploadID      string
	FileID        string
	Mode          string
	NextPart      int32
	URL           string
	ChunkSize     int
	TotalParts    int
	UploadedParts []ResumePart
	ThrottleMs    int
}

type SingleStartResponse struct {
	FileID    string
	ObjectKey string
	UploadURL string
	ExpiresAt time.Time
}

type CompletedPartInput struct {
	PartNumber int32  `json:"partNumber"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size,omitempty"`
}

type StoredPart struct {
	PartNumber int32  `json:"part"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
}

type ResumePart struct {
	PartNumber int32  `json:"partNumber"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
}

type MultipartResumeResponse struct {
	FileID        string       `json:"fileId"`
	MultipartID   string       `json:"multipartId"`
	Filename      string       `json:"filename"`
	SizeBytes     int64        `json:"sizeBytes"`
	ChunkSize     int          `json:"chunkSize"`
	TotalParts    int          `json:"totalParts"`
	UploadedParts []ResumePart `json:"uploadedParts"`
}
