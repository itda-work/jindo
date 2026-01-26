#!/bin/bash
#
# jd - Claude Code configuration manager
# Installation script
#
# Usage:
#   curl -fsSL https://cdn.jsdelivr.net/gh/itda-skills/jindo@main/install.sh | bash
#
# Options:
#   VERSION        - Specific version to install (default: latest)
#   JD_INSTALL_DIR - Installation directory (default: ~/.local/bin)
#

set -euo pipefail

REPO="itda-skills/jindo"
BINARY="jd"
DEFAULT_INSTALL_DIR="$HOME/.local/bin"

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
            echo "macos"
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
    if [ -n "${VERSION:-}" ]; then
        version="${VERSION}"
        info "Using specified version: $version"
    else
        info "Fetching latest version..."
        version=$(get_latest_version)
        info "Latest version: $version"
    fi

    # Determine install directory
    local install_dir="${JD_INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
    local install_path="${install_dir}/${BINARY}"

    # Construct download URL
    local filename="${BINARY}-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        filename="${filename}.exe"
    fi
    download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"

    info "Downloading ${filename}..."

    # Create temporary directory
    tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # Download
    if ! curl -fsSL "$download_url" -o "$tmp_dir/$BINARY"; then
        error "Failed to download $download_url"
    fi

    # Make executable
    chmod +x "$tmp_dir/$BINARY"

    # Install
    info "Installing to ${install_path}..."

    # Create directory if it doesn't exist
    if [[ ! -d "$install_dir" ]]; then
        if mkdir -p "$install_dir" 2>/dev/null; then
            :
        else
            warn "Requesting sudo permission to create ${install_dir}"
            sudo mkdir -p "$install_dir"
        fi
    fi

    if [[ -w "$install_dir" ]]; then
        mv "$tmp_dir/$BINARY" "$install_path"
    else
        warn "Requesting sudo permission to install to ${install_dir}"
        sudo mv "$tmp_dir/$BINARY" "$install_path"
    fi

    # Verify
    if [[ -x "$install_path" ]]; then
        success "Successfully installed jd ${version}"
        echo ""
        "$install_path" version
        echo ""

        # Check if install_dir is in PATH
        if [[ ":$PATH:" != *":$install_dir:"* ]]; then
            warn "Note: ${install_dir} is not in your PATH"
            echo "  Add to your shell profile:"
            echo "    export PATH=\"\$PATH:${install_dir}\""
        fi
    else
        error "Installation failed"
    fi
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
