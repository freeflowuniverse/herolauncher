#!/bin/bash

set -e

# Function to print colored output
print_color() {
    local color=$1
    local text=$2
    case $color in
        "red") echo -e "\033[0;31m$text\033[0m" ;;
        "green") echo -e "\033[0;32m$text\033[0m" ;;
        "yellow") echo -e "\033[0;33m$text\033[0m" ;;
        "blue") echo -e "\033[0;34m$text\033[0m" ;;
        *) echo "$text" ;;
    esac
}

# Function to get the latest release version from GitHub
get_latest_release() {
    # Use a simpler approach with curl and grep
    local version
    version=$(curl -s "https://api.github.com/repos/freeflowuniverse/herolauncher/releases/latest" | 
              grep -o '"tag_name": *"[^"]*"' | 
              sed 's/.*"v\?\([^"]*\)".*/\1/')
    
    # If version is empty, use a default version
    if [ -z "$version" ]; then
        print_color "yellow" "Could not determine version from API response. Using default version."
        echo "0.1.0"
        return 0
    fi
    
    echo "$version"
}

# Determine OS and architecture
os_name="$(uname -s)"
arch_name="$(uname -m)"

# Map architecture names
case "$arch_name" in
    "x86_64") arch="amd64" ;;
    "aarch64" | "arm64") arch="arm64" ;;
    *) 
        print_color "red" "Unsupported architecture: $arch_name"
        exit 1
        ;;
esac

# Map OS names
case "$os_name" in
    "Linux") os="linux" ;;
    "Darwin") os="darwin" ;;
    *) 
        print_color "red" "Unsupported operating system: $os_name"
        exit 1
        ;;
esac

# Get the latest version or use the specified one
if [ -z "$1" ]; then
    print_color "blue" "Fetching latest release version..."
    # Capture the output of get_latest_release in a variable without any debug output
    version=$(get_latest_release 2>/dev/null)
    if [ -z "$version" ]; then
        print_color "red" "Error: Could not determine latest version"
        exit 1
    fi
    print_color "green" "Latest version: $version"
else
    version="$1"
    print_color "blue" "Using specified version: $version"
fi

# Determine file extension based on OS
if [[ "$os" == "linux" || "$os" == "darwin" ]]; then
    ext="tar.gz"
    archive_name="herolauncher-${os}-${arch}.${ext}"
elif [[ "$os" == "windows" ]]; then
    ext="zip"
    archive_name="herolauncher-${os}-${arch}.${ext}"
else
    print_color "red" "Unsupported platform: $os $arch"
    exit 1
fi

# Construct the download URL - ensure version is clean
version_clean=$(echo "$version" | tr -d '\n\r')
base_url="https://github.com/freeflowuniverse/herolauncher/releases/download/v${version_clean}"
url="${base_url}/${archive_name}"

print_color "blue" "Download URL: $url"

# Check for existing herolauncher installations
print_color "blue" "Checking for existing installations..."
existing_hero=$(which herolauncher 2>/dev/null || true)
if [ ! -z "$existing_hero" ]; then
    print_color "yellow" "Found existing herolauncher installation at: $existing_hero"
    if [ -w "$(dirname "$existing_hero")" ]; then
        print_color "blue" "Removing existing herolauncher installation..."
        rm "$existing_hero" || { 
            print_color "red" "Error: Failed to remove existing herolauncher binary at $existing_hero"
            exit 1
        }
    else
        print_color "red" "Error: Cannot remove existing herolauncher installation at $existing_hero (permission denied)"
        print_color "yellow" "Please remove it manually with sudo and run this script again"
        exit 1
    fi
fi

# Check for macOS-specific requirements
if [[ "$os" == "darwin" ]]; then
    print_color "blue" "Detected macOS system, checking requirements..."
    
    # Check if /usr/local/bin/herolauncher exists and remove it
    if [ -f /usr/local/bin/herolauncher ]; then
        print_color "yellow" "Removing existing herolauncher binary from /usr/local/bin..."
        rm /usr/local/bin/herolauncher 2>/dev/null || {
            print_color "red" "Error: Failed to remove existing herolauncher binary from /usr/local/bin"
            print_color "yellow" "You may need to use sudo to remove it"
        }
    fi
fi

# Ensure URL is valid
if [ -z "$url" ]; then
    print_color "red" "Error: Could not determine download URL"
    exit 1
fi

# Set up trap to clean up temporary files on exit
trap 'rm -rf "${temp_dir}" "${temp_file}" 2>/dev/null || true' EXIT

# Create installation directories
install_dir="${HOME}/herolauncher"
bin_dir="${install_dir}/bin"
mkdir -p "${bin_dir}"

# Download the archive
print_color "blue" "Downloading HeroLauncher from: $url"
temp_dir=$(mktemp -d)
archive_path="${temp_dir}/${archive_name}"
curl -L -o "${archive_path}" "${url}" || {
    print_color "red" "Error: Failed to download HeroLauncher"
    rm -rf "${temp_dir}"
    exit 1
}

# Check if file size is reasonable (at least 1 MB)
file_size=$(du -m "${archive_path}" | cut -f1)
if [ "$file_size" -lt 1 ]; then
    print_color "red" "Error: Downloaded file is too small (${file_size}MB). Download may have failed."
    rm -rf "${temp_dir}"
    exit 1
fi

# Extract the archive based on file type
print_color "blue" "Extracting archive..."
cd "${temp_dir}"
if [[ "$ext" == "tar.gz" ]]; then
    tar -xzf "${archive_path}" || {
        print_color "red" "Error: Failed to extract tar.gz archive"
        rm -rf "${temp_dir}"
        exit 1
    }
elif [[ "$ext" == "zip" ]]; then
    unzip "${archive_path}" || {
        print_color "red" "Error: Failed to extract zip archive"
        rm -rf "${temp_dir}"
        exit 1
    }
fi

# Copy binaries to installation directory and create herolauncher symlink
print_color "blue" "Installing binaries to ${bin_dir}..."
find bin -type f -executable | while read -r binary; do
    binary_name=$(basename "${binary}")
    cp "${binary}" "${bin_dir}/" || {
        print_color "red" "Error: Failed to copy ${binary_name} to ${bin_dir}"
        rm -rf "${temp_dir}"
        exit 1
    }
    chmod +x "${bin_dir}/${binary_name}"
    print_color "green" "Installed: ${bin_dir}/${binary_name}"
    
    # Create herolauncher symlink for the first binary (pmclient)
    if [[ "$binary_name" == *"pmclient"* ]]; then
        ln -sf "${bin_dir}/${binary_name}" "${bin_dir}/herolauncher" || {
            print_color "red" "Error: Failed to create herolauncher symlink"
            rm -rf "${temp_dir}"
            exit 1
        }
        print_color "green" "Created herolauncher symlink to ${binary_name}"
    fi
done

# Clean up temporary directory
rm -rf "${temp_dir}"

# Update PATH in shell profile files
print_color "blue" "Updating PATH in shell profile files..."
for profile in "${HOME}/.bashrc" "${HOME}/.zshrc" "${HOME}/.bash_profile" "${HOME}/.zprofile"; do
    if [ -f "${profile}" ]; then
        # Remove any existing HeroLauncher PATH entries
        sed -i.bak '/herolauncher/d' "${profile}" 2>/dev/null || true
        # Add new PATH entry
        echo "export PATH=\"\${PATH}:${bin_dir}\"" >> "${profile}"
        print_color "green" "Updated ${profile}"
    fi
done

# Create symlinks in /usr/local/bin if possible (requires sudo)
if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
    print_color "blue" "Creating symlinks in /usr/local/bin..."
    # Create herolauncher symlink
    ln -sf "${bin_dir}/herolauncher" "/usr/local/bin/herolauncher" || true
    print_color "green" "Created symlink: /usr/local/bin/herolauncher"
    
    # Create symlinks for all binaries
    find "${bin_dir}" -type f -executable | while read -r binary; do
        binary_name=$(basename "${binary}")
        ln -sf "${binary}" "/usr/local/bin/${binary_name}" || true
        print_color "green" "Created symlink: /usr/local/bin/${binary_name}"
    done
else
    print_color "yellow" "Note: Could not create symlinks in /usr/local/bin (permission denied)"
    print_color "yellow" "To create symlinks manually, run:"
    print_color "yellow" "  sudo ln -sf ${bin_dir}/herolauncher /usr/local/bin/herolauncher"
fi

print_color "green" "HeroLauncher v${version} installed successfully!"
print_color "yellow" "IMPORTANT: You need to restart your terminal or source your shell profile to update your PATH."
print_color "yellow" "To update your PATH in the current terminal, run one of the following commands:"

# Determine the user's shell
current_shell=$(basename "$SHELL")
case "$current_shell" in
    "bash")
        print_color "blue" "  source ~/.bashrc"
        ;;
    "zsh")
        print_color "blue" "  source ~/.zshrc"
        ;;
    *)
        print_color "blue" "  source ~/.bashrc  # or the appropriate profile file for your shell"
        ;;
esac

print_color "blue" "To verify installation, run: herolauncher -version"
print_color "blue" "Or directly from the installation directory: ${bin_dir}/herolauncher -version"

# Check if herolauncher is in the PATH
if command -v herolauncher >/dev/null 2>&1; then
    print_color "green" "herolauncher command is available in your PATH."
else
    print_color "yellow" "herolauncher command is not yet available in your PATH."
    print_color "yellow" "You can run it directly using: ${bin_dir}/herolauncher"
    print_color "yellow" "Or create a symlink manually: sudo ln -sf ${bin_dir}/herolauncher /usr/local/bin/herolauncher"
fi
