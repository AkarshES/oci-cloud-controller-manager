#!/bin/bash
set -e

# Get minor version from command line argument
if [ -z "$1" ]; then
  echo "Usage: $0 <minor_version>"
  echo "Example: $0 1.21"
  exit 1
fi

MINOR_VERSION=$1
INSTALL_DIR=${2:-"/usr/local"}
ARCH=$(uname -m)
OS="linux"

if [ "$ARCH" == "x86_64" ]; then
  ARCH="amd64"
elif [ "$ARCH" == "aarch64" ]; then
  ARCH="arm64"
fi

echo "Looking for latest patch version for Go $MINOR_VERSION for $OS-$ARCH..."

TEMP_FILE=$(mktemp)
curl -s "https://go.dev/dl/?mode=json&include=all" > "$TEMP_FILE"

VERSIONS=$(jq -r ".[] | select(.version | startswith(\"go$MINOR_VERSION.\")) | .version" "$TEMP_FILE" | sort -V)
LATEST_VERSION=$(echo "$VERSIONS" | tail -n 1)

if [ -z "$LATEST_VERSION" ]; then
  echo "No version found for Go $MINOR_VERSION"
  rm "$TEMP_FILE"
  exit 1
fi

echo "Latest version found: $LATEST_VERSION"
DOWNLOAD_URL="https://go.dev/dl/${LATEST_VERSION}.$OS-$ARCH.tar.gz"
FILENAME="${LATEST_VERSION}.$OS-$ARCH.tar.gz"

echo "Downloading $DOWNLOAD_URL..."
wget -q "$DOWNLOAD_URL" -O "$FILENAME" || curl -L -o "$FILENAME" "$DOWNLOAD_URL"

echo "Checking for existing Go installations..."
if [ -d "$INSTALL_DIR/go" ]; then
  echo "Removing existing Go installation from $INSTALL_DIR/go"
  rm -rf "$INSTALL_DIR/go"
fi

echo "Installing Go $LATEST_VERSION to $INSTALL_DIR..."
tar -C "$INSTALL_DIR" -xzf "$FILENAME"

# Generate environment setup script
ENV_SCRIPT="/tmp/go_env_setup.sh"
cat > "$ENV_SCRIPT" << EOF
#!/bin/bash
export GOROOT="$INSTALL_DIR/go"
export PATH="\$GOROOT/bin:\$PATH"
export GOPATH="\${GOPATH:-\$HOME/go}"
export PATH="\$GOPATH/bin:\$PATH"
EOF

chmod +x "$ENV_SCRIPT"
echo "Environment setup script created at $ENV_SCRIPT"

# Verify installation
GO_VERSION=$("$INSTALL_DIR/go/bin/go" version)
echo "Go installation complete: $GO_VERSION"

rm "$TEMP_FILE"
rm "$FILENAME"

echo "To use Go in other steps, source: source /tmp/go_env_setup.sh"
