#!/bin/bash

# Configuration
ORGANISATION_NAME="dragonflydb"
PROVIDER_NAME="terraform-provider-dfcloud"
TF_PROVIDER_NAME="dfcloud"

# Fetch latest version from GitHub API
VERSION=$(curl -s "https://api.github.com/repos/${ORGANISATION_NAME}/${PROVIDER_NAME}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
if [ -z "$VERSION" ]; then
    echo "Error: Could not fetch latest version"
    exit 1
fi
echo "Latest version of the provider is: $VERSION"

OS="linux"
ARCH="amd64"

# Create necessary directories
PLUGIN_DIR="./.terraform/providers/registry.terraform.io/${ORGANISATION_NAME}/${TF_PROVIDER_NAME}/${VERSION}/${OS}_${ARCH}"
mkdir -p "$PLUGIN_DIR"

# Download URL construction
DOWNLOAD_URL="https://github.com/${ORGANISATION_NAME}/${PROVIDER_NAME}/releases/download/v${VERSION}/${PROVIDER_NAME}_${VERSION}_${OS}_${ARCH}.zip"

# Download and extract
echo "Downloading provider version ${VERSION} from: $DOWNLOAD_URL"
curl -L "$DOWNLOAD_URL" -o "${PLUGIN_DIR}/provider.zip"
cd "$PLUGIN_DIR"
unzip provider.zip
rm provider.zip

echo "Provider downloaded and installed successfully in: $PLUGIN_DIR"
