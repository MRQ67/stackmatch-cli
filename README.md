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

To install the StackMatch CLI, you need to have Go installed on your system. You can then use `go install` to build and install the binary:

```sh
go install github.com/MRQ67/stackmatch-cli@latest
```

This will place the `stackmatch-cli` executable in your Go binary path (`$GOPATH/bin`).

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

