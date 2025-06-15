package package_managers

import (
	"context"
	"fmt"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

type dnf struct {
	*basePackageManager
}

// NewDnf creates a new DNF package manager instance
func NewDnf() types.Installer {
	return &dnf{
		basePackageManager: &basePackageManager{
			name:           "DNF",
			pmType:        types.TypeDnf,
			executableName: "dnf",
		},
	}
}

func (d *dnf) InstallPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := d.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package with -y to assume yes
	_, err = d.runCommand(ctx, "install", "-y", pkg)
	if err != nil {
		return fmt.Errorf("failed to install package: %w", err)
	}

	return nil
}

func (d *dnf) InstallMultiple(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// DNF can install multiple packages in one command
	args := append([]string{"install", "-y"}, packages...)
	_, err := d.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

func (d *dnf) UpdatePackageManager(ctx context.Context) error {
	// Update all packages
	_, err := d.runCommand(ctx, "upgrade", "-y")
	if err != nil {
		return fmt.Errorf("failed to upgrade packages: %w", err)
	}

	return nil
}

// checkIfInstalled overrides the base implementation with DNF-specific logic
func (d *dnf) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	output, err := d.runCommand(ctx, "list", "--installed", pkg)
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
