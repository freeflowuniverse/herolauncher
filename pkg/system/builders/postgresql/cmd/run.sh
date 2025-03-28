#!/bin/bash
set -ex


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

# Note: Dependencies are now installed by the builder itself

# Transfer the binary to the server
log "Transferring PostgreSQL builder binary to server..."
rsync -avz --progress build/postgresql_builder "root@$SERVER:~/postgresql_builder/" 2>&1 | tee -a "$LOG_FILE"

# Ensure binary is executable
log "Setting permissions on server..."
ssh "root@$SERVER" "chmod +x ~/postgresql_builder/postgresql_builder" 2>&1 | tee -a "$LOG_FILE"
ssh "root@$SERVER" "ls -la ~/postgresql_builder/" 2>&1 | tee -a "$LOG_FILE"

# Run the PostgreSQL builder on the server
log "Running PostgreSQL builder on server..."
BUILD_OUTPUT=$(ssh "root@$SERVER" "cd ~/postgresql_builder && ./postgresql_builder" 2>&1)
BUILD_EXIT_CODE=$?
echo "$BUILD_OUTPUT" | tee -a "$LOG_FILE"

# Check for errors in output or exit code
if [[ $BUILD_EXIT_CODE -eq 0 && ! "$BUILD_OUTPUT" =~ "Error" ]]; then
    log "✅ SUCCESS: PostgreSQL builder completed successfully!"
    log "----------------------------------------------------------------"
    
    # Verify PostgreSQL installation
    log "Verifying PostgreSQL installation on server..."
    
    # Check for PostgreSQL binary
    POSTGRES_CHECK=$(ssh "root@$SERVER" "ls -la /opt/postgresql/bin/postgres 2>/dev/null || echo 'Not found'" 2>&1)
    log "PostgreSQL binary check:"
    echo "$POSTGRES_CHECK" | tee -a "$LOG_FILE"
    
    if [[ "$POSTGRES_CHECK" == *"Not found"* ]]; then
        log "❌ WARNING: PostgreSQL binary not found at expected location."
        log "This may indicate that the build process failed or installed to a different location."
        
        # Search for PostgreSQL binary in other locations
        log "Searching for PostgreSQL binary in other locations..."
        FIND_POSTGRES=$(ssh "root@$SERVER" "find / -name postgres -type f 2>/dev/null || echo 'Not found'" 2>&1)
        log "Search results:"
        echo "$FIND_POSTGRES" | tee -a "$LOG_FILE"
    else
        log "✅ PostgreSQL binary found at expected location."
    fi
    
    # Check for Go stored procedure
    GO_SP_CHECK=$(ssh "root@$SERVER" "ls -la /opt/postgresql/lib/libgosp.so 2>/dev/null || echo 'Not found'" 2>&1)
    log "Go stored procedure check:"
    echo "$GO_SP_CHECK" | tee -a "$LOG_FILE"
    
    if [[ "$GO_SP_CHECK" == *"Not found"* ]]; then
        log "❌ WARNING: Go stored procedure library not found at expected location."
        
        # Search for Go stored procedure in other locations
        log "Searching for Go stored procedure in other locations..."
        FIND_GOSP=$(ssh "root@$SERVER" "find / -name libgosp.so -type f 2>/dev/null || echo 'Not found'" 2>&1)
        log "Search results:"
        echo "$FIND_GOSP" | tee -a "$LOG_FILE"
    else
        log "✅ Go stored procedure library found at expected location."
    fi
    
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
    log "Checking for error logs on server..."
    ssh "root@$SERVER" "cd ~/postgresql_builder && ls -la" 2>&1 | tee -a "$LOG_FILE"
    
    log "----------------------------------------------------------------"
    log "📋 POSSIBLE FAILURE REASONS:"
    log "  1. ⚙️ Missing dependencies on the server"
    log "  2. 🚀 Insufficient permissions"
    log "  3. 🔒 Network connectivity issues during download"
    log "  4. 🔑 Compilation errors"
    log "================================================================"
    exit 1
fi

log "=== Deployment Completed ==="
