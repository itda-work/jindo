#!/bin/bash
#
# jd - Claude Code configuration manager
# Installation script
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/itda-work/itda-jindo/main/install.sh | bash
#
# Options:
#   INSTALL_DIR - Installation directory (default: /usr/local/bin)
#   VERSION     - Specific version to install (default: latest)
#

set -e

REPO="itda-work/itda-jindo"
BINARY="jd"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}INFO:${NC} $1"
}

success() {
    echo -e "${GREEN}SUCCESS:${NC} $1"
}

warn() {
    echo -e "${YELLOW}WARNING:${NC} $1"
}

error() {
    echo -e "${RED}ERROR:${NC} $1" >&2
    exit 1
}

# Detect OS
detect_os() {
    local os
    os="$(uname -s)"
    case "$os" in
        Darwin)
            echo "darwin"
            ;;
        Linux)
            echo "linux"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            echo "windows"
            ;;
        *)
            error "Unsupported operating system: $os"
            ;;
    esac
}

# Detect architecture
detect_arch() {
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)
            echo "amd64"
            ;;
        arm64|aarch64)
            echo "arm64"
            ;;
        *)
            error "Unsupported architecture: $arch"
            ;;
    esac
}

# Get latest version from GitHub
get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        error "Failed to get latest version"
    fi
    echo "$version"
}

# Download and install
install() {
    local os arch version download_url tmp_dir

    os=$(detect_os)
    arch=$(detect_arch)

    info "Detected OS: $os"
    info "Detected Architecture: $arch"

    # Get version
    if [ -n "$VERSION" ]; then
        version="$VERSION"
        info "Using specified version: $version"
    else
        info "Fetching latest version..."
        version=$(get_latest_version)
        info "Latest version: $version"
    fi

    # Construct download URL
    # Expected format: jd_<version>_<os>_<arch>.tar.gz
    local filename="${BINARY}_${version#v}_${os}_${arch}.tar.gz"
    download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"

    info "Downloading from: $download_url"

    # Create temporary directory
    tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # Download
    if ! curl -fsSL "$download_url" -o "$tmp_dir/$filename"; then
        error "Failed to download $download_url"
    fi

    # Extract
    info "Extracting..."
    tar -xzf "$tmp_dir/$filename" -C "$tmp_dir"

    # Install
    info "Installing to $INSTALL_DIR..."

    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$tmp_dir/$BINARY" "$INSTALL_DIR/$BINARY"
        chmod +x "$INSTALL_DIR/$BINARY"
    else
        warn "Need sudo permission to install to $INSTALL_DIR"
        sudo mv "$tmp_dir/$BINARY" "$INSTALL_DIR/$BINARY"
        sudo chmod +x "$INSTALL_DIR/$BINARY"
    fi

    success "Installed $BINARY to $INSTALL_DIR/$BINARY"

    # Verify installation
    if command -v "$BINARY" &> /dev/null; then
        info "Version: $($BINARY --version)"
    else
        echo ""
        warn "$INSTALL_DIR is not in your PATH"
        echo ""
        echo "Add the following to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
        echo "Then restart your shell or run:"
        echo ""
        echo "  source ~/.bashrc  # or ~/.zshrc"
        echo ""
    fi

    echo ""
    success "Installation complete!"
    echo ""
    echo "Get started with:"
    echo "  $BINARY --help"
    echo ""
}

# Main
main() {
    echo ""
    echo "  ╭──────────────────────────────────────╮"
    echo "  │                                      │"
    echo "  │   jd - Claude Code Config Manager    │"
    echo "  │                                      │"
    echo "  ╰──────────────────────────────────────╯"
    echo ""

    install
}

main
