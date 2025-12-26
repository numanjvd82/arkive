package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	DatabaseURL string
	Port        string
	SessionTTL  time.Duration
	Env         string
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

	return Config{
		DatabaseURL: dsn,
		Port:        addr,
		SessionTTL:  sessionTTL,
		Env:         env,
	}, nil
}

func parseDurationEnv(key string) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, errors.New(key + " is required")
	}
	return time.ParseDuration(value)
}
