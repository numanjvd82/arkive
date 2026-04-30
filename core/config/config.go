package config

import (
	"errors"
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

	PublicBaseURL       string
	PostmarkServerToken string
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

	postmarkServerToken := strings.TrimSpace(os.Getenv("POSTMARK_SERVER_TOKEN"))
	if env != "dev" {
		if postmarkServerToken == "" {
			return Config{}, errors.New("POSTMARK_SERVER_TOKEN is required")
		}
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

		PublicBaseURL:       publicBaseURL,
		PostmarkServerToken: postmarkServerToken,
	}, nil
}

func parseDurationEnv(key string) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, errors.New(key + " is required")
	}
	return time.ParseDuration(value)
}
