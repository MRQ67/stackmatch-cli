package package_managers

import (
	"context"
	"fmt"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

type winget struct {
	*basePackageManager
}

// NewWinget creates a new Winget package manager instance
func NewWinget() types.Installer {
	return &winget{
		basePackageManager: &basePackageManager{
			name:           "Winget",
			pmType:        types.TypeWinget,
			executableName: "winget",
		},
	}
}

func (w *winget) InstallPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := w.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package with --silent for non-interactive installation
	_, err = w.runCommand(ctx, "install", "--silent", "--accept-package-agreements", "--accept-source-agreements", pkg)
	if err != nil {
		return fmt.Errorf("failed to install package: %w", err)
	}

	return nil
}

func (w *winget) InstallMultiple(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// Winget doesn't support installing multiple packages in one command,
	// so we install them one by one
	for _, pkg := range packages {
		err := w.InstallPackage(ctx, pkg)
		if err != nil {
			// Check if the error is PackageAlreadyInstalledError
			if _, ok := err.(*types.PackageAlreadyInstalledError); !ok {
				return fmt.Errorf("failed to install package %s: %w", pkg, err)
			}
		}
	}

	return nil
}

func (w *winget) UpdatePackageManager(ctx context.Context) error {
	// Update winget itself
	_, err := w.runCommand(ctx, "--version")
	if err != nil {
		return fmt.Errorf("failed to check winget version: %w", err)
	}

	// Update all installed packages
	_, err = w.runCommand(ctx, "upgrade", "--all", "--silent", "--accept-package-agreements", "--accept-source-agreements")
	if err != nil {
		return fmt.Errorf("failed to update packages: %w", err)
	}

	return nil
}

// checkIfInstalled overrides the base implementation with Winget-specific logic
func (w *winget) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	output, err := w.runCommand(ctx, "list", "--name", pkg)
	if err != nil {
		return false, nil
	}

	// Check if the package appears in the list of installed packages
	lines := strings.Split(output, "\n")
	if len(lines) > 2 { // Header + separator + at least one package
		// Check if any line (after header and separator) contains the package
		for _, line := range lines[2:] {
			if strings.Contains(strings.ToLower(line), strings.ToLower(pkg)) {
				return true, nil
			}
		}
	}

	return false, nil
}
