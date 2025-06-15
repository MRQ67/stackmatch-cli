package package_managers

import (
	"context"
	"fmt"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

type yum struct {
	*basePackageManager
}

// NewYum creates a new YUM package manager instance
func NewYum() types.Installer {
	return &yum{
		basePackageManager: &basePackageManager{
			name:           "YUM",
			pmType:        types.TypeYum,
			executableName: "yum",
		},
	}
}

func (y *yum) InstallPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := y.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package with -y to assume yes
	_, err = y.runCommand(ctx, "install", "-y", pkg)
	if err != nil {
		return fmt.Errorf("failed to install package: %w", err)
	}

	return nil
}

func (y *yum) InstallMultiple(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// YUM can install multiple packages in one command
	args := append([]string{"install", "-y"}, packages...)
	_, err := y.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

func (y *yum) UpdatePackageManager(ctx context.Context) error {
	// Update all packages
	_, err := y.runCommand(ctx, "update", "-y")
	if err != nil {
		return fmt.Errorf("failed to update packages: %w", err)
	}

	return nil
}

// checkIfInstalled overrides the base implementation with YUM-specific logic
func (y *yum) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	output, err := y.runCommand(ctx, "list", "installed", pkg)
	if err != nil {
		return false, nil
	}

	// Check if the package appears in the installed packages list
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 1 {
		// First line is header, check subsequent lines
		for _, line := range lines[1:] {
			if strings.HasPrefix(strings.Fields(line)[0], pkg) {
				return true, nil
			}
		}
	}

	return false, nil
}
