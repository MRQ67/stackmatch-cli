# StackMatch CLI

**Clone environments, not just code.**

---

`stackmatch` is a command-line interface (CLI) tool designed to scan, export, and import development environment configurations. It helps teams mitigate "it works on my machine" issues by providing a snapshot of the tools, languages, and settings on a given system.

This tool is the first step toward building a robust system for managing and replicating development environments with ease.

## Features

- **Environment Scanning:** Detects system information, programming languages, development tools, package managers, and code editors.
- **JSON Export:** Outputs the environment snapshot in a clean, structured JSON format.
- **Cross-Platform:** Works on Windows, macOS, and Linux.
- **Configuration Discovery:** Finds common configuration files (e.g., `.gitconfig`, `.npmrc`).
- **Extensible:** Built with a modular scanner that can be easily extended to detect more tools and configurations.

## Installation

### Windows (Recommended)

For the best experience on Windows, we recommend using the official installer, which provides a familiar setup wizard and automatically adds the CLI to your system's PATH.

1.  Go to the [**GitHub Releases**](https://github.com/MRQ67/stackmatch-cli/releases) page.
2.  Download the `stackmatch-cli-setup.exe` file from the latest release.
3.  Run the installer and follow the on-screen instructions.

The installer will automatically add `stackmatch-cli` to your system's PATH, so you can run it from any command prompt (like PowerShell or CMD) after installation.

### macOS & Linux (or with Go)

If you are on macOS, Linux, or prefer to use Go directly, you can install the CLI using `go install`:

```sh
go install github.com/MRQ67/stackmatch-cli@latest
```

Ensure that your Go binary path (typically `$GOPATH/bin` or `$HOME/go/bin`) is included in your system's `PATH` environment variable.

## Usage

Here are the primary commands available in the StackMatch CLI:

### `scan`

Scans the local system and prints the environment data as JSON to standard output.

```sh
stackmatch-cli scan
```

### `export`

Scans the environment and saves the JSON output to a specified file.

```sh
stackmatch-cli export my-environment.json
```

### `import`

Reads an environment JSON file and displays a summary. (Note: The MVP version does not perform any installations or system modifications).

```sh
stackmatch-cli import my-environment.json
```

### `version`

Displays the current version of the StackMatch CLI.

```sh
stackmatch-cli version
```

## Development

Contributions are welcome! To get started with development:

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/MRQ67/stackmatch-cli.git
    cd stackmatch-cli
    ```

2.  **Build the binary:**
    ```sh
    go build -o stackmatch-cli main.go
    ```

3.  **Run the tests:**
    ```sh
    go test ./...
    ```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

