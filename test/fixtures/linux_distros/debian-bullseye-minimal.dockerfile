FROM debian:11

# Test Dockerfile for Debian 11 - Issue #012 PowerShell testing
LABEL maintainer="portunix-testing"
LABEL issue="012"
LABEL distribution="debian-11"

# Prevent interactive prompts during apt installation
ENV DEBIAN_FRONTEND=noninteractive

# Update and install minimal packages required for PowerShell testing
RUN apt-get update && \
    apt-get install -y \
        sudo \
        wget \
        curl \
        lsb-release \
        openssh-server \
        ca-certificates \
        gnupg \
        software-properties-common && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Setup SSH server for testing
RUN mkdir -p /var/run/sshd && \
    echo 'root:testpass123' | chpasswd && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config

# Create workspace for portunix
RUN mkdir -p /workspace
WORKDIR /workspace

# Expose SSH port for testing
EXPOSE 22

# Default command for testing
CMD ["bash", "-c", "sleep 3600"]