package types

import "time"

// EnvironmentData represents the top-level structure for the scanned environment.
// This is the structure that will be serialized to/from JSON.
type EnvironmentData struct {
	StackmatchVersion string            `json:"stackmatch_version"`
	ScanDate          time.Time         `json:"scan_date"`
	System            SystemInfo        `json:"system"`
	Tools             map[string]string `json:"tools,omitempty"`
	PackageManagers   map[string]string `json:"package_managers,omitempty"`
	CodeEditors       map[string]string `json:"code_editors,omitempty"`
	// ConfiguredLanguages stores detected programming languages and their primary versions.
	ConfiguredLanguages map[string]string `json:"configured_languages,omitempty"`
	ConfigFiles         []string          `json:"config_files,omitempty"`
}

// SystemInfo holds basic information about the operating system and architecture.
type SystemInfo struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	Shell       string `json:"shell,omitempty"`
	Hostname    string `json:"hostname,omitempty"` // Added Hostname as it's often useful
}

// EnvironmentHistory represents a version history entry for an environment
type EnvironmentHistory struct {
	ID            string    `json:"id"`
	EnvironmentID string    `json:"environment_id"`
	Name          string    `json:"name"`
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedBy     string    `json:"updated_by"`
	Data          string    `json:"data"` // JSON string of the environment data
}

// Environment is a struct that holds the environment data, name, and username
type Environment struct {
	Name     string            `json:"name"`
	Username string            `json:"username"`
	Data     EnvironmentData `json:"data"`
}

// Tool represents a detected development tool.
// We might expand this later if more structured info per tool is needed.
// For now, a simple map[string]string in EnvironmentData.Tools is used.

// Note: The MVP focuses on detection. More detailed structures for each category
// (e.g., specific fields for Docker, Git, Node versions) can be added later
// if deeper inspection or specific version parsing is required beyond a string.
