#!/bin/bash
set -e # Exit immediately if a command exits with a non-zero status.

# --- Configuration ---
# Required Environment Variables:
# SERVER:   IPv4 or IPv6 address of the target Hetzner server (already in Rescue Mode).
# HOSTNAME: The desired hostname for the installed system.
# Drives are now always auto-detected by the installer binary.

LOG_FILE="hetzner_install_$(date +%Y%m%d_%H%M%S).log"
REMOTE_USER="root" # Hetzner Rescue Mode typically uses root
REMOTE_DIR="/tmp/hetzner_installer_$$" # Temporary directory on the remote server
BINARY_NAME="hetzner_installer"
BUILD_DIR="build"

# --- Helper Functions ---
log() {
    local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    echo "[$timestamp] $1" | tee -a "$LOG_FILE"
}

cleanup_remote() {
    if [ -n "$SERVER" ]; then
        log "Cleaning up remote directory $REMOTE_DIR on $SERVER..."
        ssh "$REMOTE_USER@$SERVER" "rm -rf $REMOTE_DIR" || log "Warning: Failed to clean up remote directory (might be okay if server rebooted)."
    fi
}

# --- Main Script ---
cd "$(dirname "$0")"

log "=== Starting Hetzner Installimage Deployment ==="
log "Log file: $LOG_FILE"
log "IMPORTANT: Ensure the target server ($SERVER) is booted into Hetzner Rescue Mode!"

# Check required environment variables
if [ -z "$SERVER" ]; then
    log "❌ ERROR: SERVER environment variable is not set."
    log "Please set it to the IP address of the target server (in Rescue Mode)."
    exit 1
fi
if [ -z "$HOSTNAME" ]; then
    log "❌ ERROR: HOSTNAME environment variable is not set."
	log "Please set it to the desired hostname for the installed system."
	exit 1
fi
# Drives are auto-detected by the binary.

# Validate SERVER IP (basic check)
if ! [[ "$SERVER" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]] && \
   ! [[ "$SERVER" =~ ^[0-9a-fA-F:]+$ ]]; then
    log "❌ ERROR: SERVER ($SERVER) does not look like a valid IPv4 or IPv6 address."
    exit 1
fi

log "Target Server: $SERVER"
log "Target Hostname: $HOSTNAME"
log "Target Drives: Auto-detected by the installer."

# Build the Hetzner installer binary
log "Building $BINARY_NAME binary..."
./build.sh | tee -a "$LOG_FILE"

# Check if binary exists
BINARY_PATH="$BUILD_DIR/$BINARY_NAME"
if [ ! -f "$BINARY_PATH" ]; then
    log "❌ ERROR: $BINARY_NAME binary not found at $BINARY_PATH after build."
    exit 1
fi

log "Binary size:"
ls -lh "$BINARY_PATH" | tee -a "$LOG_FILE"

# Set up trap for cleanup
trap cleanup_remote EXIT

# Create deployment directory on server
log "Creating temporary directory $REMOTE_DIR on server..."
# Use -t to force pseudo-terminal allocation for mkdir (less critical but consistent)
ssh -t "$REMOTE_USER@$SERVER" "mkdir -p $REMOTE_DIR" 2>&1 | tee -a "$LOG_FILE"
if [ $? -ne 0 ]; then
    log "❌ ERROR: Failed to create remote directory $REMOTE_DIR on $SERVER."
    exit 1
fi

# Transfer the binary to the server
log "Transferring $BINARY_NAME binary to $SERVER:$REMOTE_DIR/ ..."
rsync -avz --progress "$BINARY_PATH" "$REMOTE_USER@$SERVER:$REMOTE_DIR/" 2>&1 | tee -a "$LOG_FILE"
if [ $? -ne 0 ]; then
    log "❌ ERROR: Failed to transfer binary to $SERVER."
    exit 1
fi

# Ensure binary is executable on the server
log "Setting permissions on server..."
# Use -t
ssh -t "$REMOTE_USER@$SERVER" "chmod +x $REMOTE_DIR/$BINARY_NAME" 2>&1 | tee -a "$LOG_FILE" || { log "❌ ERROR: Failed to set permissions on remote binary."; exit 1; }
# Use -t
ssh -t "$REMOTE_USER@$SERVER" "ls -la $REMOTE_DIR/" 2>&1 | tee -a "$LOG_FILE"

# Construct remote command arguments (only hostname needed now)
# Note: The binary expects -hostname
REMOTE_CMD_ARGS="-hostname \"$HOSTNAME\""

# Run the Hetzner installer (Go binary) on the server
log "Running Go installer binary $BINARY_NAME on server $SERVER..."
REMOTE_FULL_CMD="cd $REMOTE_DIR && ./$BINARY_NAME $REMOTE_CMD_ARGS"
log "Command: $REMOTE_FULL_CMD"

# Execute the command and capture output. Use -t for better output.
INSTALL_OUTPUT=$(ssh -t "$REMOTE_USER@$SERVER" "$REMOTE_FULL_CMD" 2>&1)
INSTALL_EXIT_CODE=$?

log "--- Go Installer Binary Output ---"
echo "$INSTALL_OUTPUT" | tee -a "$LOG_FILE"
log "--- End Go Installer Binary Output ---"
log "Go installer binary exit code: $INSTALL_EXIT_CODE"

# Analyze results - relies on Go binary output now
if [[ "$INSTALL_OUTPUT" == *"installimage command finished. System should reboot shortly if successful."* ]]; then
    log "✅ SUCCESS: Go installer reported successful initiation. The server should be rebooting into the new OS."
    log "Verification of the installed OS must be done manually after reboot."
elif [[ "$INSTALL_OUTPUT" == *"Error during Hetzner installation"* || $INSTALL_EXIT_CODE -ne 0 ]]; then
    log "❌ ERROR: Go installer reported an error or exited with code $INSTALL_EXIT_CODE."
    log "Check the output above for details. Common issues include installimage errors or config problems."
    # Don't exit immediately, allow cleanup trap to run
else
    # This might happen if the SSH connection is abruptly closed by the reboot during installimage
    log "⚠️ WARNING: The Go installer finished with exit code $INSTALL_EXIT_CODE, but the output might be incomplete due to server reboot."
    log "Assuming the installimage process was initiated. Manual verification is required after reboot."
fi

log "=== Hetzner Installimage Deployment Script Finished ==="
# Cleanup trap will run on exit
