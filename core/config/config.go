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
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2SessionToken    string
	R2Bucket          string
	R2Endpoint        string
	R2Region          string

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

	r2AccessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	if r2AccessKeyID == "" {
		return Config{}, errors.New("R2_ACCESS_KEY_ID is required")
	}

	r2SecretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	if r2SecretAccessKey == "" {
		return Config{}, errors.New("R2_SECRET_ACCESS_KEY is required")
	}

	r2Bucket := os.Getenv("R2_BUCKET")
	if r2Bucket == "" {
		return Config{}, errors.New("R2_BUCKET is required")
	}

	r2Endpoint := os.Getenv("R2_ENDPOINT")
	if r2Endpoint == "" {
		return Config{}, errors.New("R2_ENDPOINT is required")
	}

	r2Region := os.Getenv("R2_REGION")
	if r2Region == "" {
		r2Region = "auto"
	}

	publicBaseURL := strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL"))
	if publicBaseURL == "" {
		publicBaseURL = "https://arkive.sh"
	}

	postmarkServerToken := strings.TrimSpace(os.Getenv("POSTMARK_SERVER_TOKEN"))
	// Email verification can be skipped only in dev.
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
		R2AccessKeyID:     r2AccessKeyID,
		R2SecretAccessKey: r2SecretAccessKey,
		R2SessionToken:    os.Getenv("R2_SESSION_TOKEN"),
		R2Bucket:          r2Bucket,
		R2Endpoint:        r2Endpoint,
		R2Region:          r2Region,

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
