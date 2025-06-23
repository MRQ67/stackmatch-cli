package installer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

// InstallationTracker tracks package installations and supports rollback
type InstallationTracker struct {
	installations map[string]*InstallationRecord
	mu            sync.Mutex
	trackerFile  string
}

// InstallationRecord represents a single installation record
type InstallationRecord struct {
	ID          string                     `json:"id"`
	Timestamp   time.Time                  `json:"timestamp"`
	Packages    map[string]PackageInfo     `json:"packages"`
	Environment *types.EnvironmentData     `json:"environment,omitempty"`
	Metadata    map[string]string         `json:"metadata,omitempty"`
	Status      InstallationStatus         `json:"status"`
}

// PackageInfo contains information about an installed package
type PackageInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	ManagerType string `json:"manager_type"`
}

// InstallationStatus represents the status of an installation
type InstallationStatus string

const (
	// StatusInProgress indicates the installation is in progress
	StatusInProgress InstallationStatus = "in_progress"
	// StatusCompleted indicates the installation completed successfully
	StatusCompleted InstallationStatus = "completed"
	// StatusFailed indicates the installation failed
	StatusFailed InstallationStatus = "failed"
	// StatusRolledBack indicates the installation was rolled back
	StatusRolledBack InstallationStatus = "rolled_back"
)

// NewInstallationTracker creates a new InstallationTracker
func NewInstallationTracker(trackerFile string) (*InstallationTracker, error) {
	tracker := &InstallationTracker{
		installations: make(map[string]*InstallationRecord),
		trackerFile:   trackerFile,
	}

	// Load existing records if the tracker file exists
	if _, err := os.Stat(trackerFile); err == nil {
		if err := tracker.load(); err != nil {
			return nil, fmt.Errorf("failed to load tracker file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error checking tracker file: %w", err)
	}

	return tracker, nil
}

// StartInstallation starts tracking a new installation
func (t *InstallationTracker) StartInstallation(env *types.EnvironmentData) (*InstallationRecord, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	record := &InstallationRecord{
		ID:          generateID(),
		Timestamp:   time.Now(),
		Packages:    make(map[string]PackageInfo),
		Environment: env,
		Metadata:    make(map[string]string),
		Status:      StatusInProgress,
	}


	t.installations[record.ID] = record

	if err := t.save(); err != nil {
		delete(t.installations, record.ID)
		return nil, fmt.Errorf("failed to save installation record: %w", err)
	}

	return record, nil
}

// AddPackage adds a package to an installation record
func (t *InstallationTracker) AddPackage(installationID string, pkg types.PackageInfo) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	record, exists := t.installations[installationID]
	if !exists {
		return fmt.Errorf("installation record not found: %s", installationID)
	}

	// Convert types.PackageInfo to installer.PackageInfo
	record.Packages[pkg.Name] = PackageInfo{
		Name:        pkg.Name,
		Version:     pkg.Version,
		ManagerType: "", // Manager type is not available in types.PackageInfo
	}

	return t.save()
}

// CompleteInstallation marks an installation as completed
func (t *InstallationTracker) CompleteInstallation(installationID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	record, exists := t.installations[installationID]
	if !exists {
		return fmt.Errorf("installation record not found: %s", installationID)
	}

	record.Status = StatusCompleted
	record.Timestamp = time.Now()

	return t.save()
}

// FailInstallation marks an installation as failed
func (t *InstallationTracker) FailInstallation(installationID string, reason string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	record, exists := t.installations[installationID]
	if !exists {
		return fmt.Errorf("installation record not found: %s", installationID)
	}

	record.Status = StatusFailed
	record.Metadata["failure_reason"] = reason
	record.Timestamp = time.Now()

	return t.save()
}

// Rollback rolls back an installation by uninstalling all installed packages
func (t *InstallationTracker) Rollback(ctx context.Context, installationID string, manager types.Installer) error {
	t.mu.Lock()
	record, exists := t.installations[installationID]
	if !exists {
		t.mu.Unlock()
		return fmt.Errorf("installation record not found: %s", installationID)
	}

	// Mark as rolling back
	record.Status = "rolling_back"
	if err := t.save(); err != nil {
		t.mu.Unlock()
		return fmt.Errorf("failed to update installation status: %w", err)
	}
	t.mu.Unlock()

	// Rollback packages in reverse order
	var rollbackErr error
	for _, pkg := range record.Packages {
		if err := manager.UninstallPackage(ctx, pkg.Name); err != nil {
			// Log the error but continue with other packages
			if rollbackErr == nil {
				rollbackErr = fmt.Errorf("failed to uninstall package %s: %w", pkg.Name, err)
			} else {
				rollbackErr = fmt.Errorf("%w; failed to uninstall package %s: %v", rollbackErr, pkg.Name, err)
			}
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if rollbackErr != nil {
		record.Status = "rollback_failed"
		record.Metadata["rollback_error"] = rollbackErr.Error()
	} else {
		record.Status = StatusRolledBack
	}

	record.Timestamp = time.Now()

	if err := t.save(); err != nil {
		if rollbackErr != nil {
			return fmt.Errorf("rollback failed: %w; failed to save record: %v", rollbackErr, err)
		}
		return fmt.Errorf("failed to save rollback record: %w", err)
	}

	return rollbackErr
}

// GetInstallation returns an installation record by ID
func (t *InstallationTracker) GetInstallation(id string) (*InstallationRecord, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	record, exists := t.installations[id]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	recordCopy := *record
	return &recordCopy, true
}

// ListInstallations returns all installation records
func (t *InstallationTracker) ListInstallations() []InstallationRecord {
	t.mu.Lock()
	defer t.mu.Unlock()

	var records []InstallationRecord
	for _, record := range t.installations {
		records = append(records, *record)
	}

	return records
}

// save saves the installation records to disk
func (t *InstallationTracker) save() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(t.trackerFile), 0755); err != nil {
		return fmt.Errorf("failed to create tracker directory: %w", err)
	}

	file, err := os.Create(t.trackerFile)
	if err != nil {
		return fmt.Errorf("failed to create tracker file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(t.installations); err != nil {
		return fmt.Errorf("failed to encode installation records: %w", err)
	}

	return nil
}

// load loads the installation records from disk
func (t *InstallationTracker) load() error {
	file, err := os.Open(t.trackerFile)
	if err != nil {
		return fmt.Errorf("failed to open tracker file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&t.installations); err != nil {
		return fmt.Errorf("failed to decode installation records: %w", err)
	}

	return nil
}

// generateID generates a unique ID for an installation record
func generateID() string {
	return fmt.Sprintf("inst_%d", time.Now().UnixNano())
}
