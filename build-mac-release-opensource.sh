#!/bin/bash

# CCE Mac Release Build Script - Open Source Edition
# Simplified build script for unsigned open source releases
# Usage: ./build-mac-release-opensource.sh [version]

set -e  # Exit on error

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get version from argument or git tag
VERSION="${1:-}"
if [ -z "$VERSION" ]; then
    VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.0.1")
    echo -e "${YELLOW}No version specified, using: $VERSION${NC}"
fi

# Clean version (remove 'v' prefix if present)
VERSION="${VERSION#v}"
VERSION_TAG="v${VERSION}"

# Project paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROXY_DIR="$SCRIPT_DIR/proxy"
CLIENT_DIR="$SCRIPT_DIR/cce-client"
BUILD_DIR="$SCRIPT_DIR/build"
RELEASE_DIR="$SCRIPT_DIR/releases/$VERSION_TAG"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  CCE Open Source Build ${VERSION_TAG}${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Clean and create directories
echo -e "${BLUE}Preparing build environment...${NC}"
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"
mkdir -p "$RELEASE_DIR"

# Build proxy
build_proxy() {
    local arch=$1
    echo -e "${BLUE}Building proxy for darwin/$arch...${NC}"

    cd "$PROXY_DIR"

    # Build with version info
    CGO_ENABLED=0 GOOS=darwin GOARCH=$arch go build \
        -ldflags="-s -w -X main.VERSION=$VERSION -X main.BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
        -o "$BUILD_DIR/claude-proxy-$arch" \
        ./cmd/main.go

    echo -e "${GREEN}✓ Proxy built for $arch${NC}"
}

# Build client
build_client() {
    local arch=$1
    echo -e "${BLUE}Building client for darwin/$arch...${NC}"

    cd "$CLIENT_DIR"

    # Prepare proxy binary
    mkdir -p resources
    cp "$BUILD_DIR/claude-proxy-$arch" resources/claude-proxy

    # Build client
    CGO_ENABLED=1 GOOS=darwin GOARCH=$arch go build \
        -ldflags="-s -w" \
        -o "$BUILD_DIR/CCE-$arch" \
        ./cmd/main.go 2>&1 | grep -v "ld: warning: ignoring duplicate libraries" || true

    echo -e "${GREEN}✓ Client built for $arch${NC}"
}

# Create app bundle
create_app() {
    local arch=$1
    local app_name="CCE"
    local bundle_path="$BUILD_DIR/${app_name}.app"

    echo -e "${BLUE}Creating .app bundle for $arch...${NC}"

    # Remove old bundle
    rm -rf "$bundle_path"

    # Create app structure
    mkdir -p "$bundle_path/Contents/MacOS"
    mkdir -p "$bundle_path/Contents/Resources"

    # Copy binaries
    cp "$BUILD_DIR/CCE-$arch" "$bundle_path/Contents/MacOS/CCE"
    cp "$BUILD_DIR/claude-proxy-$arch" "$bundle_path/Contents/Resources/claude-proxy"

    # Copy resources
    if [ -f "$CLIENT_DIR/resources/icon.png" ]; then
        cp "$CLIENT_DIR/resources/icon.png" "$bundle_path/Contents/Resources/"
    fi

    # Copy sample config
    if [ -f "$PROXY_DIR/configs/config.example.yaml" ]; then
        cp "$PROXY_DIR/configs/config.example.yaml" "$bundle_path/Contents/Resources/proxy-config.yaml"
    fi

    # Create Info.plist
    cat > "$bundle_path/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>CCE</string>
    <key>CFBundleDisplayName</key>
    <string>Claude Code Exchange</string>
    <key>CFBundleIdentifier</key>
    <string>com.cce.client</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleExecutable</key>
    <string>CCE</string>
    <key>CFBundleIconFile</key>
    <string>icon</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.developer-tools</string>
</dict>
</plist>
EOF

    # Make executable
    chmod +x "$bundle_path/Contents/MacOS/CCE"
    chmod +x "$bundle_path/Contents/Resources/claude-proxy"

    echo -e "${GREEN}✓ App bundle created${NC}"
}

# Create DMG with instructions
create_dmg() {
    local arch=$1
    local dmg_name="CCE-${VERSION_TAG}-${arch}.dmg"
    local dmg_path="$RELEASE_DIR/$dmg_name"

    echo -e "${BLUE}Creating DMG...${NC}"

    # Create temporary directory for DMG contents
    local temp_dir=$(mktemp -d)

    # Copy app
    cp -R "$BUILD_DIR/CCE.app" "$temp_dir/"

    # Create Applications symlink
    ln -s /Applications "$temp_dir/Applications"

    # Create installation instructions
    cat > "$temp_dir/INSTALL.txt" <<'EOF'
==============================================================
                CCE - Claude Code Exchange
==============================================================

INSTALLATION INSTRUCTIONS FOR UNSIGNED APP:

Step 1: Install the Application
--------------------------------
1. Drag "CCE.app" to your "Applications" folder

Step 2: Trust the Application (REQUIRED)
-----------------------------------------
Since this is an open-source unsigned application, macOS will
block it by default. You need to manually trust it:

Method A: System Settings (macOS 13+)
1. Try to open CCE from Applications
2. You'll see: "CCE cannot be opened because the developer
   cannot be verified"
3. Go to System Settings > Privacy & Security
4. Find the message about CCE being blocked
5. Click "Open Anyway"
6. Click "Open" in the confirmation dialog

Method B: Right-Click Method (All macOS versions)
1. In Applications, right-click (or Control-click) on CCE
2. Select "Open" from the menu
3. Click "Open" in the security dialog

Method C: Terminal Command (Technical users)
1. Open Terminal
2. Run: xattr -cr /Applications/CCE.app
3. Now you can open CCE normally

TROUBLESHOOTING:
----------------
If the app still won't open:
1. Make sure you copied it to Applications first
2. Try Method C (Terminal command)
3. Check System Settings > Privacy & Security > Developer Tools
   and ensure Terminal has access if needed

CONFIGURATION:
--------------
After launching, go to Settings to configure your Claude API keys.

SUPPORT:
--------
For help and updates: https://github.com/yourusername/cce

==============================================================
EOF

    # Create DMG
    echo "Building disk image..."
    hdiutil create -volname "CCE $VERSION_TAG" \
        -srcfolder "$temp_dir" \
        -ov -format UDZO \
        "$dmg_path"

    # Clean up
    rm -rf "$temp_dir"

    if [ -f "$dmg_path" ]; then
        echo -e "${GREEN}✓ DMG created: $dmg_name${NC}"
        ls -lh "$dmg_path"
    else
        echo -e "${RED}Error: Failed to create DMG${NC}"
        return 1
    fi
}

# Generate checksums
generate_checksums() {
    echo -e "${BLUE}Generating checksums...${NC}"
    cd "$RELEASE_DIR"
    shasum -a 256 *.dmg > checksums.txt
    echo -e "${GREEN}✓ Checksums generated${NC}"
}

# Generate simple release notes
generate_release_notes() {
    echo -e "${BLUE}Generating release notes...${NC}"

    cat > "$RELEASE_DIR/RELEASE.md" <<EOF
# CCE ${VERSION_TAG} Release

## Installation Instructions

### ⚠️ Important: This is an Unsigned Open Source Application

Since CCE is an open-source project without Apple Developer certification,
macOS will show security warnings on first launch. This is normal and safe.

### Download

Choose the version for your Mac:
- **Apple Silicon (M1/M2/M3)**: CCE-${VERSION_TAG}-arm64.dmg
- **Intel Macs**: CCE-${VERSION_TAG}-amd64.dmg
- **Not sure?**: Download the universal version (larger file size)

### Installation Steps

1. **Download** the appropriate DMG file
2. **Open** the DMG file
3. **Drag** CCE to your Applications folder
4. **Important**: Read the INSTALL.txt file in the DMG for trust instructions

### First Launch (REQUIRED)

**Option 1: Right-click method (Easiest)**
- Right-click on CCE in Applications
- Select "Open" from the menu
- Click "Open" in the security dialog

**Option 2: System Settings**
- Try to open CCE normally
- When blocked, go to System Settings > Privacy & Security
- Click "Open Anyway" next to the CCE message

**Option 3: Terminal (Advanced)**
\`\`\`bash
xattr -cr /Applications/CCE.app
\`\`\`

### Verify Download

Check the SHA256 checksum:
\`\`\`bash
shasum -a 256 CCE-${VERSION_TAG}-*.dmg
\`\`\`
Compare with checksums.txt

### Support

- GitHub: https://github.com/yourusername/cce
- Issues: https://github.com/yourusername/cce/issues

---
Generated on $(date -u '+%Y-%m-%d')
EOF

    echo -e "${GREEN}✓ Release notes generated${NC}"
}

# Main build process
main() {
    # Detect architecture
    ARCH=$(uname -m)
    if [ "$ARCH" = "arm64" ]; then
        BUILD_ARCH="arm64"
    else
        BUILD_ARCH="amd64"
    fi

    echo "Building for architecture: $BUILD_ARCH"
    echo ""

    # Build steps
    build_proxy "$BUILD_ARCH"
    build_client "$BUILD_ARCH"
    create_app "$BUILD_ARCH"
    create_dmg "$BUILD_ARCH"

    # Also build universal if on Apple Silicon
    if [ "$BUILD_ARCH" = "arm64" ] && command -v lipo &> /dev/null; then
        echo ""
        echo -e "${BLUE}Creating universal binary...${NC}"

        # Build amd64 versions
        build_proxy "amd64"
        build_client "amd64"

        # Create universal binaries
        lipo -create "$BUILD_DIR/claude-proxy-arm64" "$BUILD_DIR/claude-proxy-amd64" \
            -output "$BUILD_DIR/claude-proxy-universal"
        lipo -create "$BUILD_DIR/CCE-arm64" "$BUILD_DIR/CCE-amd64" \
            -output "$BUILD_DIR/CCE-universal"

        # Create universal app
        create_app "universal"
        create_dmg "universal"
    fi

    # Generate release files
    generate_checksums
    generate_release_notes

    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}       Build Complete! 🎉${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "Release files in: $RELEASE_DIR"
    echo ""
    ls -lh "$RELEASE_DIR"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo "1. Test the DMG installation on a clean system"
    echo "2. Upload to GitHub Releases"
    echo "3. Share the trust instructions with users"
}

# Run main
main