package models

type StorageSettings struct {
	Provider          string
	LocalPath         string
	MaxStorageBytes   int64
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3SessionToken    string
	S3Bucket          string
	S3Endpoint        string
	S3Region          string
	S3UsePathStyle    bool
}

type EmailSettings struct {
	Provider      string
	From          string
	PublicBaseURL string
	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPass      string
}

type UploadSettings struct {
	MaxUploadConcurrency int
	MaxQueueItems        int
}
