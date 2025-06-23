package types

import "context"

// PackageManagerType represents the type of package manager
type PackageManagerType string

// Package manager type constants
const (
	TypeApt        PackageManagerType = "apt"
	TypeDnf        PackageManagerType = "dnf"
	TypeYum        PackageManagerType = "yum"
	TypePacman     PackageManagerType = "pacman"
	TypeSnap       PackageManagerType = "snap"
	TypeHomebrew   PackageManagerType = "homebrew"
	TypeChocolatey PackageManagerType = "chocolatey"
	TypeScoop      PackageManagerType = "scoop"
	TypeWinget     PackageManagerType = "winget"
)

// VersionConstraint represents a version constraint for a package
type VersionConstraint struct {
	Version string // The version string (e.g., "1.2.3", ">=1.2.0 <2.0.0")
}

// PackageVersionInfo contains version information about an installed package
type PackageVersionInfo struct {
	Name         string // Package name
	Version      string // Installed version
	Latest       string // Latest available version (if available)
	Satisfies    bool   // Whether the installed version satisfies the constraint
	Constraint   string // The version constraint that was checked (if any)
}

// Installer defines the interface for package manager operations
type Installer interface {
	// Name returns the name of the package manager
	Name() string

	// Type returns the package manager type (e.g., "apt", "homebrew")
	Type() PackageManagerType

	// IsAvailable checks if the package manager is available on the system
	IsAvailable() bool

	// InstallPackage installs a single package
	InstallPackage(ctx context.Context, pkg string) error

	// InstallVersion installs a specific version of a package
	InstallVersion(ctx context.Context, pkg string, version VersionConstraint) error

	// InstallMultiple installs multiple packages in a single operation when possible
	InstallMultiple(ctx context.Context, packages []string) error

	// InstallMultipleVersions installs multiple packages with specific versions
	InstallMultipleVersions(ctx context.Context, packages map[string]VersionConstraint) error

	// GetInstalledVersion gets information about an installed package
	GetInstalledVersion(ctx context.Context, pkg string) (*PackageVersionInfo, error)

	// CheckVersion checks if the installed package satisfies the version constraint
	CheckVersion(ctx context.Context, pkg string, constraint VersionConstraint) (*PackageVersionInfo, error)

	// UpdatePackageManager updates the package manager itself
	UpdatePackageManager(ctx context.Context) error

	// UninstallPackage uninstalls a package
	UninstallPackage(ctx context.Context, pkg string) error
}

// PackageInfo contains information about a package that can be installed
type PackageInfo struct {
	Name    string
	Version string // Optional version constraint
}

// InstallOptions contains options for package installation
type InstallOptions struct {
	DryRun     bool
	AssumeYes  bool
	NoDeps     bool
	SkipUpdate bool
}

// DefaultInstallOptions returns default installation options
func DefaultInstallOptions() InstallOptions {
	return InstallOptions{
		DryRun:     false,
		AssumeYes:  false,
		NoDeps:     false,
		SkipUpdate: false,
	}
}
