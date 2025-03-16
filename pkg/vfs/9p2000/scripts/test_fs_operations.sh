#!/bin/bash

# Comprehensive 9p filesystem operations test script
# Tests various file operations with different file types and sizes up to 1MB

set -e  # Exit on error

# Configuration
MOUNT_POINT="/mnt/myvfs"
SERVER_IP="127.0.0.1"  # Change this to the actual server IP if needed
SERVER_PORT="9999"
TEST_DIR="${MOUNT_POINT}/test_$(date +%s)"
LOG_FILE="9p_test_results.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Initialize log file
echo "9p Filesystem Test Results - $(date)" > $LOG_FILE
echo "=======================================" >> $LOG_FILE

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
    echo "[INFO] $1" >> $LOG_FILE
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    echo "[SUCCESS] $1" >> $LOG_FILE
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
    echo "[WARNING] $1" >> $LOG_FILE
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    echo "[ERROR] $1" >> $LOG_FILE
}

run_test() {
    local test_name=$1
    local cmd=$2
    
    echo -e "\n${BLUE}Running test:${NC} $test_name"
    echo -e "Command: $cmd"
    echo -e "\n[TEST] $test_name" >> $LOG_FILE
    echo "Command: $cmd" >> $LOG_FILE
    
    if eval $cmd >> $LOG_FILE 2>&1; then
        log_success "$test_name passed"
        return 0
    else
        log_error "$test_name failed"
        return 1
    fi
}

# Check if the mount point exists and is mounted
check_mount() {
    log_info "Checking if $MOUNT_POINT is mounted..."
    if mount | grep -q "$MOUNT_POINT"; then
        log_success "$MOUNT_POINT is mounted"
    else
        log_error "$MOUNT_POINT is not mounted. Please mount the 9p filesystem first."
        log_info "You can mount it with: sudo mount -t 9p -o version=9p2000,trans=tcp,uname=nobody $SERVER_IP:$SERVER_PORT $MOUNT_POINT"
        exit 1
    fi
}

# Create test directory structure
create_test_dirs() {
    log_info "Creating test directory structure at $TEST_DIR"
    mkdir -p $TEST_DIR
    mkdir -p $TEST_DIR/dir1
    mkdir -p $TEST_DIR/dir2
    mkdir -p $TEST_DIR/dir1/subdir1
    mkdir -p $TEST_DIR/dir2/subdir2
    log_success "Test directories created"
}

# Test 1: Basic file operations
test_basic_file_ops() {
    log_info "Testing basic file operations..."
    
    # Create a small text file
    run_test "Create small text file" "echo 'This is a test file' > $TEST_DIR/test.txt"
    
    # Read the file
    run_test "Read small text file" "cat $TEST_DIR/test.txt | grep -q 'This is a test file'"
    
    # Append to the file
    run_test "Append to text file" "echo 'Additional line' >> $TEST_DIR/test.txt"
    
    # Verify append
    run_test "Verify append" "cat $TEST_DIR/test.txt | grep -q 'Additional line'"
    
    # Copy the file
    run_test "Copy file" "cp $TEST_DIR/test.txt $TEST_DIR/test_copy.txt"
    
    # Move/rename the file
    run_test "Move/rename file" "mv $TEST_DIR/test_copy.txt $TEST_DIR/test_renamed.txt"
    
    # Delete the file
    run_test "Delete file" "rm $TEST_DIR/test_renamed.txt"
    
    log_success "Basic file operations tests completed"
}

# Test 2: Directory operations
test_directory_ops() {
    log_info "Testing directory operations..."
    
    # Create nested directories
    run_test "Create nested directory" "mkdir -p $TEST_DIR/nested/dir1/dir2"
    
    # Create files in nested directories
    run_test "Create file in nested directory" "echo 'Nested file' > $TEST_DIR/nested/dir1/dir2/nested.txt"
    
    # List directory contents
    run_test "List directory contents" "ls -la $TEST_DIR/nested/dir1/dir2 | grep -q nested.txt"
    
    # Move directory
    run_test "Move directory" "mv $TEST_DIR/nested/dir1/dir2 $TEST_DIR/nested/dir1/dir2_moved"
    
    # Verify moved directory
    run_test "Verify moved directory" "ls -la $TEST_DIR/nested/dir1 | grep -q dir2_moved"
    
    # Remove directory recursively
    run_test "Remove directory recursively" "rm -rf $TEST_DIR/nested"
    
    log_success "Directory operations tests completed"
}

# Test 3: File types and permissions
test_file_types_permissions() {
    log_info "Testing file types and permissions..."
    
    # Create files with different permissions
    run_test "Create executable file" "echo '#!/bin/bash\necho \"Hello\"' > $TEST_DIR/executable.sh && chmod +x $TEST_DIR/executable.sh"
    
    # Verify permissions
    run_test "Verify executable permissions" "ls -la $TEST_DIR/executable.sh | grep -q 'x'"
    
    # Create a symbolic link
    run_test "Create symbolic link" "ln -s $TEST_DIR/executable.sh $TEST_DIR/symlink"
    
    # Verify symbolic link
    run_test "Verify symbolic link" "ls -la $TEST_DIR | grep -q 'symlink -> '"
    
    # Create a hard link
    run_test "Create hard link" "ln $TEST_DIR/executable.sh $TEST_DIR/hardlink"
    
    # Verify hard link
    run_test "Verify hard link" "ls -la $TEST_DIR/hardlink | grep -q 'hardlink'"
    
    log_success "File types and permissions tests completed"
}

# Test 4: File sizes
test_file_sizes() {
    log_info "Testing various file sizes..."
    
    # Create small file (1KB)
    run_test "Create 1KB file" "dd if=/dev/urandom of=$TEST_DIR/1kb.bin bs=1K count=1"
    
    # Create medium file (100KB)
    run_test "Create 100KB file" "dd if=/dev/urandom of=$TEST_DIR/100kb.bin bs=1K count=100"
    
    # Create large file (1MB)
    run_test "Create 1MB file" "dd if=/dev/urandom of=$TEST_DIR/1mb.bin bs=1M count=1"
    
    # Verify file sizes
    run_test "Verify 1KB file size" "ls -la $TEST_DIR/1kb.bin | awk '{print \$5}' | grep -q '1024'"
    run_test "Verify 100KB file size" "ls -la $TEST_DIR/100kb.bin | awk '{print \$5}' | grep -q '102400'"
    run_test "Verify 1MB file size" "ls -la $TEST_DIR/1mb.bin | awk '{print \$5}' | grep -q '1048576'"
    
    log_success "File size tests completed"
}

# Test 5: File operations with different sizes
test_file_operations_with_sizes() {
    log_info "Testing file operations with different sizes..."
    
    # Copy large file
    run_test "Copy 1MB file" "cp $TEST_DIR/1mb.bin $TEST_DIR/1mb_copy.bin"
    
    # Verify copied file
    run_test "Verify copied 1MB file" "diff $TEST_DIR/1mb.bin $TEST_DIR/1mb_copy.bin"
    
    # Append to medium file
    run_test "Append to 100KB file" "dd if=/dev/urandom bs=1K count=10 >> $TEST_DIR/100kb.bin"
    
    # Verify appended file size
    run_test "Verify appended file size" "ls -la $TEST_DIR/100kb.bin | awk '{print \$5}' | grep -q '112640'"
    
    # Read chunks of large file
    run_test "Read chunks of 1MB file" "dd if=$TEST_DIR/1mb.bin of=/dev/null bs=4K"
    
    # Truncate large file
    run_test "Truncate 1MB file" "truncate -s 512K $TEST_DIR/1mb.bin"
    
    # Verify truncated file
    run_test "Verify truncated file" "ls -la $TEST_DIR/1mb.bin | awk '{print \$5}' | grep -q '524288'"
    
    log_success "File operations with different sizes tests completed"
}

# Test 6: Concurrent operations
test_concurrent_operations() {
    log_info "Testing concurrent operations..."
    
    # Create multiple files concurrently
    run_test "Create multiple files concurrently" "for i in {1..10}; do echo 'Concurrent file $i' > $TEST_DIR/concurrent_$i.txt & done; wait"
    
    # Read multiple files concurrently
    run_test "Read multiple files concurrently" "for i in {1..10}; do cat $TEST_DIR/concurrent_$i.txt > /dev/null & done; wait"
    
    # Delete multiple files concurrently
    run_test "Delete multiple files concurrently" "for i in {1..10}; do rm $TEST_DIR/concurrent_$i.txt & done; wait"
    
    log_success "Concurrent operations tests completed"
}

# Test 7: Special characters in filenames
test_special_chars() {
    log_info "Testing special characters in filenames..."
    
    # Create files with spaces
    run_test "Create file with spaces" "echo 'File with spaces' > '$TEST_DIR/file with spaces.txt'"
    
    # Create file with special characters
    run_test "Create file with special characters" "echo 'Special chars' > '$TEST_DIR/file-with_special@#$%^&()chars.txt'"
    
    # Verify files with special characters
    run_test "Verify file with spaces" "cat '$TEST_DIR/file with spaces.txt' | grep -q 'File with spaces'"
    run_test "Verify file with special characters" "cat '$TEST_DIR/file-with_special@#$%^&()chars.txt' | grep -q 'Special chars'"
    
    # Delete files with special characters
    run_test "Delete file with spaces" "rm '$TEST_DIR/file with spaces.txt'"
    run_test "Delete file with special characters" "rm '$TEST_DIR/file-with_special@#$%^&()chars.txt'"
    
    log_success "Special characters in filenames tests completed"
}

# Test 8: File content verification
test_file_content() {
    log_info "Testing file content verification..."
    
    # Create file with specific pattern
    run_test "Create file with pattern" "for i in {1..1000}; do echo \"Line $i with some random text\" >> $TEST_DIR/pattern.txt; done"
    
    # Search for pattern
    run_test "Search for pattern" "grep -q 'Line 500' $TEST_DIR/pattern.txt"
    
    # Count lines
    run_test "Count lines" "wc -l $TEST_DIR/pattern.txt | grep -q '1000'"
    
    # Create binary file with pattern
    run_test "Create binary file with pattern" "dd if=/dev/zero bs=1024 count=100 | tr '\\000' '\\377' > $TEST_DIR/binary_pattern.bin"
    
    # Verify binary pattern
    run_test "Verify binary pattern" "hexdump -C $TEST_DIR/binary_pattern.bin | head -n 2 | grep -q 'ff ff ff ff'"
    
    log_success "File content verification tests completed"
}

# Clean up
cleanup() {
    if [ "$1" != "--no-cleanup" ]; then
        log_info "Cleaning up test files and directories..."
        rm -rf $TEST_DIR
        log_success "Cleanup completed"
    else
        log_warning "Skipping cleanup. Test files remain at $TEST_DIR"
    fi
}

# Main function
main() {
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   9p Filesystem Operations Test Suite   ${NC}"
    echo -e "${BLUE}=========================================${NC}"
    
    # Check if mount point is mounted
    check_mount
    
    # Create test directory structure
    create_test_dirs
    
    # Run tests
    test_basic_file_ops
    test_directory_ops
    test_file_types_permissions
    test_file_sizes
    test_file_operations_with_sizes
    test_concurrent_operations
    test_special_chars
    test_file_content
    
    # Clean up
    if [ "$1" == "--no-cleanup" ]; then
        cleanup "--no-cleanup"
    else
        cleanup
    fi
    
    echo -e "\n${GREEN}All tests completed!${NC}"
    echo -e "See $LOG_FILE for detailed results."
}

# Parse command line arguments
if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  --no-cleanup    Don't remove test files after running tests"
    echo "  --help, -h      Show this help message"
    exit 0
fi

# Run the main function
main "$1"
