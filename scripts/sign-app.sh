#!/bin/bash

# Code signing helper script for macOS applications
# Handles signing, verification, and notarization

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
APP_PATH=""
DEVELOPER_ID=""
TEAM_ID=""
APPLE_ID=""
APP_PASSWORD=""
ENTITLEMENTS_PATH=""
NOTARIZE=false
STAPLE=false

# Usage function
usage() {
    echo "Usage: $0 -a APP_PATH -d DEVELOPER_ID [options]"
    echo ""
    echo "Required:"
    echo "  -a, --app PATH            Path to the .app bundle"
    echo "  -d, --developer-id ID     Developer ID for signing (e.g., 'Developer ID Application: Name (TEAMID)')"
    echo ""
    echo "Optional:"
    echo "  -e, --entitlements PATH   Path to entitlements.plist file"
    echo "  -n, --notarize           Submit for notarization"
    echo "  -t, --team-id ID         Team ID for notarization"
    echo "  -i, --apple-id EMAIL     Apple ID for notarization"
    echo "  -p, --app-password PASS  App-specific password for notarization"
    echo "  -s, --staple             Staple notarization ticket to app"
    echo "  -h, --help               Show this help message"
    echo ""
    echo "Examples:"
    echo "  # Sign only"
    echo "  $0 -a MyApp.app -d 'Developer ID Application: John Doe (ABC123)'"
    echo ""
    echo "  # Sign and notarize"
    echo "  $0 -a MyApp.app -d 'Developer ID Application: John Doe (ABC123)' -n -t ABC123 -i john@example.com -p app-password"
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -a|--app)
            APP_PATH="$2"
            shift 2
            ;;
        -d|--developer-id)
            DEVELOPER_ID="$2"
            shift 2
            ;;
        -e|--entitlements)
            ENTITLEMENTS_PATH="$2"
            shift 2
            ;;
        -n|--notarize)
            NOTARIZE=true
            shift
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
        -s|--staple)
            STAPLE=true
            shift
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

# Validate required arguments
if [ -z "$APP_PATH" ] || [ -z "$DEVELOPER_ID" ]; then
    echo -e "${RED}Error: Missing required arguments${NC}"
    usage
fi

if [ ! -d "$APP_PATH" ]; then
    echo -e "${RED}Error: App not found at $APP_PATH${NC}"
    exit 1
fi

# Function to find and list certificates
list_certificates() {
    echo -e "${BLUE}Available Developer ID certificates:${NC}"
    security find-identity -v -p codesigning | grep "Developer ID"
}

# Function to verify developer ID
verify_developer_id() {
    if ! security find-identity -v -p codesigning | grep -q "$DEVELOPER_ID"; then
        echo -e "${YELLOW}Warning: Developer ID not found in keychain${NC}"
        echo ""
        list_certificates
        echo ""
        read -p "Continue anyway? (y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Function to create default entitlements
create_default_entitlements() {
    local temp_entitlements=$(mktemp)
    cat > "$temp_entitlements" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <!-- Allow app to be run from anywhere -->
    <key>com.apple.security.app-sandbox</key>
    <false/>

    <!-- Network access -->
    <key>com.apple.security.network.client</key>
    <true/>
    <key>com.apple.security.network.server</key>
    <true/>

    <!-- File access -->
    <key>com.apple.security.files.user-selected.read-write</key>
    <true/>

    <!-- Required for notarization -->
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>

    <!-- Hardened runtime -->
    <key>com.apple.security.runtime</key>
    <true/>
</dict>
</plist>
EOF
    echo "$temp_entitlements"
}

# Function to sign the app
sign_app() {
    echo -e "${BLUE}Starting code signing process...${NC}"
    echo "App: $APP_PATH"
    echo "Developer ID: $DEVELOPER_ID"

    # Verify developer ID
    verify_developer_id

    # Use provided entitlements or create default
    if [ -z "$ENTITLEMENTS_PATH" ]; then
        echo -e "${YELLOW}No entitlements file provided, using defaults${NC}"
        ENTITLEMENTS_PATH=$(create_default_entitlements)
    fi

    # Remove old signatures
    echo "Removing old signatures..."
    find "$APP_PATH" -type f -name "*.dylib" -exec codesign --remove-signature {} \; 2>/dev/null || true
    find "$APP_PATH" -type f -name "*.so" -exec codesign --remove-signature {} \; 2>/dev/null || true
    codesign --remove-signature "$APP_PATH" 2>/dev/null || true

    # Sign embedded binaries and frameworks first
    echo "Signing embedded components..."

    # Sign frameworks
    if [ -d "$APP_PATH/Contents/Frameworks" ]; then
        find "$APP_PATH/Contents/Frameworks" -name "*.framework" -type d | while read -r framework; do
            echo "  Signing framework: $(basename "$framework")"
            codesign --force --sign "$DEVELOPER_ID" --timestamp --options runtime "$framework"
        done
    fi

    # Sign dylibs
    find "$APP_PATH" -name "*.dylib" -type f | while read -r dylib; do
        echo "  Signing dylib: $(basename "$dylib")"
        codesign --force --sign "$DEVELOPER_ID" --timestamp --options runtime "$dylib"
    done

    # Sign executables in MacOS directory
    if [ -d "$APP_PATH/Contents/MacOS" ]; then
        find "$APP_PATH/Contents/MacOS" -type f -perm +111 | while read -r exe; do
            if [ "$(basename "$exe")" != "$(basename "$APP_PATH" .app)" ]; then
                echo "  Signing executable: $(basename "$exe")"
                codesign --force --sign "$DEVELOPER_ID" --timestamp --options runtime "$exe"
            fi
        done
    fi

    # Sign executables in Resources directory
    if [ -d "$APP_PATH/Contents/Resources" ]; then
        find "$APP_PATH/Contents/Resources" -type f -perm +111 | while read -r exe; do
            echo "  Signing resource executable: $(basename "$exe")"
            codesign --force --sign "$DEVELOPER_ID" --timestamp --options runtime "$exe"
        done
    fi

    # Finally, sign the main app bundle
    echo "Signing main app bundle..."
    codesign --force --deep --sign "$DEVELOPER_ID" \
        --entitlements "$ENTITLEMENTS_PATH" \
        --timestamp \
        --options runtime \
        "$APP_PATH"

    echo -e "${GREEN}✓ App signed successfully${NC}"

    # Verify signature
    echo ""
    echo "Verifying signature..."
    codesign --verify --deep --strict --verbose=2 "$APP_PATH"

    echo ""
    echo "Checking signature assessment..."
    spctl --assess --verbose "$APP_PATH" 2>&1 || true

    # Clean up temporary entitlements if created
    if [[ "$ENTITLEMENTS_PATH" == /tmp/* ]]; then
        rm "$ENTITLEMENTS_PATH"
    fi
}

# Function to notarize the app
notarize_app() {
    if [ "$NOTARIZE" != true ]; then
        return
    fi

    if [ -z "$TEAM_ID" ] || [ -z "$APPLE_ID" ] || [ -z "$APP_PASSWORD" ]; then
        echo -e "${YELLOW}Skipping notarization: Missing credentials${NC}"
        echo "Required: --team-id, --apple-id, --app-password"
        return
    fi

    echo -e "${BLUE}Starting notarization process...${NC}"

    # Create a zip for notarization
    local zip_path="${APP_PATH%.app}.zip"
    echo "Creating zip archive..."
    ditto -c -k --keepParent "$APP_PATH" "$zip_path"

    # Submit for notarization
    echo "Submitting to Apple for notarization..."
    echo "(This may take several minutes...)"

    xcrun notarytool submit "$zip_path" \
        --apple-id "$APPLE_ID" \
        --password "$APP_PASSWORD" \
        --team-id "$TEAM_ID" \
        --wait \
        --progress

    # Check notarization status
    echo ""
    echo "Checking notarization status..."
    xcrun notarytool info "$zip_path" \
        --apple-id "$APPLE_ID" \
        --password "$APP_PASSWORD" \
        --team-id "$TEAM_ID"

    # Clean up zip
    rm "$zip_path"

    echo -e "${GREEN}✓ App notarized successfully${NC}"

    # Staple if requested
    if [ "$STAPLE" = true ]; then
        staple_app
    fi
}

# Function to staple notarization ticket
staple_app() {
    echo -e "${BLUE}Stapling notarization ticket...${NC}"

    xcrun stapler staple "$APP_PATH"

    # Verify stapling
    echo "Verifying stapled ticket..."
    xcrun stapler validate "$APP_PATH"

    echo -e "${GREEN}✓ Notarization ticket stapled${NC}"
}

# Function to create a signed DMG
sign_dmg() {
    local dmg_path="$1"

    if [ -f "$dmg_path" ]; then
        echo -e "${BLUE}Signing DMG: $dmg_path${NC}"
        codesign --force --sign "$DEVELOPER_ID" --timestamp "$dmg_path"
        echo -e "${GREEN}✓ DMG signed${NC}"

        if [ "$NOTARIZE" = true ]; then
            echo "Notarizing DMG..."
            xcrun notarytool submit "$dmg_path" \
                --apple-id "$APPLE_ID" \
                --password "$APP_PASSWORD" \
                --team-id "$TEAM_ID" \
                --wait

            echo "Stapling DMG..."
            xcrun stapler staple "$dmg_path"
            echo -e "${GREEN}✓ DMG notarized and stapled${NC}"
        fi
    fi
}

# Main execution
main() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}   macOS App Code Signing Tool${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""

    # Sign the app
    sign_app

    # Notarize if requested
    if [ "$NOTARIZE" = true ]; then
        notarize_app
    fi

    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}   Code signing complete!${NC}"
    echo -e "${GREEN}========================================${NC}"

    # Display final status
    echo ""
    echo "App Information:"
    codesign -dv "$APP_PATH" 2>&1 | grep -E "Identifier=|Authority="

    if [ "$NOTARIZE" = true ]; then
        echo ""
        echo "Notarization Status: ✓ Submitted and processed"
        if [ "$STAPLE" = true ]; then
            echo "Stapling Status: ✓ Ticket attached"
        fi
    fi
}

# Run main function
main