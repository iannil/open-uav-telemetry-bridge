#!/bin/bash
# Build script for Linux ARM64 (Raspberry Pi 4, Jetson Nano, etc.)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/bin"
BINARY_NAME="outb-linux-arm64"

echo "=== Open-UAV-Telemetry-Bridge ARM64 Build ==="
echo "Project root: $PROJECT_ROOT"
echo "Output: $BUILD_DIR/$BINARY_NAME"
echo ""

cd "$PROJECT_ROOT"

# Ensure dependencies are up to date
echo "Downloading dependencies..."
go mod download

# Build for Linux ARM64
echo "Building for Linux ARM64..."
mkdir -p "$BUILD_DIR"
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o "$BUILD_DIR/$BINARY_NAME" ./cmd/outb

# Show build result
echo ""
echo "Build complete!"
ls -lh "$BUILD_DIR/$BINARY_NAME"

echo ""
echo "To deploy to Raspberry Pi:"
echo "  scp $BUILD_DIR/$BINARY_NAME pi@<raspberry-pi-ip>:~/"
echo "  scp configs/config.example.yaml pi@<raspberry-pi-ip>:~/config.yaml"
