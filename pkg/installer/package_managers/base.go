package package_managers

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
	"github.com/MRQ67/stackmatch-cli/pkg/version"
)

// basePackageManager provides common functionality for all package managers
type basePackageManager struct {
	name           string
	pmType        types.PackageManagerType
	executableName string
	// versionCommand is the command to get version information for a package
	versionCommand string
	// versionRegex is a regex pattern to extract version from command output
	versionRegex string
	// installWithFlags indicates if the package manager supports version flags (e.g., apt install pkg=1.0)
	installWithFlags bool
	// installPackageFunc is a function to install a package
	installPackageFunc func(ctx context.Context, pkg string) error
	// installMultipleFunc is a function to install multiple packages
	installMultipleFunc func(ctx context.Context, packages []string) error
	// uninstallPackageFunc is a function to uninstall a package
	uninstallPackageFunc func(ctx context.Context, pkg string) error
}

// UninstallPackage uninstalls a package using the package manager's uninstall command
func (b *basePackageManager) UninstallPackage(ctx context.Context, pkg string) error {
	if b.uninstallPackageFunc != nil {
		return b.uninstallPackageFunc(ctx, pkg)
	}
	
	// Default implementation tries to remove the package using the package manager's remove command
	_, err := b.runCommand(ctx, "remove", pkg)
	if err != nil {
		return fmt.Errorf("failed to uninstall package %s: %w", pkg, err)
	}
	return nil
}

// Installer is an interface that all package managers must implement
type Installer interface {
	types.Installer
	// InstallPackage installs a single package
	InstallPackage(ctx context.Context, pkg string) error
	// InstallMultiple installs multiple packages
	InstallMultiple(ctx context.Context, packages []string) error
	// InstallVersion installs a specific version of a package
	InstallVersion(ctx context.Context, pkg string, version types.VersionConstraint) error
	// InstallMultipleVersions installs multiple packages with specific versions
	InstallMultipleVersions(ctx context.Context, packages map[string]types.VersionConstraint) error
	// GetInstalledVersion gets the installed version of a package
	GetInstalledVersion(ctx context.Context, pkg string) (*types.PackageVersionInfo, error)
	// CheckVersion checks if the installed package satisfies the version constraint
	CheckVersion(ctx context.Context, pkg string, constraint types.VersionConstraint) (*types.PackageVersionInfo, error)
	// UninstallPackage uninstalls a package
	UninstallPackage(ctx context.Context, pkg string) error
}

func (b *basePackageManager) Name() string {
	return b.name
}

// Type returns the package manager type
func (b *basePackageManager) Type() types.PackageManagerType {
	return b.pmType
}

func (b *basePackageManager) IsAvailable() bool {
	_, err := exec.LookPath(b.executableName)
	return err == nil
}

// runCommand is a helper method to run shell commands
func (b *basePackageManager) runCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, b.executableName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\nOutput: %s", err, string(output))
	}
	return string(output), nil
}

// GetInstalledVersion gets the installed version of a package
func (b *basePackageManager) GetInstalledVersion(ctx context.Context, pkg string) (*types.PackageVersionInfo, error) {
	if b.versionCommand == "" {
		return &types.PackageVersionInfo{
			Name: pkg,
		}, nil
	}

	output, err := b.runCommand(ctx, b.versionCommand, pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	// If no regex is provided, return the raw output
	if b.versionRegex == "" {
		return &types.PackageVersionInfo{
			Name:    pkg,
			Version: strings.TrimSpace(output),
		}, nil
	}

	// Extract version using regex
	re := regexp.MustCompile(b.versionRegex)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return &types.PackageVersionInfo{
			Name:    pkg,
			Version: matches[1],
		}, nil
	}

	return &types.PackageVersionInfo{
		Name: pkg,
	}, nil
}

// CheckVersion checks if the installed package satisfies the version constraint
func (b *basePackageManager) CheckVersion(ctx context.Context, pkg string, constraint types.VersionConstraint) (*types.PackageVersionInfo, error) {
	info, err := b.GetInstalledVersion(ctx, pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed version: %w", err)
	}

	// If we couldn't determine the version, return what we have
	if info.Version == "" {
		info.Satisfies = false
		info.Constraint = constraint.Version
		return info, nil
	}

	// Parse the installed version
	installedVer, err := version.Parse(info.Version)
	if err != nil {
		info.Satisfies = false
		info.Constraint = constraint.Version
		return info, nil
	}

	// Check if it satisfies the constraint
	satisfies, err := installedVer.Satisfies(constraint.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid version constraint: %w", err)
	}

	info.Satisfies = satisfies
	info.Constraint = constraint.Version
	return info, nil
}

// InstallVersion installs a specific version of a package
func (b *basePackageManager) InstallVersion(ctx context.Context, pkg string, version types.VersionConstraint) error {
	// Use the provided install function if available
	if b.installPackageFunc != nil {
		return b.installPackageFunc(ctx, pkg)
	}

	// By default, try to append the version to the package name
	// This works for many package managers like apt, yum, etc.
	versionedPkg := fmt.Sprintf("%s=%s", pkg, version.Version)
	return b.installPackageFunc(ctx, versionedPkg)
}

// InstallMultipleVersions installs multiple packages with specific versions
func (b *basePackageManager) InstallMultipleVersions(ctx context.Context, packages map[string]types.VersionConstraint) error {
	// Use the provided install multiple function if available
	if b.installMultipleFunc != nil {
		var pkgs []string
		for pkg, ver := range packages {
			if ver.Version != "" {
				pkgs = append(pkgs, fmt.Sprintf("%s=%s", pkg, ver.Version))
			} else {
				pkgs = append(pkgs, pkg)
			}
		}
		return b.installMultipleFunc(ctx, pkgs)
	}

	// Fall back to installing one by one
	for pkg, ver := range packages {
		if err := b.InstallVersion(ctx, pkg, ver); err != nil {
			return fmt.Errorf("failed to install %s: %w", pkg, err)
		}
	}
	return nil
}

