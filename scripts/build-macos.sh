#!/bin/bash
# Build script for macOS Cntrl.app bundle

set -e

# Prompt for version if not provided
if [ -z "$1" ]; then
    read -p "Enter version (e.g., 0.0.24-beta): " VERSION
    if [ -z "$VERSION" ]; then
        echo "Error: Version is required."
        exit 1
    fi
else
    VERSION="$1"
fi

ARCH_INPUT="${2:-auto}"
APP_NAME="Cntrl"

# Map architecture input to GOARCH
case "${ARCH_INPUT}" in
    intel|amd64|x86_64)
        GOARCH="amd64"
        ARCH_SUFFIX="Intel"
        ;;
    arm|arm64|apple|silicon)
        GOARCH="arm64"
        ARCH_SUFFIX="AppleSilicon"
        ;;
    auto|"")
        # Auto-detect based on current machine
        if [ "$(uname -m)" = "arm64" ]; then
            GOARCH="arm64"
            ARCH_SUFFIX="AppleSilicon"
        else
            GOARCH="amd64"
            ARCH_SUFFIX="Intel"
        fi
        ;;
    *)
        echo "Error: Unknown architecture '${ARCH_INPUT}'"
        echo "Usage: $0 [VERSION] [intel|arm|auto]"
        exit 1
        ;;
esac

# Use architecture-specific app bundle path
BUNDLE_DIR="macos/${APP_NAME}-${ARCH_SUFFIX}.app"
CONTENTS_DIR="${BUNDLE_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"

echo "Building ${APP_NAME}.app for macOS (${ARCH_SUFFIX} / ${GOARCH})..."

# Create bundle structure
mkdir -p "${MACOS_DIR}"
mkdir -p "${RESOURCES_DIR}"

# Copy Info.plist from template
if [ -f "macos/Cntrl.app/Contents/Info.plist" ]; then
    cp "macos/Cntrl.app/Contents/Info.plist" "${CONTENTS_DIR}/Info.plist"
fi

# Copy existing icon if present
if [ -f "macos/Cntrl.app/Contents/Resources/AppIcon.icns" ]; then
    cp "macos/Cntrl.app/Contents/Resources/AppIcon.icns" "${RESOURCES_DIR}/AppIcon.icns"
fi

# Build the binary
# Note: CGO is required for systray on macOS
echo "Compiling binary for GOARCH=${GOARCH}..."
GOARCH="${GOARCH}" CGO_ENABLED=1 go build -ldflags="-s -w -X main.Version=${VERSION}" -o "${MACOS_DIR}/${APP_NAME}" ./cmd/cntrl

# Update version in Info.plist
if [ -f "${CONTENTS_DIR}/Info.plist" ]; then
    echo "Updating Info.plist with version ${VERSION}..."
    sed -i '' "s/<string>1.0.0<\/string>/<string>${VERSION}<\/string>/" "${CONTENTS_DIR}/Info.plist" 2>/dev/null || true
fi

echo ""
echo "✅ ${APP_NAME}.app built successfully!"
echo "   Location: ${BUNDLE_DIR}"
echo "   Version: ${VERSION}"
echo "   Architecture: ${ARCH_SUFFIX} (${GOARCH})"
echo ""
echo "To test: open ${BUNDLE_DIR}"
