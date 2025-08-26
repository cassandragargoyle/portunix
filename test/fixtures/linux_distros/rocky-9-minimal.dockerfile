FROM rockylinux:9

# Test Dockerfile for Rocky Linux 9 - Issue #012 PowerShell testing
LABEL maintainer="portunix-testing"
LABEL issue="012"
LABEL distribution="rocky-9"

# Update and install minimal packages required for PowerShell testing
RUN dnf update -y && \
    dnf install -y \
        sudo \
        curl \
        wget \
        openssh-server \
        ca-certificates \
        gnupg \
        rpm \
        dnf-plugins-core && \
    dnf clean all

# Setup SSH server for testing
RUN ssh-keygen -A && \
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