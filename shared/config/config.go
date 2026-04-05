package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Base struct {
	HTTPPort        string
	ShutdownTimeout time.Duration
	LogLevel        string
	LogFormat       string
	Environment     string
}

func GetEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func GetEnvInt(key string, defaultVal int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("env %s: expected integer, got %q", key, v)
	}
	return n, nil
}

func GetEnvDuration(key string, defaultVal time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("env %s: invalid duration %q (e.g. 5s, 1m, 500ms)", key, v)
	}
	return d, nil
}

var (
	validLogLevels    = map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	validLogFormats   = map[string]bool{"text": true, "json": true}
	validEnvironments = map[string]bool{"dev": true, "prod": true}
)

func ValidateBase(b *Base) error {
	if b.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT is required")
	}
	if !validLogLevels[b.LogLevel] {
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error (got %q)", b.LogLevel)
	}
	if !validLogFormats[b.LogFormat] {
		return fmt.Errorf("LOG_FORMAT must be 'text' or 'json' (got %q)", b.LogFormat)
	}
	if !validEnvironments[b.Environment] {
		return fmt.Errorf("ENVIRONMENT must be 'dev' or 'prod' (got %q)", b.Environment)
	}
	return nil
}

func LoadBase(defaultPort string) (Base, error) {
	shutdownTimeout, err := GetEnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	if err != nil {
		return Base{}, err
	}
	return Base{
		HTTPPort:        GetEnv("HTTP_PORT", defaultPort),
		ShutdownTimeout: shutdownTimeout,
		LogLevel:        GetEnv("LOG_LEVEL", "info"),
		LogFormat:       GetEnv("LOG_FORMAT", "text"),
		Environment:     GetEnv("ENVIRONMENT", "dev"),
	}, nil
}

// MaskPassword скрывает пароль в DSN: postgres://user:secret@host → postgres://user:***@host
func MaskPassword(dsn string) string {
	colonCount := 0
	passwordStart := -1
	for i, c := range dsn {
		if c == ':' {
			colonCount++
			if colonCount == 2 {
				passwordStart = i + 1
			}
		}
		if c == '@' && passwordStart != -1 {
			return dsn[:passwordStart] + "***" + dsn[i:]
		}
	}
	return dsn
}
