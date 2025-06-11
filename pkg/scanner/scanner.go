package scanner

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

// DetectSystemInfo gathers basic OS and architecture details.
func DetectSystemInfo(sysInfo *types.SystemInfo) {
	sysInfo.OS = runtime.GOOS
	sysInfo.Arch = runtime.GOARCH

	// Shell detection
	if runtime.GOOS == "windows" {
		// On Windows, COMSPEC is a reliable env var for the command prompt.
		// PowerShell is also common, so we check for it in the PATH.
		if comspec, ok := os.LookupEnv("COMSPEC"); ok {
			sysInfo.Shell = comspec
		} else if _, err := exec.LookPath("powershell.exe"); err == nil {
			sysInfo.Shell = "powershell.exe"
		} else if _, err := exec.LookPath("pwsh.exe"); err == nil {
			sysInfo.Shell = "pwsh.exe" // PowerShell Core
		} else {
			sysInfo.Shell = "cmd.exe" // A fallback default
		}
	} else {
		// On Unix-like systems, SHELL is the standard.
		if shell, ok := os.LookupEnv("SHELL"); ok {
			sysInfo.Shell = shell
		} else {
			sysInfo.Shell = "/bin/sh" // A common default
		}
	}

	hostname, err := os.Hostname()
	if err == nil {
		sysInfo.Hostname = hostname
	}
}

// DetectConfigFiles checks for the existence of common configuration files in the user's home directory.
func DetectConfigFiles(envData *types.EnvironmentData) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Could not determine user home directory: %v", err)
		return
	}

	filesToScan := []string{
		".gitconfig",
		".npmrc",
		".zshrc",
		".bashrc",
		".bash_profile",
		".profile",
		// Add other common config files here
	}

	for _, file := range filesToScan {
		filePath := filepath.Join(homeDir, file)
		if _, err := os.Stat(filePath); err == nil {
			log.Printf("Found config file: %s", filePath)
			envData.ConfigFiles = append(envData.ConfigFiles, filePath)
		}
	}
}

// Executable represents a generic command-line tool to be scanned.
type Executable struct {
	Name         string
	Command      string
	VersionArg   string
	VersionRegex *regexp.Regexp
}

// detectExecutables is a generic helper to find tools, package managers, etc.
func detectExecutables(categoryName string, executables []Executable, dataMap map[string]string) {
	for _, exe := range executables {
		if _, err := exec.LookPath(exe.Command); err != nil {
			continue // Command not found in PATH, skip
		}

		if version := getCommandVersion(exe.Command, exe.VersionArg, exe.VersionRegex); version != "" {
			log.Printf("Found %s version %s", exe.Name, version)
			dataMap[exe.Name] = version
		} else {
			// If version command fails but executable exists, record its presence.
			dataMap[exe.Name] = "Installed"
		}
	}
}

// getCommandVersion executes a command and parses its version.
func getCommandVersion(command, versionArg string, versionRegex *regexp.Regexp) string {
	// Most version commands are fast, but we set a timeout to avoid hangs.
	cmd := exec.Command(command, versionArg)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := out.String()
	if err != nil {
		// Some tools print version to stderr (e.g., python --version)
		// We'll use stderr as a fallback if stdout is empty.
		if stderr.Len() > 0 {
			output = stderr.String()
		} else {
			log.Printf("Warning: Command '%s %s' failed: %v", command, versionArg, err)
			return ""
		}
	}

	return parseVersion(output, versionRegex)
}

// parseVersion extracts the version string using a regex.
func parseVersion(output string, regex *regexp.Regexp) string {
	if regex == nil {
		return ""
	}
	matches := regex.FindStringSubmatch(output)
	if len(matches) > 1 {
		// The first submatch is the captured version group.
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// DetectPackageManagers finds common package managers based on the OS.
func DetectPackageManagers(envData *types.EnvironmentData) {
	var executables []Executable

	// Common, cross-platform package managers
	crossPlatformExecutables := []Executable{
		{Name: "pip", Command: "pip", VersionArg: "--version", VersionRegex: regexp.MustCompile(`pip ([\d\.]+)`)},
		{Name: "pip3", Command: "pip3", VersionArg: "--version", VersionRegex: regexp.MustCompile(`pip ([\d\.]+)`)},
	}
	executables = append(executables, crossPlatformExecutables...)

	// OS-specific package managers
	switch runtime.GOOS {
	case "darwin":
		executables = append(executables, Executable{Name: "Homebrew", Command: "brew", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Homebrew ([\d\.]+)`)})
	case "linux":
		// For Linux, we can check for a few common ones.
		executables = append(executables, Executable{Name: "apt-get", Command: "apt-get", VersionArg: "--version", VersionRegex: regexp.MustCompile(`apt ([\d\.]+)`)})
		executables = append(executables, Executable{Name: "yum", Command: "yum", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)})
	case "windows":
		executables = append(executables, Executable{Name: "Chocolatey", Command: "choco", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)})
	}

	detectExecutables("Package Managers", executables, envData.PackageManagers)
}

// DetectProgrammingLanguages finds common programming languages.
func DetectProgrammingLanguages(envData *types.EnvironmentData) {
	languages := []Executable{
		{Name: "Go", Command: "go", VersionArg: "version", VersionRegex: regexp.MustCompile(`go version go([\d\.]+)`)},
		{Name: "Node.js", Command: "node", VersionArg: "--version", VersionRegex: regexp.MustCompile(`v?([\d\.]+)`)},
		{Name: "Python", Command: "python", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Python ([\d\.]+)`)},
		{Name: "Python 3", Command: "python3", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Python ([\d\.]+)`)},
	}
	detectExecutables("programming languages", languages, envData.ConfiguredLanguages)
}

// DetectTools finds common development tools and their versions.
func DetectTools(envData *types.EnvironmentData) {
	tools := []Executable{
		{Name: "Git", Command: "git", VersionArg: "--version", VersionRegex: regexp.MustCompile(`git version ([\d\.]+)`)},
		{Name: "Docker", Command: "docker", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Docker version ([\d\.]+)`)},
		{Name: "npm", Command: "npm", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "yarn", Command: "yarn", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "pnpm", Command: "pnpm", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
	}
	detectExecutables("development tools", tools, envData.Tools)
}

// DetectEditors finds common code editors.
func DetectEditors(envData *types.EnvironmentData) {
	editors := []Executable{
		{Name: "VS Code", Command: "code", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
	}
	detectExecutables("code editors", editors, envData.CodeEditors)
}
