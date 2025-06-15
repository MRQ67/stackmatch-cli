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
		// Version control
		".gitconfig", ".git-credentials", ".gitignore_global",

		// Shell configurations
		".zshrc", ".bashrc", ".bash_profile", ".profile", ".zprofile", ".zshenv", ".bash_login",

		// Package manager configs
		".npmrc", ".yarnrc", ".pypirc", ".m2/settings.xml", ".gradle/gradle.properties",

		// Editor/IDE configs
		".vscode/settings.json", ".idea/", ".vimrc", ".vim/", ".emacs.d/",

		// Container and virtualization
		"Dockerfile", "docker-compose.yml", "docker-compose.yaml", ".dockerignore",

		// Environment files
		".env", ".env.local", ".env.development", ".env.production",

		// Language/framework specific
		"package.json", "yarn.lock", "package-lock.json", "pnpm-lock.yaml",
		"go.mod", "go.sum", "requirements.txt", "Pipfile", "poetry.lock",
		"Cargo.toml", "Gemfile", "Gemfile.lock", "composer.json", "composer.lock",
		"tsconfig.json", "webpack.config.js", "babel.config.js",
		".eslintrc", ".eslintrc.js", ".eslintrc.json", ".prettierrc", ".prettierrc.js",
		".babelrc", ".babelrc.js", ".babelrc.json", "jest.config.js", ".npmignore",
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
		// Python
		{Name: "pip", Command: "pip", VersionArg: "--version", VersionRegex: regexp.MustCompile(`pip ([\d\.]+)`)},
		{Name: "pip3", Command: "pip3", VersionArg: "--version", VersionRegex: regexp.MustCompile(`pip ([\d\.]+)`)},
		{Name: "pipx", Command: "pipx", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "poetry", Command: "poetry", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Poetry version ([\d\.]+)`)},

		// JavaScript/Node.js
		{Name: "npm", Command: "npm", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "yarn", Command: "yarn", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "pnpm", Command: "pnpm", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},

		// Container
		{Name: "Docker", Command: "docker", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Docker version ([\d\.]+)`)},
		{Name: "Podman", Command: "podman", VersionArg: "--version", VersionRegex: regexp.MustCompile(`podman version ([\d\.]+)`)},
	}
	executables = append(executables, crossPlatformExecutables...)

	// OS-specific package managers
	switch runtime.GOOS {
	case "darwin":
		executables = append(executables,
			Executable{Name: "Homebrew", Command: "brew", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Homebrew ([\d\.]+)`)},
			Executable{Name: "MacPorts", Command: "port", VersionArg: "version", VersionRegex: regexp.MustCompile(`version ([\d\.]+)`)},
		)
	case "linux":
		// For Linux, we can check for common package managers
		executables = append(executables,
			Executable{Name: "apt", Command: "apt", VersionArg: "--version", VersionRegex: regexp.MustCompile(`apt ([\d\.]+)`)},
			Executable{Name: "apt-get", Command: "apt-get", VersionArg: "--version", VersionRegex: regexp.MustCompile(`apt-get ([\d\.]+)`)},
			Executable{Name: "yum", Command: "yum", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
			Executable{Name: "dnf", Command: "dnf", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
			Executable{Name: "pacman", Command: "pacman", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Pacman v([\d\.]+)`)},
			Executable{Name: "zypper", Command: "zypper", VersionArg: "--version", VersionRegex: regexp.MustCompile(`zypper ([\d\.]+)`)},
			Executable{Name: "snap", Command: "snap", VersionArg: "--version", VersionRegex: regexp.MustCompile(`snap\\s+([\d\.]+)`)},
		)
	case "windows":
		executables = append(executables,
			Executable{Name: "Chocolatey", Command: "choco", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
			Executable{Name: "Scoop", Command: "scoop", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
			Executable{Name: "Winget", Command: "winget", VersionArg: "--version", VersionRegex: regexp.MustCompile(`v([\d\.]+)`)},
		)
	}

	detectExecutables("Package Managers", executables, envData.PackageManagers)
}

// DetectProgrammingLanguages finds common programming languages.
func DetectProgrammingLanguages(envData *types.EnvironmentData) {
	languages := []Executable{
		// Compiled Languages
		{Name: "Go", Command: "go", VersionArg: "version", VersionRegex: regexp.MustCompile(`go version go([\d\.]+)`)},
		{Name: "Rust", Command: "rustc", VersionArg: "--version", VersionRegex: regexp.MustCompile(`rustc ([\d\.]+)`)},
		{Name: "Java", Command: "java", VersionArg: "-version", VersionRegex: regexp.MustCompile(`version "([\d\._]+)"`)},
		{Name: "Kotlin", Command: "kotlin", VersionArg: "-version", VersionRegex: regexp.MustCompile(`Kotlin version ([\d\.]+)`)},
		{Name: "C#", Command: "dotnet", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "Scala", Command: "scala", VersionArg: "-version", VersionRegex: regexp.MustCompile(`version ([\d\.]+)`)},

		// Scripting Languages
		{Name: "Node.js", Command: "node", VersionArg: "--version", VersionRegex: regexp.MustCompile(`v?([\d\.]+)`)},
		{Name: "Python", Command: "python", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Python ([\d\.]+)`)},
		{Name: "Python 3", Command: "python3", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Python ([\d\.]+)`)},
		{Name: "Ruby", Command: "ruby", VersionArg: "--version", VersionRegex: regexp.MustCompile(`ruby ([\d\.p]+)`)},
		{Name: "PHP", Command: "php", VersionArg: "--version", VersionRegex: regexp.MustCompile(`PHP ([\d\.]+)`)},
		{Name: "Perl", Command: "perl", VersionArg: "--version", VersionRegex: regexp.MustCompile(`v([\d\.]+)`)},
		{Name: "Lua", Command: "lua", VersionArg: "-v", VersionRegex: regexp.MustCompile(`Lua ([\d\.]+)`)},

		// JVM Languages
		{Name: "Groovy", Command: "groovy", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Groovy Version: ([\d\.]+)`)},

		// Functional Languages
		{Name: "Haskell", Command: "ghc", VersionArg: "--version", VersionRegex: regexp.MustCompile(`version ([\d\.]+)`)},
		{Name: "Elixir", Command: "elixir", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Elixir ([\d\.]+)`)},
		{Name: "Clojure", Command: "clj", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Clojure CLI version ([\d\.]+)`)},

		// Web Technologies
		{Name: "TypeScript", Command: "tsc", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Version ([\d\.]+)`)},
		{Name: "Dart", Command: "dart", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Dart SDK version: ([\d\.]+)`)},

		// Shells
		{Name: "Bash", Command: "bash", VersionArg: "--version", VersionRegex: regexp.MustCompile(`version ([\d\.]+)`)},
		{Name: "Zsh", Command: "zsh", VersionArg: "--version", VersionRegex: regexp.MustCompile(`zsh ([\d\.]+)`)},
		{Name: "Fish", Command: "fish", VersionArg: "--version", VersionRegex: regexp.MustCompile(`fish, version ([\d\.]+)`)},

		// Database and Query Languages
		{Name: "SQLite", Command: "sqlite3", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "PostgreSQL", Command: "psql", VersionArg: "--version", VersionRegex: regexp.MustCompile(`psql \(PostgreSQL\) ([\d\.]+)`)},
		{Name: "MySQL", Command: "mysql", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Ver ([\d\.]+)`)},
	}
	detectExecutables("programming languages", languages, envData.ConfiguredLanguages)
}

// DetectTools finds common development tools and their versions.
func DetectTools(envData *types.EnvironmentData) {
	tools := []Executable{
		// Version Control
		{Name: "Git", Command: "git", VersionArg: "--version", VersionRegex: regexp.MustCompile(`git version ([\d\.]+)`)},
		{Name: "Mercurial", Command: "hg", VersionArg: "--version", VersionRegex: regexp.MustCompile(`version ([\d\.]+)`)},
		{Name: "Subversion", Command: "svn", VersionArg: "--version --quiet", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},

		// Containerization
		{Name: "Docker", Command: "docker", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Docker version ([\d\.]+)`)},
		{Name: "Docker Compose", Command: "docker-compose", VersionArg: "--version", VersionRegex: regexp.MustCompile(`docker-compose version ([\d\.]+)`)},
		{Name: "Kubernetes", Command: "kubectl", VersionArg: "version --client --short", VersionRegex: regexp.MustCompile(`Client Version: v([\d\.]+)`)},
		{Name: "Helm", Command: "helm", VersionArg: "version --short", VersionRegex: regexp.MustCompile(`v([\d\.]+)`)},

		// Build Tools
		{Name: "Make", Command: "make", VersionArg: "--version", VersionRegex: regexp.MustCompile(`GNU Make ([\d\.]+)`)},
		{Name: "CMake", Command: "cmake", VersionArg: "--version", VersionRegex: regexp.MustCompile(`cmake version ([\d\.]+)`)},
		{Name: "Gradle", Command: "gradle", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Gradle ([\d\.]+)`)},
		{Name: "Maven", Command: "mvn", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Apache Maven ([\d\.]+)`)},

		// Package Managers (not in package managers to avoid duplication)
		{Name: "npm", Command: "npm", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "yarn", Command: "yarn", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "pnpm", Command: "pnpm", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "pip", Command: "pip", VersionArg: "--version", VersionRegex: regexp.MustCompile(`pip ([\d\.]+)`)},
		{Name: "pip3", Command: "pip3", VersionArg: "--version", VersionRegex: regexp.MustCompile(`pip ([\d\.]+)`)},

		// Cloud CLIs
		{Name: "AWS CLI", Command: "aws", VersionArg: "--version", VersionRegex: regexp.MustCompile(`aws-cli/([\d\.]+)`)},
		{Name: "Azure CLI", Command: "az", VersionArg: "--version", VersionRegex: regexp.MustCompile(`azure-cli\s+([\d\.]+)`)},
		{Name: "Google Cloud SDK", Command: "gcloud", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Google Cloud SDK ([\d\.]+)`)},

		// Infrastructure as Code
		{Name: "Terraform", Command: "terraform", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Terraform v([\d\.]+)`)},
		{Name: "Ansible", Command: "ansible", VersionArg: "--version", VersionRegex: regexp.MustCompile(`ansible \[core ([\d\.]+)\](?:\n|\r\n)?`)},
		{Name: "Packer", Command: "packer", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},

		// Security
		{Name: "OpenSSL", Command: "openssl", VersionArg: "version", VersionRegex: regexp.MustCompile(`OpenSSL ([\d\.]+[a-z]*)`)},

		// Testing
		{Name: "Jest", Command: "jest", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "Pytest", Command: "pytest", VersionArg: "--version", VersionRegex: regexp.MustCompile(`pytest ([\d\.]+)`)},
	}
	detectExecutables("development tools", tools, envData.Tools)
}

// DetectEditors finds common code editors and IDEs.
func DetectEditors(envData *types.EnvironmentData) {
	editors := []Executable{
		// Lightweight Editors
		{Name: "VS Code", Command: "code", VersionArg: "--version", VersionRegex: regexp.MustCompile(`([\d\.]+)`)},
		{Name: "Sublime Text", Command: "subl", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Sublime Text Build ([\d\.]+)`)},
		{Name: "Atom", Command: "atom", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Atom\s+:\s+([\d\.]+)`)},
		{Name: "Vim", Command: "vim", VersionArg: "--version", VersionRegex: regexp.MustCompile(`VIM - Vi IMproved ([\d\.]+)`)},
		{Name: "Neovim", Command: "nvim", VersionArg: "--version", VersionRegex: regexp.MustCompile(`NVIM v([\d\.]+)`)},
		{Name: "Emacs", Command: "emacs", VersionArg: "--version", VersionRegex: regexp.MustCompile(`GNU Emacs ([\d\.]+)`)},
		{Name: "Nano", Command: "nano", VersionArg: "--version", VersionRegex: regexp.MustCompile(`nano version ([\d\.]+)`)},

		// Full IDEs
		{Name: "IntelliJ IDEA", Command: "idea", VersionArg: "--version", VersionRegex: regexp.MustCompile(`(?:IntelliJ IDEA|IntelliJ IDEA Community Edition) ([\d\.]+)`)},
		{Name: "PyCharm", Command: "pycharm", VersionArg: "--version", VersionRegex: regexp.MustCompile(`PyCharm ([\d\.]+)`)},
		{Name: "WebStorm", Command: "webstorm", VersionArg: "--version", VersionRegex: regexp.MustCompile(`WebStorm ([\d\.]+)`)},
		{Name: "GoLand", Command: "goland", VersionArg: "--version", VersionRegex: regexp.MustCompile(`GoLand ([\d\.]+)`)},
		{Name: "Android Studio", Command: "studio", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Android Studio ([\d\.]+)`)},
		{Name: "Xcode", Command: "xcodebuild", VersionArg: "-version", VersionRegex: regexp.MustCompile(`Xcode ([\d\.]+)`)},
		{Name: "Visual Studio", Command: "devenv", VersionArg: "/?", VersionRegex: regexp.MustCompile(`Microsoft Visual Studio ([\d\.]+)`)},

		// Database Tools
		{Name: "DBeaver", Command: "dbeaver", VersionArg: "--version", VersionRegex: regexp.MustCompile(`DBeaver ([\d\.]+)`)},
		{Name: "TablePlus", Command: "tableplus", VersionArg: "--version", VersionRegex: regexp.MustCompile(`TablePlus ([\d\.]+)`)},

		// Version Control GUIs
		{Name: "GitHub Desktop", Command: "github", VersionArg: "--version", VersionRegex: regexp.MustCompile(`GitHub Desktop ([\d\.]+)`)},
		{Name: "GitKraken", Command: "gitkraken", VersionArg: "--version", VersionRegex: regexp.MustCompile(`GitKraken ([\d\.]+)`)},
		{Name: "Sourcetree", Command: "sourcetree", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Sourcetree ([\d\.]+)`)},

		// AI Code Editors
		{Name: "Windsurf", Command: "windsurf", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Windsurf ([\d\.]+)`)},
		{Name: "Cursor", Command: "cursor", VersionArg: "--version", VersionRegex: regexp.MustCompile(`Cursor ([\d\.]+)`)},
	}
	detectExecutables("code editors", editors, envData.CodeEditors)
}
