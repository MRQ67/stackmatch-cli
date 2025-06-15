package package_managers

import (
	"context"
	"fmt"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

type pacman struct {
	*basePackageManager
}

// NewPacman creates a new Pacman package manager instance
func NewPacman() types.Installer {
	return &pacman{
		basePackageManager: &basePackageManager{
			name:           "Pacman",
			pmType:        types.TypePacman,
			executableName: "pacman",
		},
	}
}

func (p *pacman) InstallPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := p.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package with --noconfirm to avoid prompts
	_, err = p.runCommand(ctx, "-S", "--noconfirm", pkg)
	if err != nil {
		return fmt.Errorf("failed to install package: %w", err)
	}

	return nil
}

func (p *pacman) InstallMultiple(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// Pacman can install multiple packages in one command
	args := append([]string{"-S", "--noconfirm"}, packages...)
	_, err := p.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

func (p *pacman) UpdatePackageManager(ctx context.Context) error {
	// Update package lists and upgrade all packages
	_, err := p.runCommand(ctx, "-Syu", "--noconfirm")
	if err != nil {
		return fmt.Errorf("failed to update packages: %w", err)
	}

	return nil
}

// checkIfInstalled overrides the base implementation with Pacman-specific logic
func (p *pacman) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	// pacman -Qs returns 0 if package is installed
	output, err := p.runCommand(ctx, "-Qs", "^"+pkg+"$")
	if err != nil {
		return false, nil
	}

	// If we get output, the package is installed
	return strings.TrimSpace(output) != "", nil
}
