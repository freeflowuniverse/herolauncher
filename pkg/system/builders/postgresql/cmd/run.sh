#!/bin/bash
set -e


export SERVER="65.109.18.183"
LOG_FILE="postgresql_deployment_$(date +%Y%m%d_%H%M%S).log"

cd "$(dirname "$0")"

# Configure logging
log() {
    local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    echo "[$timestamp] $1" | tee -a "$LOG_FILE"
}

log "=== Starting PostgreSQL Builder Deployment ==="
log "Log file: $LOG_FILE"

# Check if SERVER environment variable is set
if [ -z "$SERVER" ]; then
    log "Error: SERVER environment variable is not set."
    log "Please set it to the IPv4 or IPv6 address of the target server."
    log "Example: export SERVER=192.168.1.100"
    exit 1
fi

# Validate if SERVER is a valid IP address (IPv4 or IPv6)
if ! [[ "$SERVER" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]] && \
   ! [[ "$SERVER" =~ ^[0-9a-fA-F:]+$ ]]; then
    log "Error: SERVER must be a valid IPv4 or IPv6 address."
    exit 1
fi

log "Using server: $SERVER"

# Build the PostgreSQL builder binary
log "Building PostgreSQL builder binary..."
./build.sh | tee -a "$LOG_FILE"

# Check if binary exists
if [ ! -f "build/postgresql_builder" ]; then
    log "Error: PostgreSQL builder binary not found after build."
    exit 1
fi

log "Binary size:"
ls -lh build/ | tee -a "$LOG_FILE"

# Create deployment directory on server
log "Creating deployment directory on server..."
ssh "root@$SERVER" "mkdir -p ~/postgresql_builder" 2>&1 | tee -a "$LOG_FILE"

# Transfer the binary to the server
log "Transferring PostgreSQL builder binary to server..."
rsync -avz --progress build/postgresql_builder "root@$SERVER:~/postgresql_builder/" 2>&1 | tee -a "$LOG_FILE"


# Run the PostgreSQL builder on the server
log "Running PostgreSQL builder on server..."
ssh -t "root@$SERVER" "cd ~/postgresql_builder && ./postgresql_builder" 2>&1 | tee -a "$LOG_FILE"
BUILD_EXIT_CODE=${PIPESTATUS[0]}

# If there was an error, make it very clear
if [ $BUILD_EXIT_CODE -ne 0 ]; then
    log "⚠️  PostgreSQL builder failed with exit code: $BUILD_EXIT_CODE"
fi

# Check for errors in exit code
if [ $BUILD_EXIT_CODE -eq 0 ]; then
    log "✅ SUCCESS: PostgreSQL builder completed successfully!"
    log "----------------------------------------------------------------"
    
    # Note: Verification is now handled by the builder itself
    
    # Check for build logs or error messages
    log "Checking for build logs on server..."
    BUILD_LOGS=$(ssh "root@$SERVER" "cd ~/postgresql_builder && ls -la *.log 2>/dev/null || echo 'No log files found'" 2>&1)
    log "Build log files:"
    echo "$BUILD_LOGS" | tee -a "$LOG_FILE"
    
    log "----------------------------------------------------------------"
    log "🎉 PostgreSQL Builder deployment COMPLETED"
    log "================================================================"
else
    log "❌ ERROR: PostgreSQL builder failed to run properly on the server."
    
    # Get more detailed error information
    # log "Checking for error logs on server..."
    # ssh "root@$SERVER" "cd ~/postgresql_builder && ls -la" 2>&1 | tee -a "$LOG_FILE"    
    exit 1
fi

log "=== Deployment Completed ==="
