#!/bin/bash
# Create a styled DMG installer for Cntrl

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
BACKGROUND_IMG="macos/dmg-background.png"

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
DMG_TEMP_PATH="dist/${DMG_NAME}_temp.dmg"
VOLUME_NAME="${APP_NAME}"

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

echo "Creating styled DMG installer for ${ARCH_SUFFIX}..."

# Create dist directory
mkdir -p dist

# Create temporary DMG directory
DMG_TEMP=$(mktemp -d)

# Copy app to temp directory (rename to Cntrl.app for cleaner install)
cp -R "${BUNDLE_PATH}" "${DMG_TEMP}/${APP_NAME}.app"

# Create Applications symlink for drag-and-drop install
ln -s /Applications "${DMG_TEMP}/Applications"

# Copy background image
if [ -f "${BACKGROUND_IMG}" ]; then
    mkdir -p "${DMG_TEMP}/.background"
    cp "${BACKGROUND_IMG}" "${DMG_TEMP}/.background/background.png"
    HAS_BACKGROUND=true
else
    echo "⚠️  Warning: Background image not found at ${BACKGROUND_IMG}"
    HAS_BACKGROUND=false
fi

echo "Packaging ${DMG_NAME}.dmg..."

# Remove old DMG if exists
rm -f "${DMG_PATH}" "${DMG_TEMP_PATH}"

# Create a read-write DMG first (needed for customization)
hdiutil create -volname "${VOLUME_NAME}" \
    -srcfolder "${DMG_TEMP}" \
    -ov -format UDRW \
    "${DMG_TEMP_PATH}"

# Mount the DMG
echo "Customizing DMG appearance..."
MOUNT_DIR=$(hdiutil attach -readwrite -noverify "${DMG_TEMP_PATH}" | grep "/Volumes/${VOLUME_NAME}" | tail -1 | awk '{print $3}')

if [ -z "${MOUNT_DIR}" ]; then
    # Try alternate parsing
    MOUNT_DIR="/Volumes/${VOLUME_NAME}"
fi

# Wait for mount
sleep 2

# Apply custom styling via AppleScript
if [ "${HAS_BACKGROUND}" = true ]; then
    osascript <<EOF
tell application "Finder"
    tell disk "${VOLUME_NAME}"
        open
        set current view of container window to icon view
        set toolbar visible of container window to false
        set statusbar visible of container window to false
        set bounds of container window to {100, 100, 700, 500}
        set viewOptions to the icon view options of container window
        set arrangement of viewOptions to not arranged
        set icon size of viewOptions to 80
        set background picture of viewOptions to file ".background:background.png"
        set position of item "${APP_NAME}.app" of container window to {160, 240}
        set position of item "Applications" of container window to {440, 240}
        close
        open
        update without registering applications
        delay 2
        close
    end tell
end tell
EOF
else
    # Fallback: just position icons without background
    osascript <<EOF
tell application "Finder"
    tell disk "${VOLUME_NAME}"
        open
        set current view of container window to icon view
        set toolbar visible of container window to false
        set statusbar visible of container window to false
        set bounds of container window to {100, 100, 700, 500}
        set viewOptions to the icon view options of container window
        set arrangement of viewOptions to not arranged
        set icon size of viewOptions to 80
        set position of item "${APP_NAME}.app" of container window to {160, 240}
        set position of item "Applications" of container window to {440, 240}
        close
        open
        update without registering applications
        delay 2
        close
    end tell
end tell
EOF
fi

# Unmount
sync
hdiutil detach "${MOUNT_DIR}" -quiet || hdiutil detach "${MOUNT_DIR}" -force

# Convert to compressed read-only DMG
hdiutil convert "${DMG_TEMP_PATH}" -format UDZO -o "${DMG_PATH}"

# Cleanup
rm -rf "${DMG_TEMP}" "${DMG_TEMP_PATH}"

echo ""
echo "✅ DMG created successfully!"
echo "   Location: ${DMG_PATH}"
echo "   Size: $(du -h "${DMG_PATH}" | cut -f1)"
