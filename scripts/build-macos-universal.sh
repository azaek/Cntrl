#!/bin/bash
# Build universal (fat) binary for macOS that runs on both Intel and Apple Silicon

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

APP_NAME="Cntrl"

# Paths
BUNDLE_DIR="macos/${APP_NAME}.app"
CONTENTS_DIR="${BUNDLE_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"

INTEL_APP="macos/${APP_NAME}-Intel.app"
ARM_APP="macos/${APP_NAME}-AppleSilicon.app"

echo "Building Universal ${APP_NAME}.app for macOS..."
echo ""

# Build Intel binary
echo "=== Building Intel (amd64) binary ==="
./scripts/build-macos.sh "${VERSION}" intel

# Build Apple Silicon binary
echo ""
echo "=== Building Apple Silicon (arm64) binary ==="
./scripts/build-macos.sh "${VERSION}" arm

# Check both builds exist
if [ ! -f "${INTEL_APP}/Contents/MacOS/${APP_NAME}" ]; then
    echo "Error: Intel binary not found at ${INTEL_APP}/Contents/MacOS/${APP_NAME}"
    exit 1
fi

if [ ! -f "${ARM_APP}/Contents/MacOS/${APP_NAME}" ]; then
    echo "Error: Apple Silicon binary not found at ${ARM_APP}/Contents/MacOS/${APP_NAME}"
    exit 1
fi

# Create universal app bundle
echo ""
echo "=== Creating Universal binary ==="
mkdir -p "${MACOS_DIR}"
mkdir -p "${RESOURCES_DIR}"

# Copy Info.plist and resources from one of the builds
cp "${INTEL_APP}/Contents/Info.plist" "${CONTENTS_DIR}/Info.plist" 2>/dev/null || true
cp "${INTEL_APP}/Contents/Resources/AppIcon.icns" "${RESOURCES_DIR}/AppIcon.icns" 2>/dev/null || true

# Create universal binary using lipo
echo "Combining binaries with lipo..."
lipo -create \
    "${INTEL_APP}/Contents/MacOS/${APP_NAME}" \
    "${ARM_APP}/Contents/MacOS/${APP_NAME}" \
    -output "${MACOS_DIR}/${APP_NAME}"

# Verify the universal binary
echo ""
echo "Verifying universal binary:"
lipo -info "${MACOS_DIR}/${APP_NAME}"

echo ""
echo "✅ Universal ${APP_NAME}.app built successfully!"
echo "   Location: ${BUNDLE_DIR}"
echo "   Version: ${VERSION}"
echo ""
echo "App bundles created:"
echo "   - ${INTEL_APP} (Intel only)"
echo "   - ${ARM_APP} (Apple Silicon only)"
echo "   - ${BUNDLE_DIR} (Universal)"
echo ""
echo "To create DMGs:"
echo "   ./scripts/create-dmg.sh ${VERSION} intel"
echo "   ./scripts/create-dmg.sh ${VERSION} arm"
echo "   ./scripts/create-dmg.sh ${VERSION} universal"
