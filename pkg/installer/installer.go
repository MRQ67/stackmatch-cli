package installer

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/installer/package_managers"
	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

// Installer is an alias for types.Installer for backward compatibility
type Installer = types.Installer

// DetectPackageManager detects the best available package manager for the current system
func DetectPackageManager() (Installer, error) {
	// Check package managers in order of preference based on OS
	var managers []Installer

	switch runtime.GOOS {
	case "windows":
		managers = []Installer{
			package_managers.NewChocolatey(),
			package_managers.NewScoop(),
			package_managers.NewWinget(),
		}
	case "darwin":
		managers = []Installer{
			package_managers.NewHomebrew(),
		}
	default: // Linux and others
		managers = []Installer{
			package_managers.NewApt(),
			package_managers.NewDnf(),
			package_managers.NewYum(),
			package_managers.NewPacman(),
			package_managers.NewSnap(),
		}
	}

	// Return the first available package manager
	for _, mgr := range managers {
		if mgr.IsAvailable() {
			return mgr, nil
		}
	}

	return nil, fmt.Errorf("no supported package manager found")
}

// installWithMapping installs a package using the appropriate package name for the installer
func installWithMapping(ctx context.Context, installerInst Installer, pkg string) error {
	// Get the package name for this specific package manager
	pmType := installerInst.Type()
	mappedPkg, err := GetPackageName(pkg, pmType)
	if err != nil {
		return fmt.Errorf("package mapping error: %w", err)
	}

	// If we get an empty package name, it means no mapping was found
	// and we should use the original package name
	if mappedPkg == "" {
		mappedPkg = pkg
	}

	// Install the package
	err = installerInst.InstallPackage(ctx, mappedPkg)
	if err != nil {
		// If we get a PackageNotFoundError, try with the original package name
		if _, ok := err.(*types.PackageNotFoundError); ok && mappedPkg != pkg {
			err = installerInst.InstallPackage(ctx, pkg)
		}
		return err
	}
	return nil
}

// runCommand is a helper function to run shell commands
func runCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\nOutput: %s", err, string(output))
	}
	return string(output), nil
}

// InstallPackage installs a package using the best available package manager
func InstallPackage(ctx context.Context, pkg string) error {
	// Get the best available package manager
	pm, err := DetectPackageManager()
	if err != nil {
		return fmt.Errorf("no package manager available: %w", err)
	}

	return installWithMapping(ctx, pm, pkg)
}

// InstallPackages installs multiple packages using the best available package manager
func InstallPackages(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// Get the best available package manager
	pm, err := DetectPackageManager()
	if err != nil {
		return fmt.Errorf("no package manager available: %w", err)
	}

	// Try to install all packages at once first
	err = pm.InstallMultiple(ctx, packages)
	if err == nil {
		return nil
	}

	// If batch install fails, try installing one by one
	var failed []string
	for _, pkg := range packages {
		err := installWithMapping(ctx, pm, pkg)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", pkg, err))
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("failed to install packages: %s", strings.Join(failed, "; "))
	}

	return nil
}
