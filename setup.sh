#!/bin/bash

# Configuration
ORGANIZATION_NAME="dragonflydb"
PROVIDER_NAME="terraform-provider-dfcloud"
TF_PROVIDER_NAME="dfcloud"

# Fetch latest version from GitHub API
VERSION=$(curl -s "https://api.github.com/repos/${ORGANIZATION_NAME}/${PROVIDER_NAME}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
if [ -z "$VERSION" ]; then
  echo "Error: Could not fetch latest version"
  exit 1
fi
echo "Latest version of the provider is: $VERSION"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Normalize OS name
case "$OS" in
  linux)
  OS="linux"
  ;;
  darwin)
  OS="darwin"
  ;;
  *)
  echo "Unsupported OS: $OS"
  exit 1
  ;;
esac

# Normalize architecture name
case "$ARCH" in
  x86_64)
  ARCH="amd64"
  ;;
  arm64|aarch64)
  ARCH="arm64"
  ;;
  *)
  echo "Unsupported architecture: $ARCH"
  exit 1
  ;;
esac

# Installation directories
DEV_OVERRIDES_PATH="$HOME/bin"
PLUGIN_DIR="$DEV_OVERRIDES_PATH"

# Create necessary directories
mkdir -p "$PLUGIN_DIR"

# Download URL construction
DOWNLOAD_URL="https://github.com/${ORGANIZATION_NAME}/${PROVIDER_NAME}/releases/download/v${VERSION}/${PROVIDER_NAME}_${VERSION}_${OS}_${ARCH}.zip"

# Download and extract
echo "Downloading provider version ${VERSION} from: $DOWNLOAD_URL"
curl -L "$DOWNLOAD_URL" -o "$PLUGIN_DIR/provider.zip"
unzip -o "$PLUGIN_DIR/provider.zip" -d "$PLUGIN_DIR"
rm "$PLUGIN_DIR/provider.zip"

# Ensure the binary is executable
chmod +x "$PLUGIN_DIR/terraform-provider-${TF_PROVIDER_NAME}_v${VERSION}"

# Configure ~/.terraformrc with dev_overrides
TERRAFORMRC="$HOME/.terraformrc"

# Backup existing ~/.terraformrc if it exists
if [ -f "$TERRAFORMRC" ]; then
  echo "Backing up existing ~/.terraformrc to ~/.terraformrc.bak"
  cp "$TERRAFORMRC" "$TERRAFORMRC.bak"
fi

echo "Updating ~/.terraformrc with dev_overrides"
cat > "$TERRAFORMRC" <<EOF
provider_installation {
  dev_overrides {
  "registry.terraform.io/${ORGANIZATION_NAME}/${TF_PROVIDER_NAME}" = "$DEV_OVERRIDES_PATH"
  }
  direct {}
}
EOF

echo "Provider downloaded and installed successfully in: $PLUGIN_DIR"
