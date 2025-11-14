#!/bin/bash

# CCE Mac Release Build Script
# This script automates the build and packaging process for macOS distribution
# Usage: ./build-mac-release.sh [options]

set -e  # Exit on error

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
VERSION=""
BUILD_ARCH="universal"  # Options: arm64, amd64, universal
SIGN_APP=false
NOTARIZE_APP=false
OUTPUT_DIR="releases"
DEVELOPER_ID=""
TEAM_ID=""
APPLE_ID=""
APP_PASSWORD=""

# Project paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROXY_DIR="$SCRIPT_DIR/proxy"
CLIENT_DIR="$SCRIPT_DIR/cce-client"
BUILD_DIR="$SCRIPT_DIR/build"
TEMP_DIR="$SCRIPT_DIR/temp"

# Print usage
usage() {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  -v, --version VERSION      Specify version (e.g., 1.0.0)"
    echo "  -a, --arch ARCH           Build architecture (arm64/amd64/universal) [default: universal]"
    echo "  -s, --sign                Enable code signing"
    echo "  -n, --notarize            Submit for Apple notarization (requires --sign)"
    echo "  -d, --developer-id ID     Developer ID for code signing"
    echo "  -t, --team-id ID          Team ID for notarization"
    echo "  -i, --apple-id EMAIL      Apple ID for notarization"
    echo "  -p, --app-password PASS   App-specific password for notarization"
    echo "  -o, --output DIR          Output directory [default: releases]"
    echo "  -h, --help                Show this help message"
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -a|--arch)
            BUILD_ARCH="$2"
            shift 2
            ;;
        -s|--sign)
            SIGN_APP=true
            shift
            ;;
        -n|--notarize)
            NOTARIZE_APP=true
            SIGN_APP=true
            shift
            ;;
        -d|--developer-id)
            DEVELOPER_ID="$2"
            shift 2
            ;;
        -t|--team-id)
            TEAM_ID="$2"
            shift 2
            ;;
        -i|--apple-id)
            APPLE_ID="$2"
            shift 2
            ;;
        -p|--app-password)
            APP_PASSWORD="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Get version from git tag if not specified
if [ -z "$VERSION" ]; then
    VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.0.1")
    echo -e "${YELLOW}No version specified, using git tag: $VERSION${NC}"
else
    echo -e "${BLUE}Building version: $VERSION${NC}"
fi

# Clean version (remove 'v' prefix if present)
VERSION="${VERSION#v}"
VERSION_TAG="v${VERSION}"

# Create output directory
RELEASE_DIR="$OUTPUT_DIR/$VERSION_TAG"
mkdir -p "$RELEASE_DIR"
mkdir -p "$BUILD_DIR"
mkdir -p "$TEMP_DIR"

echo -e "${GREEN}Starting CCE Mac Release Build${NC}"
echo "Version: $VERSION"
echo "Architecture: $BUILD_ARCH"
echo "Output: $RELEASE_DIR"
echo "----------------------------------------"

# Function to build proxy for specific architecture
build_proxy() {
    local arch=$1
    local os="darwin"

    echo -e "${BLUE}Building proxy for $os/$arch...${NC}"

    cd "$PROXY_DIR"

    # Build with version info
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build \
        -ldflags="-s -w -X main.VERSION=$VERSION -X main.BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
        -o "$BUILD_DIR/claude-proxy-$arch" \
        ./cmd/main.go

    echo -e "${GREEN}✓ Proxy built for $arch${NC}"
}

# Function to build client for specific architecture
build_client() {
    local arch=$1

    echo -e "${BLUE}Building client for darwin/$arch...${NC}"

    cd "$CLIENT_DIR"

    # Prepare proxy binary
    echo "Copying proxy binary to client resources..."
    mkdir -p resources
    cp "$BUILD_DIR/claude-proxy-$arch" resources/claude-proxy

    # Build client
    CGO_ENABLED=1 GOOS=darwin GOARCH=$arch go build \
        -ldflags="-s -w" \
        -o "$BUILD_DIR/CCE-$arch" \
        ./cmd/main.go

    echo -e "${GREEN}✓ Client built for $arch${NC}"
}

# Function to create universal binary
create_universal_binary() {
    echo -e "${BLUE}Creating universal binary...${NC}"

    # Create universal proxy
    lipo -create \
        "$BUILD_DIR/claude-proxy-arm64" \
        "$BUILD_DIR/claude-proxy-amd64" \
        -output "$BUILD_DIR/claude-proxy-universal"

    # Create universal client
    lipo -create \
        "$BUILD_DIR/CCE-arm64" \
        "$BUILD_DIR/CCE-amd64" \
        -output "$BUILD_DIR/CCE-universal"

    echo -e "${GREEN}✓ Universal binaries created${NC}"
}

# Function to create app bundle
create_app_bundle() {
    local arch=$1
    local app_name="CCE"
    local bundle_path="$BUILD_DIR/${app_name}-${arch}.app"

    echo -e "${BLUE}Creating .app bundle for $arch...${NC}"

    # Create app structure
    rm -rf "$bundle_path"
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

    echo -e "${GREEN}✓ App bundle created for $arch${NC}"
}

# Function to sign app
sign_app() {
    local app_path=$1

    if [ "$SIGN_APP" = true ] && [ -n "$DEVELOPER_ID" ]; then
        echo -e "${BLUE}Signing app...${NC}"

        # Sign all binaries and frameworks
        find "$app_path" -type f -perm +111 -exec codesign --force --sign "$DEVELOPER_ID" --timestamp --options runtime {} \;

        # Sign the app bundle
        codesign --force --sign "$DEVELOPER_ID" --timestamp --options runtime --entitlements "$SCRIPT_DIR/entitlements.plist" "$app_path"

        # Verify signature
        codesign --verify --verbose "$app_path"

        echo -e "${GREEN}✓ App signed${NC}"
    fi
}

# Function to create DMG
create_dmg() {
    local arch=$1
    local app_name="CCE"
    local dmg_name="${app_name}-${VERSION_TAG}-${arch}.dmg"
    local dmg_path="$RELEASE_DIR/$dmg_name"
    local app_path="$BUILD_DIR/${app_name}-${arch}.app"

    echo -e "${BLUE}Creating DMG for $arch...${NC}"

    # Verify app exists
    if [ ! -d "$app_path" ]; then
        echo -e "${RED}Error: App bundle not found at $app_path${NC}"
        return 1
    fi

    # Create temporary DMG directory
    local dmg_temp="$TEMP_DIR/dmg-${arch}"
    rm -rf "$dmg_temp"
    mkdir -p "$dmg_temp"

    # Copy app to DMG directory
    cp -R "$app_path" "$dmg_temp/"

    # Create Applications symlink
    ln -s /Applications "$dmg_temp/Applications"

    # Create README with trust instructions
    cat > "$dmg_temp/README.txt" <<EOF
Claude Code Exchange (CCE) ${VERSION_TAG}

IMPORTANT - First Launch Instructions:
=======================================
This is an unsigned open-source application. On first launch:

macOS Ventura (13.0) and later:
1. Drag CCE.app to your Applications folder
2. Try to open CCE normally - you'll see a security warning
3. Go to System Settings > Privacy & Security
4. Scroll down to find "CCE was blocked from use because it is not from an identified developer"
5. Click "Open Anyway"
6. Click "Open" in the dialog that appears

macOS Monterey (12.0) and earlier:
1. Drag CCE.app to your Applications folder
2. Right-click (or Control-click) on CCE.app
3. Select "Open" from the context menu
4. Click "Open" in the security dialog

Alternative method (Terminal):
1. Open Terminal
2. Run: xattr -cr /Applications/CCE.app
3. Open CCE normally

Configuration:
==============
After launching, configure your Claude API keys in the app settings.

For more information and support:
https://github.com/yourusername/cce

© $(date +%Y) CCE Project - Open Source Software
EOF

    # Ensure output directory exists
    mkdir -p "$RELEASE_DIR"

    # Create DMG with error checking
    echo "Creating disk image..."
    if ! hdiutil create -volname "CCE ${VERSION_TAG}" \
        -srcfolder "$dmg_temp" \
        -ov \
        -format UDZO \
        "$dmg_path" 2>&1; then
        echo -e "${RED}Failed to create DMG. Trying alternative method...${NC}"

        # Alternative method using ditto
        local temp_dmg="$dmg_path.tmp"
        hdiutil create -size 100m -fs HFS+ -volname "CCE ${VERSION_TAG}" "$temp_dmg"
        hdiutil attach "$temp_dmg" -mountpoint "/Volumes/CCE-${VERSION_TAG}-temp"
        cp -R "$dmg_temp"/* "/Volumes/CCE-${VERSION_TAG}-temp/"
        hdiutil detach "/Volumes/CCE-${VERSION_TAG}-temp"
        hdiutil convert "$temp_dmg" -format UDZO -o "$dmg_path"
        rm -f "$temp_dmg"
    fi

    # Clean up temp directory
    rm -rf "$dmg_temp"

    # Verify DMG was created
    if [ -f "$dmg_path" ]; then
        echo -e "${GREEN}✓ DMG created: $dmg_name${NC}"
        ls -lh "$dmg_path"
    else
        echo -e "${RED}Error: Failed to create DMG${NC}"
        return 1
    fi
}

# Function to notarize app
notarize_app() {
    local dmg_path=$1

    if [ "$NOTARIZE_APP" = true ] && [ -n "$APPLE_ID" ] && [ -n "$APP_PASSWORD" ] && [ -n "$TEAM_ID" ]; then
        echo -e "${BLUE}Submitting for notarization...${NC}"

        # Submit for notarization
        xcrun notarytool submit "$dmg_path" \
            --apple-id "$APPLE_ID" \
            --password "$APP_PASSWORD" \
            --team-id "$TEAM_ID" \
            --wait

        # Staple the notarization
        xcrun stapler staple "$dmg_path"

        echo -e "${GREEN}✓ App notarized${NC}"
    fi
}

# Function to generate checksums
generate_checksums() {
    echo -e "${BLUE}Generating checksums...${NC}"

    cd "$RELEASE_DIR"
    shasum -a 256 *.dmg > checksums.txt

    echo -e "${GREEN}✓ Checksums generated${NC}"
}

# Function to generate release notes
generate_release_notes() {
    echo -e "${BLUE}Generating release notes...${NC}"

    local release_notes="$RELEASE_DIR/release-notes.md"

    cat > "$release_notes" <<EOF
# CCE ${VERSION_TAG} Release Notes

## What's New

### Features
- [Add new features here]

### Improvements
- [Add improvements here]

### Bug Fixes
- [Add bug fixes here]

## Installation

### macOS

1. Download the appropriate DMG for your system:
   - **Apple Silicon (M1/M2/M3)**: CCE-${VERSION_TAG}-arm64.dmg
   - **Intel**: CCE-${VERSION_TAG}-amd64.dmg
   - **Universal (works on both)**: CCE-${VERSION_TAG}-universal.dmg

2. Open the DMG and drag CCE to your Applications folder

3. On first launch, you may need to right-click and select "Open" due to macOS security settings

## System Requirements

- macOS 10.15 (Catalina) or later
- 100MB free disk space

## Checksums

See \`checksums.txt\` for SHA256 verification.

## Contributors

[Generated from git log]
$(git log --pretty=format:"- %an" $(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")..HEAD 2>/dev/null | sort -u)

---
Generated on $(date -u '+%Y-%m-%d %H:%M:%S') UTC
EOF

    echo -e "${GREEN}✓ Release notes generated${NC}"
}

# Main build process
main() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}   CCE Mac Release Build Starting${NC}"
    echo -e "${GREEN}========================================${NC}"

    # Clean previous builds
    echo -e "${YELLOW}Cleaning previous builds...${NC}"
    rm -rf "$BUILD_DIR"/*
    rm -rf "$TEMP_DIR"/*

    # Build based on architecture choice
    case $BUILD_ARCH in
        arm64)
            build_proxy "arm64"
            build_client "arm64"
            create_app_bundle "arm64"
            sign_app "$BUILD_DIR/CCE-arm64.app"
            create_dmg "arm64"
            notarize_app "$RELEASE_DIR/CCE-${VERSION_TAG}-arm64.dmg"
            ;;
        amd64)
            build_proxy "amd64"
            build_client "amd64"
            create_app_bundle "amd64"
            sign_app "$BUILD_DIR/CCE-amd64.app"
            create_dmg "amd64"
            notarize_app "$RELEASE_DIR/CCE-${VERSION_TAG}-amd64.dmg"
            ;;
        universal)
            # Build for both architectures
            build_proxy "arm64"
            build_proxy "amd64"
            build_client "arm64"
            build_client "amd64"

            # Create universal binary
            create_universal_binary

            # Create app bundles
            create_app_bundle "arm64"
            create_app_bundle "amd64"
            create_app_bundle "universal"

            # Sign apps
            sign_app "$BUILD_DIR/CCE-arm64.app"
            sign_app "$BUILD_DIR/CCE-amd64.app"
            sign_app "$BUILD_DIR/CCE-universal.app"

            # Create DMGs
            create_dmg "arm64"
            create_dmg "amd64"
            create_dmg "universal"

            # Notarize if enabled
            notarize_app "$RELEASE_DIR/CCE-${VERSION_TAG}-arm64.dmg"
            notarize_app "$RELEASE_DIR/CCE-${VERSION_TAG}-amd64.dmg"
            notarize_app "$RELEASE_DIR/CCE-${VERSION_TAG}-universal.dmg"
            ;;
        *)
            echo -e "${RED}Invalid architecture: $BUILD_ARCH${NC}"
            exit 1
            ;;
    esac

    # Generate checksums and release notes
    generate_checksums
    generate_release_notes

    # Clean up temp directory
    rm -rf "$TEMP_DIR"

    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}   Build Complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo -e "${BLUE}Release artifacts available at:${NC}"
    echo "$RELEASE_DIR"
    echo ""
    ls -lh "$RELEASE_DIR"
}

# Run main function
main