package package_managers

import (
	"context"
	"fmt"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

type snap struct {
	*basePackageManager
}

// NewSnap creates a new Snap package manager instance
func NewSnap() types.Installer {
	return &snap{
		basePackageManager: &basePackageManager{
			name:           "Snap",
			pmType:        types.TypeSnap,
			executableName: "snap",
		},
	}
}

func (s *snap) InstallPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := s.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package with --classic for classic confinement if needed
	_, err = s.runCommand(ctx, "install", "--classic", pkg)
	if err != nil {
		// Try without --classic if that fails
		_, err = s.runCommand(ctx, "install", pkg)
		if err != nil {
			return fmt.Errorf("failed to install package: %w", err)
		}
	}

	return nil
}

func (s *snap) InstallMultiple(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// Snap can install multiple packages in one command
	args := append([]string{"install"}, packages...)
	_, err := s.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

func (s *snap) UpdatePackageManager(ctx context.Context) error {
	// Update all snaps
	_, err := s.runCommand(ctx, "refresh")
	if err != nil {
		return fmt.Errorf("failed to update snaps: %w", err)
	}

	return nil
}

// checkIfInstalled overrides the base implementation with Snap-specific logic
func (s *snap) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	output, err := s.runCommand(ctx, "list", pkg)
	if err != nil {
		return false, nil
	}

	// Check if the package appears in the list of installed packages
	lines := strings.Split(strings.TrimSpace(output), "\n")
	return len(lines) > 1, nil // First line is header
}
