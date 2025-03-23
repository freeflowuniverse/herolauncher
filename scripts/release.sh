#!/usr/bin/env bash

# Exit on error
set -e

# Function to get the latest release from GitHub
get_latest_release() {
    # Use a simpler approach with curl and grep
    local version
    version=$(curl -s "https://api.github.com/repos/freeflowuniverse/herolauncher/releases/latest" | 
              grep -o '"tag_name": *"[^"]*"' | 
              sed 's/.*"v\?\([^"]*\)".*/\1/')
    
    # If version is empty, check for error message
    if [ -z "$version" ]; then
        local error_message
        error_message=$(curl -s "https://api.github.com/repos/freeflowuniverse/herolauncher/releases/latest" | 
                        grep -o '"message": *"[^"]*"' | 
                        sed 's/.*"\([^"]*\)".*/\1/')
        
        if [[ "$error_message" == "Not Found" ]]; then
            echo "No releases found. You will create the first release." >&2
            return 0
        else
            echo "GitHub API Error: $error_message" >&2
            return 1
        fi
    fi
    
    echo "$version"
}

# Get repository root directory
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"

# Change to repository root
cd "$repo_root" || { echo "Error: Could not change to repository root directory" >&2; exit 1; }

# Show current version information
echo "Current project information:"
echo "----------------------------"

# Show Go version from go.mod
if [ -f "go.mod" ]; then
    go_version=$(grep "^go " go.mod | awk '{print $2}')
    echo "Go version in go.mod: $go_version"
fi

# Show current version from local git tags
echo "Fetching latest tag from local git repository..."
if git tag -l | grep -q "^v"; then
    latest_local_tag=$(git tag -l "v*" --sort=-v:refname | head -n 1)
    echo "Latest local git tag: $latest_local_tag"
else
    echo "No version tags found in local git repository."
fi

# Show current version from GitHub releases
echo "Fetching latest release information from GitHub..."
latest_release=$(get_latest_release 2>/dev/null)
if [ -z "$latest_release" ]; then
    echo "No previous GitHub releases found. This will be the first release."
else
    # Clean the version string
    latest_release=$(echo "$latest_release" | tr -d '\n\r')
    echo "Current latest GitHub release: $latest_release"
fi

# Show current git branch
current_branch=$(git rev-parse --abbrev-ref HEAD)
echo "Current git branch: $current_branch"

echo "----------------------------"

# Determine current version to suggest as default
current_version=""
if [ -n "$latest_local_tag" ]; then
    # Remove 'v' prefix if present
    current_version=${latest_local_tag#v}
elif [ -n "$latest_release" ]; then
    # Remove 'v' prefix if present
    current_version=${latest_release#v}
else
    current_version="0.1.0"  # Default if no version found
fi

# Suggest an incremented patch version
IFS='.' read -r major minor patch <<< "$current_version"
suggested_version="$major.$minor.$((patch + 1))"

# Ask for new version with the suggested version as default
read -p "Enter new version (current: $current_version, suggested: $suggested_version, press Enter to use suggested): " new_version

# Use suggested version if input is empty
if [ -z "$new_version" ]; then
    new_version="$suggested_version"
    echo "Using suggested version: $new_version"
fi

# Validate version format
if ! [[ $new_version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format X.Y.Z (e.g., 1.0.4)" >&2
    exit 1
fi

echo "Creating release v$new_version..."
echo "This will trigger the GitHub Actions release workflow which will:"
echo "- Build binaries for all platforms (Linux, macOS, Windows)"
echo "- Create a GitHub release with the tag name"
echo "- Upload the binaries as assets to the release"
echo "- Generate release notes based on the commits since the last release"

# Confirm with user
read -p "Do you want to continue? (y/n): " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
    echo "Release cancelled."
    exit 0
fi

# Prepare git commands
echo "Preparing git commands..."
cmd="
git remote set-url origin git@github.com:freeflowuniverse/herolauncher.git
git pull origin main
git tag -a \"v$new_version\" -m \"Release version $new_version\"
git push origin \"v$new_version\"
"

echo "Will execute:"
echo "$cmd"

# Confirm with user again
read -p "Execute these commands? (y/n): " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
    echo "Release cancelled."
    exit 0
fi

# Execute git commands
eval "$cmd"

echo "Release v$new_version created and pushed!"
echo "The GitHub Actions release workflow has been triggered."
echo "You can check the progress at: https://github.com/freeflowuniverse/herolauncher/actions/workflows/release.yml"
echo "Once completed, the release will be available at: https://github.com/freeflowuniverse/herolauncher/releases/tag/v$new_version"
