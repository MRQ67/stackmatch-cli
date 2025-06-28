# StackMatch CLI

**Clone environments, not just code.**

---

`stackmatch` is a command-line interface (CLI) tool designed to scan, export, and import development environment configurations. It helps teams mitigate "it works on my machine" issues by providing a snapshot of the tools, languages, and settings on a given system.

This tool is the first step toward building a robust system for managing and replicating development environments with ease.

## Features

- **Environment Scanning:** Detects system information, programming languages, development tools, package managers, and code editors.
- **JSON Export/Import:** Outputs the environment snapshot in a clean, structured JSON format, and can import it to set up a new environment.
- **Supabase Integration:** Authenticate and share your environment configurations with others using a Supabase backend.
- **Cross-Platform:** Works on Windows, macOS, and Linux.
- **Configuration Discovery:** Finds common configuration files (e.g., `.gitconfig`, `.npmrc`).
- **Extensible:** Built with a modular scanner that can be easily extended to detect more tools and configurations.

## Installation

To install the StackMatch CLI, you need to have Go installed on your system. You can then use `go install` to build and install the binary:

```sh
go install github.com/MRQ67/stackmatch-cli@latest
```

This will place the `stackmatch` executable in your Go binary path (`$GOPATH/bin`).

## Usage

Here are the primary commands available in the StackMatch CLI:

### Authentication

- `stackmatch register`: Create a new account.
- `stackmatch login`: Authenticate with your StackMatch account.
- `stackmatch logout`: Log out of your account.
- `stackmatch whoami`: Display the currently logged-in user.

### Environment Management

- `stackmatch export [filename]`: Scan the local environment and export it to a JSON file.
- `stackmatch import [filename]`: Import an environment from a local file.
- `stackmatch import --from-supabase --id <env_id>`: Import an environment from Supabase.
- `stackmatch push`: Push a local environment configuration to Supabase.
- `stackmatch pull`: Pull an environment configuration from Supabase.
- `stackmatch clone <username>/<env-name>`: Clone another user's public environment from Supabase.

### Other Commands

- `stackmatch version`: Display the current version of the StackMatch CLI.

## Development

Contributions are welcome! To get started with development:

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/MRQ67/stackmatch-cli.git
    cd stackmatch-cli
    ```

2.  **Build the binary:**
    ```sh
    go build -o stackmatch main.go
    ```

3.  **Run the tests:**
    ```sh
    go test ./...
    ```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.