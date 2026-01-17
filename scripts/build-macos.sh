#!/bin/bash
# Build script for macOS Cntrl.app bundle

set -e

VERSION="${1:-dev}"
ARCH="${2:-$(uname -m)}"
APP_NAME="Cntrl"
BUNDLE_DIR="macos/${APP_NAME}.app"
CONTENTS_DIR="${BUNDLE_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"
ICON_SOURCE="winres/app.png"

echo "Building ${APP_NAME}.app for macOS ${ARCH}..."

# Create bundle structure
mkdir -p "${MACOS_DIR}"
mkdir -p "${RESOURCES_DIR}"

# Build the binary
# Note: CGO is required for systray on macOS
echo "Compiling binary..."
GOARCH="${ARCH}" CGO_ENABLED=1 go build -ldflags="-s -w -X main.Version=${VERSION}" -o "${MACOS_DIR}/${APP_NAME}" ./cmd/cntrl

# Update version in Info.plist
echo "Updating Info.plist with version ${VERSION}..."
sed -i '' "s/<string>1.0.0<\/string>/<string>${VERSION}<\/string>/" "${CONTENTS_DIR}/Info.plist" 2>/dev/null || true

# Create icns from the source PNG
if [ -f "${ICON_SOURCE}" ]; then
    echo "Creating app icon from ${ICON_SOURCE}..."
    
    ICONSET_DIR=$(mktemp -d)/AppIcon.iconset
    mkdir -p "${ICONSET_DIR}"
    
    # Create all required icon sizes
    sips -z 16 16     "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_16x16.png" >/dev/null
    sips -z 32 32     "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_16x16@2x.png" >/dev/null
    sips -z 32 32     "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_32x32.png" >/dev/null
    sips -z 64 64     "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_32x32@2x.png" >/dev/null
    sips -z 128 128   "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_128x128.png" >/dev/null
    sips -z 256 256   "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_128x128@2x.png" >/dev/null
    sips -z 256 256   "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_256x256.png" >/dev/null
    sips -z 512 512   "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_256x256@2x.png" >/dev/null
    sips -z 512 512   "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_512x512.png" >/dev/null
    sips -z 1024 1024 "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_512x512@2x.png" >/dev/null
    
    # Generate icns
    iconutil -c icns "${ICONSET_DIR}" -o "${RESOURCES_DIR}/AppIcon.icns"
    rm -rf "$(dirname ${ICONSET_DIR})"
    
    echo "✅ App icon created"
else
    echo "⚠️  Warning: ${ICON_SOURCE} not found, skipping icon creation"
fi

echo ""
echo "✅ ${APP_NAME}.app built successfully!"
echo "   Location: ${BUNDLE_DIR}"
echo "   Version: ${VERSION}"
echo "   Architecture: ${ARCH}"
echo ""
echo "To test: open ${BUNDLE_DIR}"
echo "To install: cp -r ${BUNDLE_DIR} /Applications/"
