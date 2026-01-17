#!/bin/bash
# Create a DMG installer for Cntrl

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

ARCH="${2:-universal}"
APP_NAME="Cntrl"

# Map architecture to proper names
case "${ARCH}" in
    intel|amd64)
        ARCH_SUFFIX="Intel"
        BUNDLE_PATH="macos/${APP_NAME}-Intel.app"
        ;;
    arm|arm64|apple|silicon)
        ARCH_SUFFIX="AppleSilicon"
        BUNDLE_PATH="macos/${APP_NAME}-AppleSilicon.app"
        ;;
    universal)
        ARCH_SUFFIX="Universal"
        BUNDLE_PATH="macos/${APP_NAME}.app"
        ;;
    *)
        echo "Error: Unknown architecture '${ARCH}'"
        echo "Usage: $0 [VERSION] [intel|arm|universal]"
        exit 1
        ;;
esac

DMG_NAME="${APP_NAME}_${VERSION}_macOS_${ARCH_SUFFIX}"
DMG_PATH="dist/${DMG_NAME}.dmg"
VOLUME_NAME="${APP_NAME} ${VERSION} (${ARCH_SUFFIX})"

# Check if app bundle exists
if [ ! -d "${BUNDLE_PATH}" ]; then
    echo "Error: ${BUNDLE_PATH} not found."
    echo ""
    echo "Build first with:"
    if [ "${ARCH}" = "universal" ]; then
        echo "  ./scripts/build-macos-universal.sh ${VERSION}"
    else
        echo "  ./scripts/build-macos.sh ${VERSION} ${ARCH}"
    fi
    exit 1
fi

echo "Creating DMG installer for ${ARCH_SUFFIX}..."

# Create dist directory
mkdir -p dist

# Create temporary DMG directory
DMG_TEMP=$(mktemp -d)
mkdir -p "${DMG_TEMP}/.background"

# Copy app to temp directory (rename to Cntrl.app for cleaner install)
cp -R "${BUNDLE_PATH}" "${DMG_TEMP}/${APP_NAME}.app"

# Create Applications symlink for drag-and-drop install
ln -s /Applications "${DMG_TEMP}/Applications"

# Create the DMG
echo "Packaging ${DMG_NAME}.dmg..."

# Remove old DMG if exists
rm -f "${DMG_PATH}"

# Create DMG using hdiutil
hdiutil create -volname "${VOLUME_NAME}" \
    -srcfolder "${DMG_TEMP}" \
    -ov -format UDZO \
    "${DMG_PATH}"

# Cleanup
rm -rf "${DMG_TEMP}"

echo ""
echo "✅ DMG created successfully!"
echo "   Location: ${DMG_PATH}"
echo "   Size: $(du -h "${DMG_PATH}" | cut -f1)"
