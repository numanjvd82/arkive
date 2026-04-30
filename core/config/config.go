package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL       string
	Port              string
	SessionTTL        time.Duration
	Env               string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3SessionToken    string
	S3Bucket          string
	S3Endpoint        string
	S3Region          string

	PublicBaseURL        string
	MaxFileSizeBytes     int64
	MaxUploadConcurrency int
	MaxQueueItems        int
	EmailProvider string
	EmailFrom     string
	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPass      string
}

func Load() (Config, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	sessionTTL, err := parseDurationEnv("SESSION_TTL")
	if err != nil {
		return Config{}, err
	}

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "prod"
	}

	s3AccessKeyID := os.Getenv("S3_ACCESS_KEY_ID")
	if s3AccessKeyID == "" {
		return Config{}, errors.New("S3_ACCESS_KEY_ID is required")
	}

	s3SecretAccessKey := os.Getenv("S3_SECRET_ACCESS_KEY")
	if s3SecretAccessKey == "" {
		return Config{}, errors.New("S3_SECRET_ACCESS_KEY is required")
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		return Config{}, errors.New("S3_BUCKET is required")
	}

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint == "" {
		return Config{}, errors.New("S3_ENDPOINT is required")
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		s3Region = "auto"
	}

	publicBaseURL := strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL"))
	if publicBaseURL == "" {
		publicBaseURL = "https://arkive.sh"
	}

	emailProvider := strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_PROVIDER")))
	if emailProvider == "" {
		emailProvider = "noop"
	}

	emailFrom := strings.TrimSpace(os.Getenv("EMAIL_FROM"))
	if emailFrom == "" {
		emailFrom = "noreply@arkive.sh"
	}

	smtpHost := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	smtpPort := 587
	if v := os.Getenv("SMTP_PORT"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			smtpPort = n
		}
	}
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	maxFileSizeBytes := int64(10 * 1024 * 1024 * 1024)
	if v := os.Getenv("MAX_FILE_SIZE_BYTES"); v != "" {
		n, err := parseInt64(v)
		if err != nil || n <= 0 {
			return Config{}, errors.New("MAX_FILE_SIZE_BYTES must be a positive integer")
		}
		maxFileSizeBytes = n
	}

	maxUploadConcurrency := 10
	if v := os.Getenv("MAX_UPLOAD_CONCURRENCY"); v != "" {
		n, err := parseInt(v)
		if err != nil || n <= 0 {
			return Config{}, errors.New("MAX_UPLOAD_CONCURRENCY must be a positive integer")
		}
		maxUploadConcurrency = n
	}

	maxQueueItems := 300
	if v := os.Getenv("MAX_QUEUE_ITEMS"); v != "" {
		n, err := parseInt(v)
		if err != nil || n <= 0 {
			return Config{}, errors.New("MAX_QUEUE_ITEMS must be a positive integer")
		}
		maxQueueItems = n
	}

	return Config{
		DatabaseURL:       dsn,
		Port:              addr,
		SessionTTL:        sessionTTL,
		Env:               env,
		S3AccessKeyID:     s3AccessKeyID,
		S3SecretAccessKey: s3SecretAccessKey,
		S3SessionToken:    os.Getenv("S3_SESSION_TOKEN"),
		S3Bucket:          s3Bucket,
		S3Endpoint:        s3Endpoint,
		S3Region:          s3Region,

		PublicBaseURL:        publicBaseURL,
		MaxFileSizeBytes:     maxFileSizeBytes,
		MaxUploadConcurrency: maxUploadConcurrency,
		MaxQueueItems:        maxQueueItems,
		EmailProvider:        emailProvider,
		EmailFrom:            emailFrom,
		SMTPHost:             smtpHost,
		SMTPPort:             smtpPort,
		SMTPUser: smtpUser,
		SMTPPass: smtpPass,
	}, nil
}

func parseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func parseDurationEnv(key string) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, errors.New(key + " is required")
	}
	return time.ParseDuration(value)
}
