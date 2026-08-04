package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds everything Datara needs to start.
type Config struct {
	PostgresDSN string
	MaxRows     int
}

// LoadFromEnv reads configuration from environment variables:
//   - DATARA_POSTGRES_DSN (required)
//   - DATARA_MAX_ROWS (optional, default 1000)
func LoadFromEnv() (Config, error) {
	dsn := os.Getenv("DATARA_POSTGRES_DSN")
	if dsn == "" {
		return Config{}, fmt.Errorf("DATARA_POSTGRES_DSN is required")
	}

	maxRows := 1000
	if raw := os.Getenv("DATARA_MAX_ROWS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("invalid DATARA_MAX_ROWS: %w", err)
		}
		maxRows = parsed
	}

	return Config{PostgresDSN: dsn, MaxRows: maxRows}, nil
}