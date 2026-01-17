#!/bin/bash
# Create a DMG installer for Cntrl

set -e

VERSION="${1:-dev}"
APP_NAME="Cntrl"
BUNDLE_PATH="macos/${APP_NAME}.app"
DMG_NAME="${APP_NAME}_${VERSION}_macOS"
DMG_PATH="dist/${DMG_NAME}.dmg"
VOLUME_NAME="${APP_NAME} ${VERSION}"

# Check if app bundle exists
if [ ! -d "${BUNDLE_PATH}" ]; then
    echo "Error: ${BUNDLE_PATH} not found. Run scripts/build-macos.sh first."
    exit 1
fi

echo "Creating DMG installer..."

# Create dist directory
mkdir -p dist

# Create temporary DMG directory
DMG_TEMP=$(mktemp -d)
mkdir -p "${DMG_TEMP}/.background"

# Copy app to temp directory
cp -R "${BUNDLE_PATH}" "${DMG_TEMP}/"

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
