package package_managers

import (
	"context"
	"fmt"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

type homebrew struct {
	*basePackageManager
}

// NewHomebrew creates a new Homebrew package manager instance
func NewHomebrew() types.Installer {
	return &homebrew{
		basePackageManager: &basePackageManager{
			name:           "Homebrew",
			pmType:        types.TypeHomebrew,
			executableName: "brew",
		},
	}
}

func (h *homebrew) InstallPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := h.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package
	_, err = h.runCommand(ctx, "install", pkg)
	if err != nil {
		// Check if package was not found
		if strings.Contains(err.Error(), "No available formula or cask") {
			return &types.PackageNotFoundError{
				Package: pkg,
			}
		}
		return fmt.Errorf("failed to install package: %w", err)
	}

	return nil
}

func (h *homebrew) InstallMultiple(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// Homebrew can install multiple packages in one command
	args := append([]string{"install"}, packages...)
	_, err := h.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

func (h *homebrew) UpdatePackageManager(ctx context.Context) error {
	_, err := h.runCommand(ctx, "update")
	if err != nil {
		return fmt.Errorf("failed to update Homebrew: %w", err)
	}
	return nil
}

// checkIfInstalled overrides the base implementation with Homebrew-specific logic
func (h *homebrew) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	output, err := h.runCommand(ctx, "list", "--versions", pkg)
	if err != nil {
		// If the package is not installed, list will return an error
		if strings.Contains(err.Error(), "No available formula or cask") ||
			strings.Contains(err.Error(), "No such keg") {
			return false, nil
		}
		return false, err
	}

	// If we get output, the package is installed
	return strings.TrimSpace(output) != "", nil
}
