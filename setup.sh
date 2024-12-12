#!/bin/bash

# Configuration
ORGANISATION_NAME="dragonflydb"
PROVIDER_NAME="terraform-provider-dfcloud"
TF_PROVIDER_NAME="dfcloud"

VERSION="0.0.1"
OS="linux"
ARCH="amd64"

# Create necessary directories
PLUGIN_DIR="./.terraform/providers/registry.terraform.io/${ORGANISATION_NAME}/${TF_PROVIDER_NAME}/${VERSION}/${OS}_${ARCH}"
mkdir -p "$PLUGIN_DIR"

# Download URL construction
DOWNLOAD_URL="https://github.com/${ORGANISATION_NAME}/${PROVIDER_NAME}/releases/download/v${VERSION}/${PROVIDER_NAME}_${VERSION}_${OS}_${ARCH}.zip"

# Download and extract
echo "Downloading provider from: $DOWNLOAD_URL"
curl -L "$DOWNLOAD_URL" -o "${PLUGIN_DIR}/provider.zip"
cd "$PLUGIN_DIR"
unzip provider.zip
rm provider.zip

echo "Provider downloaded and installed successfully in: $PLUGIN_DIR"
