#!/bin/bash

# StackMatch installer script
# Usage: curl -sSL https://raw.githubusercontent.com/yourusername/stackmatch/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) 
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *) 
        echo -e "${RED}Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

# GitHub repository
REPO="MRQ67/stackmatch-cli"
BINARY_NAME="stackmatch"

if [ "$OS" = "windows" ]; then
    BINARY_NAME="stackmatch.exe"
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/stackmatch-$OS-$ARCH"
if [ "$OS" = "windows" ]; then
    DOWNLOAD_URL="$DOWNLOAD_URL.exe"
fi

echo -e "${GREEN}Installing StackMatch CLI...${NC}"
echo -e "OS: $OS"
echo -e "Architecture: $ARCH"
echo -e "Download URL: $DOWNLOAD_URL"
echo ""

# Create temp directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Download binary
echo -e "${YELLOW}Downloading StackMatch...${NC}"
if command -v curl >/dev/null 2>&1; then
    curl -L "$DOWNLOAD_URL" -o "$BINARY_NAME"
elif command -v wget >/dev/null 2>&1; then
    wget "$DOWNLOAD_URL" -O "$BINARY_NAME"
else
    echo -e "${RED}Error: curl or wget is required${NC}"
    exit 1
fi

# Make executable
chmod +x "$BINARY_NAME"

# Install to /usr/local/bin or user's bin directory
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

echo -e "${YELLOW}Installing to $INSTALL_DIR...${NC}"
mv "$BINARY_NAME" "$INSTALL_DIR/stackmatch"

# Clean up
cd - > /dev/null
rm -rf "$TMP_DIR"

# Verify installation
if command -v stackmatch >/dev/null 2>&1; then
    echo -e "${GREEN}✓ StackMatch installed successfully!${NC}"
    echo ""
    echo "Try it out:"
    echo "  stackmatch --help"
    echo "  stackmatch scan"
    echo ""
    echo "Join our community: https://github.com/$REPO"
else
    echo -e "${YELLOW}Installation complete, but 'stackmatch' not found in PATH.${NC}"
    echo "You may need to add $INSTALL_DIR to your PATH:"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
fi