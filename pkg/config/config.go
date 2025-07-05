package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

var (
	// These variables are set at build time using ldflags
	SupabaseURL    string
	SupabaseAPIKey string
)

// Config holds the application configuration
type Config struct {
	SupabaseURL    string `json:"supabase_url,omitempty"`
	SupabaseAPIKey string `json:"supabase_key,omitempty"`
	configPath     string `json:"-"` // Path to config file, not serialized
}

// New creates a new configuration with values from environment variables and config file
func New() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Get config directory
	configDir, _ := os.UserConfigDir()
	if configDir == "" {
		configDir = "."
	}

	// Ensure stackmatch directory exists
	configDir = filepath.Join(configDir, "stackmatch")
	_ = os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "config.json")
	cfg := &Config{
		configPath: configPath,
	}

	// Try to load existing config
	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, cfg)
	}

	// Override with build-time variables if available
	if SupabaseURL != "" {
		cfg.SupabaseURL = SupabaseURL
	}
	if SupabaseAPIKey != "" {
		cfg.SupabaseAPIKey = SupabaseAPIKey
	}

	// Fallback to environment variables if not set by build flags or config
	if cfg.SupabaseURL == "" {
		cfg.SupabaseURL = os.Getenv("SUPABASE_URL")
	}
	if cfg.SupabaseAPIKey == "" {
		cfg.SupabaseAPIKey = os.Getenv("SUPABASE_ANON_KEY")
	}

	return cfg
}


// Save writes the configuration to disk
func (c *Config) Save() error {
	if c.configPath == "" {
		return nil // Skip if no config path is set (e.g., in tests)
	}

	// Don't save sensitive information
	saveCfg := *c
	saveCfg.SupabaseAPIKey = ""

	data, err := json.MarshalIndent(saveCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
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

// IsConfigured returns true if the required configuration is present
func (c *Config) IsConfigured() bool {
	return c.SupabaseURL != "" && c.SupabaseAPIKey != ""
}

// BindFlags binds command-line flags to the configuration
func (c *Config) BindFlags(flags *pflag.FlagSet) error {
	if flags.Changed("supabase-url") {
		url, err := flags.GetString("supabase-url")
		if err != nil {
			return fmt.Errorf("failed to get supabase-url flag: %w", err)
		}
		c.SupabaseURL = url
	}

	if flags.Changed("supabase-key") {
		key, err := flags.GetString("supabase-key")
		if err != nil {
			return fmt.Errorf("failed to get supabase-key flag: %w", err)
		}
		c.SupabaseAPIKey = key
	}

	// Save the updated configuration
	if err := c.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
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
