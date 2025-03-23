#!/usr/bin/env bash

# Exit on error
set -e

# Function to get the latest release from GitHub
get_latest_release() {
    local url="https://api.github.com/repos/freeflowuniverse/herolauncher/releases/latest"
    local response
    response=$(curl -s "$url")
    if [ $? -ne 0 ]; then
        echo "Error: Failed to fetch from GitHub API" >&2
        exit 1
    fi
    
    # Extract tag_name using grep and cut
    echo "$response" | grep -o '"tag_name":"[^"]*' | cut -d'"' -f4
}

# Show current version
latest_release=$(get_latest_release)
if [ -z "$latest_release" ]; then
    echo "Error getting latest release" >&2
    exit 1
fi
echo "Current latest release: $latest_release"

# Ask for new version
read -p "Enter new version (e.g., 1.0.4): " new_version

# Validate version format
if ! [[ $new_version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format X.Y.Z (e.g., 1.0.4)" >&2
    exit 1
fi

# Get script directory
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to script directory
cd "$script_dir" || { echo "Error: Could not change to script directory" >&2; exit 1; }

# hero_v_path="$script_dir/cli/hero.v"

# # Read hero.v and check if it exists
# if [ ! -f "$hero_v_path" ]; then
#     echo "Error: $hero_v_path does not exist" >&2
#     exit 1
# fi

# # Create backup of hero.v
# cp "$hero_v_path" "$hero_v_path.backup" || {
#     echo "Error creating backup of $hero_v_path" >&2
#     exit 1
# }

# # Update version in hero.v
# if ! sed -i.bak "s/\(.*version:[ ]*\)'[^']*'/\1'$new_version'/" "$hero_v_path"; then
#     echo "Error updating version in $hero_v_path" >&2
#     # Restore backup
#     cp "$hero_v_path.backup" "$hero_v_path" || echo "Error restoring backup" >&2
#     exit 1
# fi

# # Clean up backup
# rm -f "$hero_v_path.backup" "$hero_v_path.bak" || echo "Warning: Could not remove backup file" >&2

# # Update version in install_hero.sh
# install_hero_path="$script_dir/install_hero.sh"

# # Check if install_hero.sh exists
# if [ ! -f "$install_hero_path" ]; then
#     echo "Error: $install_hero_path does not exist" >&2
#     exit 1
# fi

# # Create backup of install_hero.sh
# cp "$install_hero_path" "$install_hero_path.backup" || {
#     echo "Error creating backup of $install_hero_path" >&2
#     exit 1
# }

# # Update version in install_hero.sh
# if ! sed -i.bak "s/version='[^']*'/version='$new_version'/" "$install_hero_path"; then
#     echo "Error updating version in $install_hero_path" >&2
#     # Restore backup
#     cp "$install_hero_path.backup" "$install_hero_path" || echo "Error restoring backup" >&2
#     exit 1
# fi

# # Clean up backup
# rm -f "$install_hero_path.backup" "$install_hero_path.bak" || echo "Warning: Could not remove backup file" >&2

# Prepare git commands
cmd="
git remote set-url origin git@github.com:freeflowuniverse/herolauncher.git
git add $hero_v_path $install_hero_path
git commit -m \"bump version to $new_version\"
git pull git@github.com:freeflowuniverse/herolauncher.git main
git tag -a \"v$new_version\" -m \"Release version $new_version\"
git push git@github.com:freeflowuniverse/herolauncher.git \"v$new_version\"
"

echo "$cmd"

# Execute git commands
eval "$cmd"

echo "Release v$new_version created and pushed!"
