package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
	SessionTTL  time.Duration
	Env         string
}

func Load() (Config, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}

	accessTTL, err := parseDurationEnv("ACCESS_TOKEN_TTL", "15m")
	if err != nil {
		return Config{}, err
	}
	refreshTTL, err := parseDurationEnv("REFRESH_TOKEN_TTL", "720h")
	if err != nil {
		return Config{}, err
	}
	sessionTTL, err := parseDurationEnv("SESSION_TTL", "168h")
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

	return Config{
		DatabaseURL: dsn,
		Port:        addr,
		JWTSecret:   jwtSecret,
		AccessTTL:   accessTTL,
		RefreshTTL:  refreshTTL,
		SessionTTL:  sessionTTL,
		Env:         env,
	}, nil
}

func parseDurationEnv(key, fallback string) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}
	return time.ParseDuration(value)
}
