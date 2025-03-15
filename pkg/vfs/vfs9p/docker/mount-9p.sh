#!/bin/bash

# Check if the mount point exists, create it if it doesn't
if [ ! -d /mnt/vfs9p ]; then
    mkdir -p /mnt/vfs9p
fi

# Mount the 9P filesystem
# The host.docker.internal hostname resolves to the host machine's IP address
mount -t 9p -o trans=tcp,port=5640,version=9p2000 host.docker.internal /mnt/vfs9p

echo "9P filesystem mounted at /mnt/vfs9p"
echo "You can now access the files using: ls -la /mnt/vfs9p"
