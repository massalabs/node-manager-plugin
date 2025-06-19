#!/bin/bash

# ci-setup-node-from-linux.sh
# Should be called by taskfile and in github action workflow ci.
# Script to setup windows Massa node binaries in Linux environment
# Downloads zip archives and extracts massa-node and massa-client from the massa/ folder
# 
# Usage: ./ci-setup-node-from-linux.sh <MAINNET_VERSION> <BUILDNET_VERSION> <NODE_MASSA_DIR>
# Example: ./ci-setup-node-from-linux.sh "MAIN.3.0" "DEVN.28.3" "build/node-massa"

set -e  # Exit on any error

# Check if correct number of parameters provided
if [ $# -ne 3 ]; then
    echo "Error: This script requires exactly 3 parameters"
    echo "Usage: $0 <MAINNET_VERSION> <BUILDNET_VERSION> <NODE_MASSA_DIR>"
    echo "Example: $0 \"MAIN.3.0\" \"DEVN.28.3\" \"build/node-massa\""
    exit 1
fi

# Parse parameters
MAINNET_VERSION="$1"
BUILDNET_VERSION="$2"
NODE_MASSA_DIR="$3"

# Generate archive names
MAINNET_NODEBIN="massa_${MAINNET_VERSION}_release_windows.zip"
BUILDNET_NODEBIN="massa_${BUILDNET_VERSION}_release_windows.zip"

echo "Setting up Massa node binaries from Linux ..."
echo "Mainnet version: $MAINNET_VERSION"
echo "Buildnet version: $BUILDNET_VERSION"
echo "Node Massa directory: $NODE_MASSA_DIR"
echo "Mainnet binary: $MAINNET_NODEBIN"
echo "Buildnet binary: $BUILDNET_NODEBIN"

# Create necessary directories
if [ ! -d "$NODE_MASSA_DIR" ]; then
    echo "Creating directories..."
    mkdir -p "$NODE_MASSA_DIR"
fi

# Download and extract mainnet binary
echo "Downloading mainnet binary..."
curl -Ls -o "$MAINNET_NODEBIN" "https://github.com/massalabs/massa/releases/download/$MAINNET_VERSION/$MAINNET_NODEBIN"

echo "Extracting mainnet binary..."
mkdir -p "$NODE_MASSA_DIR/$MAINNET_VERSION"
unzip -q "$MAINNET_NODEBIN"
rm -rf "$NODE_MASSA_DIR/$MAINNET_VERSION"/*
mv massa/* "$NODE_MASSA_DIR/$MAINNET_VERSION/"
rm -rf massa "$MAINNET_NODEBIN"


# Download and extract buildnet binary
echo "Downloading buildnet binary..."
curl -Ls -o "$BUILDNET_NODEBIN" "https://github.com/massalabs/massa/releases/download/$BUILDNET_VERSION/$BUILDNET_NODEBIN"

echo "Extracting buildnet binary..."
mkdir -p "$NODE_MASSA_DIR/$BUILDNET_VERSION"
unzip -q "$BUILDNET_NODEBIN"
rm -rf "$NODE_MASSA_DIR/$BUILDNET_VERSION"/*
mv massa/* "$NODE_MASSA_DIR/$BUILDNET_VERSION/"
rm -rf massa "$BUILDNET_NODEBIN"
