FROM ubuntu:22.04

# Install required packages
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    git \
    golang \
    iproute2 \
    net-tools \
    procps \
    sudo \
    vim \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create a mount point for the 9P filesystem
RUN mkdir -p /mnt/vfs9p

# Add a script to mount the 9P filesystem
COPY mount-9p.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/mount-9p.sh

# Set the working directory
WORKDIR /app

CMD ["/bin/bash"]
