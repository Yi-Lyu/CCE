#!/bin/bash

# Advanced DMG creation script with custom appearance
# Creates a professional-looking DMG installer with background image and custom layout

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parameters
APP_NAME="${1:-CCE}"
VERSION="${2:-1.0.0}"
APP_PATH="${3}"
OUTPUT_PATH="${4}"

if [ -z "$APP_PATH" ] || [ -z "$OUTPUT_PATH" ]; then
    echo "Usage: $0 [APP_NAME] [VERSION] APP_PATH OUTPUT_PATH"
    echo "Example: $0 CCE 1.0.0 /path/to/CCE.app /path/to/output.dmg"
    exit 1
fi

# Temporary directory for DMG contents
TEMP_DIR=$(mktemp -d)
VOL_NAME="${APP_NAME} ${VERSION}"

echo -e "${BLUE}Creating DMG: ${VOL_NAME}${NC}"

# Copy app to temp directory
echo "Copying application..."
cp -R "$APP_PATH" "$TEMP_DIR/"

# Create Applications symlink
ln -s /Applications "$TEMP_DIR/Applications"

# Create background directory
mkdir "$TEMP_DIR/.background"

# Create a simple background image using ImageMagick (if available) or create placeholder
if command -v convert &> /dev/null; then
    echo "Creating background image..."
    convert -size 540x380 xc:'#f0f0f0' \
        -fill '#333333' -pointsize 24 -font "Helvetica" \
        -draw "text 270,50 '${APP_NAME}'" \
        -fill '#666666' -pointsize 14 \
        -draw "text 270,100 'Drag to Applications folder to install'" \
        -stroke '#cccccc' -strokewidth 2 \
        -draw "roundrectangle 50,140,240,320,10,10" \
        -draw "roundrectangle 300,140,490,320,10,10" \
        -stroke none -fill '#999999' -pointsize 12 \
        -draw "text 145,340 '${APP_NAME}.app'" \
        -draw "text 395,340 'Applications'" \
        "$TEMP_DIR/.background/background.png"
else
    echo -e "${YELLOW}ImageMagick not found. Skipping background image.${NC}"
fi

# Create the initial DMG
echo "Creating initial DMG..."
hdiutil create -srcfolder "$TEMP_DIR" \
    -volname "$VOL_NAME" \
    -fs HFS+ \
    -fsargs "-c c=64,a=16,e=16" \
    -format UDRW \
    -size 200m \
    "$OUTPUT_PATH.temp.dmg"

# Mount the DMG
echo "Mounting DMG for customization..."
MOUNT_DIR="/Volumes/$VOL_NAME"
hdiutil attach "$OUTPUT_PATH.temp.dmg" -readwrite -noverify -noautoopen

# Wait for mount
sleep 2

# Apply custom settings using AppleScript
echo "Applying custom appearance..."
osascript <<EOF
tell application "Finder"
    tell disk "$VOL_NAME"
        open

        -- Window settings
        set current view of container window to icon view
        set toolbar visible of container window to false
        set statusbar visible of container window to false
        set bounds of container window to {400, 100, 940, 480}

        -- Icon view options
        set viewOptions to icon view options of container window
        set arrangement of viewOptions to not arranged
        set icon size of viewOptions to 72
        set background picture of viewOptions to file ".background:background.png"

        -- Icon positions
        set position of item "${APP_NAME}.app" to {145, 200}
        set position of item "Applications" to {395, 200}

        -- Hide background folder
        set position of item ".background" to {999, 999}

        update without registering applications
        delay 2
        close
    end tell
end tell
EOF

# Sync and unmount
echo "Finalizing DMG..."
sync
hdiutil detach "$MOUNT_DIR"

# Convert to compressed read-only DMG
echo "Compressing DMG..."
hdiutil convert "$OUTPUT_PATH.temp.dmg" \
    -format UDZO \
    -imagekey zlib-level=9 \
    -o "$OUTPUT_PATH"

# Clean up
rm -f "$OUTPUT_PATH.temp.dmg"
rm -rf "$TEMP_DIR"

echo -e "${GREEN}✓ DMG created successfully: $OUTPUT_PATH${NC}"

# Show file info
ls -lh "$OUTPUT_PATH"