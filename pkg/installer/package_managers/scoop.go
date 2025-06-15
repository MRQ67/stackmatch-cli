package package_managers

import (
	"context"
	"fmt"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

type scoop struct {
	*basePackageManager
}

// NewScoop creates a new Scoop package manager instance
func NewScoop() types.Installer {
	return &scoop{
		basePackageManager: &basePackageManager{
			name:           "Scoop",
			pmType:        types.TypeScoop,
			executableName: "scoop",
		},
	}
}

func (s *scoop) InstallPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := s.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package
	_, err = s.runCommand(ctx, "install", pkg)
	if err != nil {
		return fmt.Errorf("failed to install package: %w", err)
	}

	return nil
}

func (s *scoop) InstallMultiple(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// Scoop can install multiple packages in one command
	args := append([]string{"install"}, packages...)
	_, err := s.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

func (s *scoop) UpdatePackageManager(ctx context.Context) error {
	// Update scoop itself
	_, err := s.runCommand(ctx, "update")
	if err != nil {
		return fmt.Errorf("failed to update scoop: %w", err)
	}

	// Update all installed packages
	_, err = s.runCommand(ctx, "update", "*")
	if err != nil {
		return fmt.Errorf("failed to update packages: %w", err)
	}

	return nil
}

// checkIfInstalled overrides the base implementation with Scoop-specific logic
func (s *scoop) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	output, err := s.runCommand(ctx, "list")
	if err != nil {
		return false, err
	}

	// Check if the package appears in the list of installed packages
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == pkg {
			return true, nil
		}
	}

	return false, nil
}
