package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	SupabaseURL    string
	SupabaseAPIKey string
}

// New creates a new configuration with values from environment variables
func New() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_ANON_KEY")

	log.Printf("Debug - Supabase URL: %s", url)
	log.Printf("Debug - Supabase Key: %s... (first 10 chars)", safeSubstring(key, 0, 10))

	return &Config{
		SupabaseURL:    url,
		SupabaseAPIKey: key,
	}
}

// safeSubstring returns a substring of s from start to end, handling edge cases
func safeSubstring(s string, start, end int) string {
	if s == "" {
		return ""
	}
	if start > len(s) {
		start = len(s)
	}
	if end > len(s) {
		end = len(s)
	}
	if start > end {
		start = end
	}
	return s[start:end]
}

// Validate checks if the required configuration values are set
func (c *Config) Validate() error {
	if c.SupabaseURL == "" {
		return ErrMissingSupabaseURL
	}
	if c.SupabaseAPIKey == "" {
		return ErrMissingSupabaseAPIKey
	}
	return nil
}

// Error definitions
var (
	ErrMissingSupabaseURL    = newConfigError("missing Supabase URL")
	ErrMissingSupabaseAPIKey = newConfigError("missing Supabase API key")
)

type configError struct {
	msg string
}

func newConfigError(msg string) *configError {
	return &configError{msg: msg}
}

func (e *configError) Error() string {
	return "config error: " + e.msg
}
