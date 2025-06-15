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

	// InstallMultiple installs multiple packages in a single operation when possible
	InstallMultiple(ctx context.Context, packages []string) error

	// UpdatePackageManager updates the package manager itself
	UpdatePackageManager(ctx context.Context) error
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
