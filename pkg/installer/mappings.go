package installer

import (
	"fmt"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

// PackageMapping defines a mapping for a package across different package managers
type PackageMapping struct {
	Name        string
	Description string
	Packages   map[types.PackageManagerType]string
}

// packageMappings contains the mapping of common packages across different package managers
var packageMappings = []PackageMapping{
	// Programming Languages
	{
		Name:        "nodejs",
		Description: "Node.js JavaScript runtime",
		Packages: map[types.PackageManagerType]string{
			types.TypeApt:        "nodejs",
			types.TypeDnf:        "nodejs",
			types.TypeYum:        "nodejs",
			types.TypePacman:     "nodejs",
			types.TypeHomebrew:   "node",
			types.TypeChocolatey: "nodejs",
			types.TypeScoop:      "nodejs",
			types.TypeWinget:     "OpenJS.NodeJS",
		},
	},
	{
		Name:        "python3",
		Description: "Python 3 interpreter",
		Packages: map[types.PackageManagerType]string{
			types.TypeApt:        "python3",
			types.TypeDnf:        "python3",
			types.TypeYum:        "python3",
			types.TypePacman:     "python",
			types.TypeHomebrew:    "python",
			types.TypeChocolatey: "python",
			types.TypeScoop:      "python",
			types.TypeWinget:     "Python.Python.3",
		},
	},

	// Development Tools
	{
		Name:        "git",
		Description: "Distributed version control system",
		Packages: map[types.PackageManagerType]string{
			types.TypeApt:        "git",
			types.TypeDnf:        "git",
			types.TypeYum:        "git",
			types.TypePacman:     "git",
			types.TypeHomebrew:    "git",
			types.TypeChocolatey: "git",
			types.TypeScoop:      "git",
			types.TypeWinget:     "Git.Git",
		},
	},

	// Databases
	{
		Name:        "postgresql",
		Description: "PostgreSQL database server",
		Packages: map[types.PackageManagerType]string{
			types.TypeApt:        "postgresql",
			types.TypeDnf:        "postgresql-server",
			types.TypeYum:        "postgresql-server",
			types.TypePacman:     "postgresql",
			types.TypeHomebrew:    "postgresql@14",
			types.TypeChocolatey: "postgresql",
			types.TypeScoop:      "postgresql",
			types.TypeWinget:     "PostgreSQL.pgAdmin",
		},
	},

	// Containerization
	{
		Name:        "docker",
		Description: "Docker container platform",
		Packages: map[types.PackageManagerType]string{
			types.TypeApt:        "docker.io",
			types.TypeDnf:        "docker",
			types.TypeYum:        "docker",
			types.TypePacman:     "docker",
			types.TypeHomebrew:    "docker",
			types.TypeChocolatey: "docker-desktop",
			types.TypeScoop:      "docker",
			types.TypeWinget:     "Docker.DockerDesktop",
		},
	},
}

// packageNameCache caches package name lookups to avoid repeated searches
var packageNameCache = make(map[string]map[types.PackageManagerType]string)

// init initializes the package name cache
func init() {
	for _, mapping := range packageMappings {
		packageNameCache[strings.ToLower(mapping.Name)] = mapping.Packages
	}
}

// GetPackageName returns the package name for a given package and package manager
func GetPackageName(pkg string, pmType types.PackageManagerType) (string, error) {
	// Check if the package name exists in our mappings
	for _, mapping := range packageMappings {
		if strings.EqualFold(mapping.Name, pkg) {
			// Check if we have a mapping for this package manager
			if pkgName, ok := mapping.Packages[pmType]; ok {
				return pkgName, nil
			}
			// No mapping for this package manager
			return "", fmt.Errorf("no mapping found for package '%s' on package manager %s", pkg, pmType)
		}
	}
	// No mapping found, return the original package name
	return pkg, nil
}

// GetPackageManagerType returns the PackageManagerType for a given installer
func GetPackageManagerType(installerInst Installer) types.PackageManagerType {
	if installerInst == nil {
		return ""
	}
	return installerInst.Type()
}

// GetPackageManagerName returns the display name for a package manager type
func GetPackageManagerName(pmType types.PackageManagerType) string {
	switch pmType {
	case types.TypeApt:
		return "APT"
	case types.TypeDnf:
		return "DNF"
	case types.TypeYum:
		return "YUM"
	case types.TypePacman:
		return "Pacman"
	case types.TypeSnap:
		return "Snap"
	case types.TypeHomebrew:
		return "Homebrew"
	case types.TypeChocolatey:
		return "Chocolatey"
	case types.TypeScoop:
		return "Scoop"
	case types.TypeWinget:
		return "Winget"
	default:
		return string(pmType)
	}
}

// GetAllPackageMappings returns all package mappings
func GetAllPackageMappings() []PackageMapping {
	return packageMappings
}

// AddPackageMapping adds a new package mapping
func AddPackageMapping(mapping PackageMapping) error {
	// Validate the mapping
	if mapping.Name == "" {
		return fmt.Errorf("package name cannot be empty")
	}
	if len(mapping.Packages) == 0 {
		return fmt.Errorf("at least one package manager mapping is required")
	}

	// Add the new mapping
	packageMappings = append(packageMappings, mapping)
	return nil
}
