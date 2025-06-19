package config

import "sync"

var (
	configInstance *Config
	configOnce     sync.Once
)

// GetConfig returns the singleton instance of Config
func GetConfig() *Config {
	configOnce.Do(func() {
		configInstance = New()
	})
	return configInstance
}
