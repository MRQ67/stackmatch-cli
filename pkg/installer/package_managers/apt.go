package package_managers

import (
	"context"
	"fmt"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

type apt struct {
	*basePackageManager
}

// NewApt creates a new APT package manager instance
func NewApt() types.Installer {
	return &apt{
		basePackageManager: &basePackageManager{
			name:           "APT",
			pmType:        types.TypeApt,
			executableName: "apt",
		},
	}
}

func (a *apt) InstallPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := a.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package with --assume-yes to avoid prompts
	_, err = a.runCommand(ctx, "install", "--assume-yes", pkg)
	if err != nil {
		return fmt.Errorf("failed to install package: %w", err)
	}

	return nil
}

func (a *apt) InstallMultiple(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// APT can install multiple packages in one command
	args := append([]string{"install", "--assume-yes"}, packages...)
	_, err := a.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

func (a *apt) UpdatePackageManager(ctx context.Context) error {
	// Update package lists
	_, err := a.runCommand(ctx, "update")
	if err != nil {
		return fmt.Errorf("failed to update package lists: %w", err)
	}

	// Upgrade all packages
	_, err = a.runCommand(ctx, "upgrade", "--assume-yes")
	if err != nil {
		return fmt.Errorf("failed to upgrade packages: %w", err)
	}

	return nil
}

// checkIfInstalled overrides the base implementation with APT-specific logic
func (a *apt) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	// dpkg -s returns 0 if package is installed
	_, err := a.runCommand(ctx, "dpkg", "-s", pkg)
	if err == nil {
		return true, nil
	}

	// Check if the error is because the package is not installed
	output, _ := a.runCommand(ctx, "dpkg", "-l", pkg)
	if strings.Contains(output, pkg) && strings.Contains(output, "ii") {
		return true, nil
	}

	return false, nil
}
