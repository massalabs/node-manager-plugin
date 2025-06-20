#!/bin/bash

# setup-node-unix.sh
# Script to setup Massa node binaries for Unix systems (Linux/macOS)
# Should be called by taskfile.
# 
# Usage: ./setup-node-unix.sh <MAINNET_VERSION> <BUILDNET_VERSION> <MAINNET_NODEBIN> <BUILDNET_NODEBIN> <NODE_MASSA_DIR>
# Example: ./setup-node-unix.sh "MAIN.3.0" "DEVN.28.3" "massa_MAIN.3.0_release_linux.tar.gz" "massa_DEVN.28.3_release_linux.tar.gz" "build/node-massa"

set -e  # Exit on any error

# Check if correct number of parameters provided
if [ $# -ne 5 ]; then
    echo "Error: This script requires exactly 5 parameters"
    echo "Usage: $0 <MAINNET_VERSION> <BUILDNET_VERSION> <MAINNET_NODEBIN> <BUILDNET_NODEBIN> <NODE_MASSA_DIR>"
    echo "Example: $0 \"MAIN.3.0\" \"DEVN.28.3\" \"massa_MAIN.3.0_release_linux.tar.gz\" \"massa_DEVN.28.3_release_linux.tar.gz\" \"build/node-massa\""
    exit 1
fi

# Parse parameters
MAINNET_VERSION="$1"
BUILDNET_VERSION="$2"
MAINNET_NODEBIN="$3"
BUILDNET_NODEBIN="$4"
NODE_MASSA_DIR="$5"

echo "Setting up Massa node binaries..."
echo "Mainnet version: $MAINNET_VERSION"
echo "Buildnet version: $BUILDNET_VERSION"
echo "Mainnet binary: $MAINNET_NODEBIN"
echo "Buildnet binary: $BUILDNET_NODEBIN"
echo "Node Massa directory: $NODE_MASSA_DIR"

# Create necessary directories
if [ ! -d "$NODE_MASSA_DIR" ]; then
    echo "Creating directories..."
    mkdir -p "$NODE_MASSA_DIR"
fi


# Download and extract mainnet binary
echo "Downloading mainnet binary..."
curl -Ls -o "$MAINNET_NODEBIN" "https://github.com/massalabs/massa/releases/download/$MAINNET_VERSION/$MAINNET_NODEBIN"

echo "Extracting mainnet binary..."
mkdir -p "$NODE_MASSA_DIR"/"$MAINNET_VERSION"
tar -xzf "$MAINNET_NODEBIN"
rm -rf "$NODE_MASSA_DIR"/"$MAINNET_VERSION"/*
mv massa/* "$NODE_MASSA_DIR"/"$MAINNET_VERSION"
rm -r massa/ "$MAINNET_NODEBIN"

# Download and extract buildnet binary
echo "Downloading buildnet binary..."
curl -Ls -o "$BUILDNET_NODEBIN" "https://github.com/massalabs/massa/releases/download/$BUILDNET_VERSION/$BUILDNET_NODEBIN"

echo "Extracting buildnet binary..."
mkdir -p "$NODE_MASSA_DIR"/"$BUILDNET_VERSION"
tar -xzf "$BUILDNET_NODEBIN"
mv massa/* "$NODE_MASSA_DIR"/"$BUILDNET_VERSION"
rm -r massa/ "$BUILDNET_NODEBIN"