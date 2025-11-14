#!/bin/bash

# Release notes generator script
# Generates comprehensive release notes from git commits and repository information

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default values
VERSION=""
OUTPUT_FILE=""
PREVIOUS_TAG=""
FORMAT="markdown"  # markdown or plain
INCLUDE_STATS=true
INCLUDE_CONTRIBUTORS=true
INCLUDE_COMMIT_LINKS=true
REPO_URL=""

# Usage function
usage() {
    echo "Usage: $0 -v VERSION [options]"
    echo ""
    echo "Options:"
    echo "  -v, --version VERSION      Version to generate notes for (required)"
    echo "  -o, --output FILE         Output file path (default: release-notes-VERSION.md)"
    echo "  -p, --previous TAG        Previous version tag (default: auto-detect)"
    echo "  -f, --format FORMAT       Output format: markdown or plain (default: markdown)"
    echo "  -r, --repo URL           Repository URL for commit links"
    echo "  --no-stats               Exclude statistics"
    echo "  --no-contributors        Exclude contributors list"
    echo "  --no-links              Exclude commit links"
    echo "  -h, --help              Show this help message"
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -p|--previous)
            PREVIOUS_TAG="$2"
            shift 2
            ;;
        -f|--format)
            FORMAT="$2"
            shift 2
            ;;
        -r|--repo)
            REPO_URL="$2"
            shift 2
            ;;
        --no-stats)
            INCLUDE_STATS=false
            shift
            ;;
        --no-contributors)
            INCLUDE_CONTRIBUTORS=false
            shift
            ;;
        --no-links)
            INCLUDE_COMMIT_LINKS=false
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
if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Version is required${NC}"
    usage
fi

# Clean version format
VERSION="${VERSION#v}"
VERSION_TAG="v${VERSION}"

# Set default output file if not specified
if [ -z "$OUTPUT_FILE" ]; then
    OUTPUT_FILE="release-notes-${VERSION}.md"
fi

# Auto-detect repository URL if not provided
if [ -z "$REPO_URL" ] && [ "$INCLUDE_COMMIT_LINKS" = true ]; then
    REPO_URL=$(git remote get-url origin 2>/dev/null | sed 's/\.git$//' | sed 's/git@github\.com:/https:\/\/github.com\//' || echo "")
fi

# Auto-detect previous tag if not provided
if [ -z "$PREVIOUS_TAG" ]; then
    PREVIOUS_TAG=$(git describe --tags --abbrev=0 "${VERSION_TAG}^" 2>/dev/null || echo "")
    if [ -z "$PREVIOUS_TAG" ]; then
        echo -e "${YELLOW}Warning: Could not auto-detect previous tag. Using initial commit.${NC}"
        PREVIOUS_TAG=$(git rev-list --max-parents=0 HEAD)
    fi
fi

echo -e "${BLUE}Generating release notes for ${VERSION_TAG}${NC}"
echo "Previous version: ${PREVIOUS_TAG}"
echo "Output file: ${OUTPUT_FILE}"
echo ""

# Function to categorize commits
categorize_commit() {
    local message="$1"
    local type=""

    # Standard conventional commit types
    if [[ "$message" =~ ^feat(\(.*\))?:  ]]; then
        type="feature"
    elif [[ "$message" =~ ^fix(\(.*\))?:  ]]; then
        type="fix"
    elif [[ "$message" =~ ^docs(\(.*\))?:  ]]; then
        type="docs"
    elif [[ "$message" =~ ^style(\(.*\))?:  ]]; then
        type="style"
    elif [[ "$message" =~ ^refactor(\(.*\))?:  ]]; then
        type="refactor"
    elif [[ "$message" =~ ^perf(\(.*\))?:  ]]; then
        type="performance"
    elif [[ "$message" =~ ^test(\(.*\))?:  ]]; then
        type="test"
    elif [[ "$message" =~ ^build(\(.*\))?:  ]]; then
        type="build"
    elif [[ "$message" =~ ^ci(\(.*\))?:  ]]; then
        type="ci"
    elif [[ "$message" =~ ^chore(\(.*\))?:  ]]; then
        type="chore"
    elif [[ "$message" =~ ^revert(\(.*\))?:  ]]; then
        type="revert"
    # Alternative patterns
    elif [[ "$message" =~ ^[Aa]dd  ]]; then
        type="feature"
    elif [[ "$message" =~ ^[Ff]ix  ]]; then
        type="fix"
    elif [[ "$message" =~ ^[Uu]pdate  ]]; then
        type="improvement"
    elif [[ "$message" =~ ^[Rr]emove  ]] || [[ "$message" =~ ^[Dd]elete  ]]; then
        type="removal"
    elif [[ "$message" =~ ^[Ii]mprove  ]] || [[ "$message" =~ ^[Ee]nhance  ]]; then
        type="improvement"
    elif [[ "$message" =~ ^[Dd]oc  ]]; then
        type="docs"
    elif [[ "$message" =~ ^[Rr]efactor  ]]; then
        type="refactor"
    elif [[ "$message" =~ BREAKING[\- ]CHANGE ]]; then
        type="breaking"
    else
        type="other"
    fi

    echo "$type"
}

# Function to format commit message
format_commit_message() {
    local hash="$1"
    local message="$2"
    local formatted=""

    # Remove conventional commit prefixes
    message=$(echo "$message" | sed -E 's/^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.*\))?:\s*//')

    if [ "$INCLUDE_COMMIT_LINKS" = true ] && [ -n "$REPO_URL" ]; then
        formatted="- ${message} ([${hash:0:7}](${REPO_URL}/commit/${hash}))"
    else
        formatted="- ${message} (${hash:0:7})"
    fi

    echo "$formatted"
}

# Initialize category arrays
declare -a features
declare -a fixes
declare -a improvements
declare -a breaking
declare -a docs
declare -a refactors
declare -a performance
declare -a removals
declare -a other

# Get commit range
if [ -n "$PREVIOUS_TAG" ]; then
    COMMIT_RANGE="${PREVIOUS_TAG}..${VERSION_TAG}"
else
    COMMIT_RANGE="${VERSION_TAG}"
fi

# Process commits
echo "Processing commits..."
while IFS= read -r line; do
    hash=$(echo "$line" | cut -d'|' -f1)
    message=$(echo "$line" | cut -d'|' -f2-)

    category=$(categorize_commit "$message")
    formatted=$(format_commit_message "$hash" "$message")

    case "$category" in
        feature)
            features+=("$formatted")
            ;;
        fix)
            fixes+=("$formatted")
            ;;
        improvement)
            improvements+=("$formatted")
            ;;
        breaking)
            breaking+=("$formatted")
            ;;
        docs)
            docs+=("$formatted")
            ;;
        refactor)
            refactors+=("$formatted")
            ;;
        performance)
            performance+=("$formatted")
            ;;
        removal)
            removals+=("$formatted")
            ;;
        *)
            other+=("$formatted")
            ;;
    esac
done < <(git log --format="%H|%s" $COMMIT_RANGE 2>/dev/null || echo "")

# Get statistics
if [ "$INCLUDE_STATS" = true ]; then
    TOTAL_COMMITS=$(git rev-list --count $COMMIT_RANGE 2>/dev/null || echo "0")
    FILES_CHANGED=$(git diff --stat $COMMIT_RANGE 2>/dev/null | tail -1 | awk '{print $1}' || echo "0")
    INSERTIONS=$(git diff --stat $COMMIT_RANGE 2>/dev/null | tail -1 | grep -o '[0-9]\+ insertion' | awk '{print $1}' || echo "0")
    DELETIONS=$(git diff --stat $COMMIT_RANGE 2>/dev/null | tail -1 | grep -o '[0-9]\+ deletion' | awk '{print $1}' || echo "0")
fi

# Get contributors
if [ "$INCLUDE_CONTRIBUTORS" = true ]; then
    CONTRIBUTORS=$(git log --format="%an <%ae>" $COMMIT_RANGE 2>/dev/null | sort -u)
    CONTRIBUTOR_COUNT=$(echo "$CONTRIBUTORS" | wc -l | tr -d ' ')
fi

# Generate release notes
{
    # Header
    if [ "$FORMAT" = "markdown" ]; then
        echo "# Release Notes for ${VERSION_TAG}"
        echo ""
        echo "_Released on $(date '+%B %d, %Y')_"
        echo ""

        # Statistics
        if [ "$INCLUDE_STATS" = true ]; then
            echo "## 📊 Statistics"
            echo ""
            echo "- **Commits:** ${TOTAL_COMMITS}"
            echo "- **Files Changed:** ${FILES_CHANGED}"
            echo "- **Lines Added:** ${INSERTIONS}"
            echo "- **Lines Removed:** ${DELETIONS}"
            echo "- **Contributors:** ${CONTRIBUTOR_COUNT}"
            echo ""
        fi

        # Breaking changes
        if [ ${#breaking[@]} -gt 0 ]; then
            echo "## ⚠️ Breaking Changes"
            echo ""
            printf '%s\n' "${breaking[@]}"
            echo ""
        fi

        # Features
        if [ ${#features[@]} -gt 0 ]; then
            echo "## ✨ New Features"
            echo ""
            printf '%s\n' "${features[@]}"
            echo ""
        fi

        # Improvements
        if [ ${#improvements[@]} -gt 0 ]; then
            echo "## 💪 Improvements"
            echo ""
            printf '%s\n' "${improvements[@]}"
            echo ""
        fi

        # Bug fixes
        if [ ${#fixes[@]} -gt 0 ]; then
            echo "## 🐛 Bug Fixes"
            echo ""
            printf '%s\n' "${fixes[@]}"
            echo ""
        fi

        # Performance
        if [ ${#performance[@]} -gt 0 ]; then
            echo "## ⚡ Performance"
            echo ""
            printf '%s\n' "${performance[@]}"
            echo ""
        fi

        # Refactoring
        if [ ${#refactors[@]} -gt 0 ]; then
            echo "## 🔨 Refactoring"
            echo ""
            printf '%s\n' "${refactors[@]}"
            echo ""
        fi

        # Removals
        if [ ${#removals[@]} -gt 0 ]; then
            echo "## 🗑️ Removed"
            echo ""
            printf '%s\n' "${removals[@]}"
            echo ""
        fi

        # Documentation
        if [ ${#docs[@]} -gt 0 ]; then
            echo "## 📚 Documentation"
            echo ""
            printf '%s\n' "${docs[@]}"
            echo ""
        fi

        # Other changes
        if [ ${#other[@]} -gt 0 ]; then
            echo "## 📝 Other Changes"
            echo ""
            printf '%s\n' "${other[@]}"
            echo ""
        fi

        # Contributors
        if [ "$INCLUDE_CONTRIBUTORS" = true ] && [ -n "$CONTRIBUTORS" ]; then
            echo "## 👥 Contributors"
            echo ""
            echo "Thanks to all contributors who made this release possible:"
            echo ""
            echo "$CONTRIBUTORS" | while read -r contributor; do
                name=$(echo "$contributor" | sed 's/ <.*//')
                echo "- $name"
            done
            echo ""
        fi

        # Installation
        echo "## 📦 Installation"
        echo ""
        echo "### macOS"
        echo ""
        echo "Download the appropriate DMG for your system:"
        echo "- **Apple Silicon (M1/M2/M3):** CCE-${VERSION_TAG}-arm64.dmg"
        echo "- **Intel:** CCE-${VERSION_TAG}-amd64.dmg"
        echo "- **Universal (works on both):** CCE-${VERSION_TAG}-universal.dmg"
        echo ""
        echo "### Verification"
        echo ""
        echo "To verify the downloaded file, check the SHA256 checksum:"
        echo "\`\`\`bash"
        echo "shasum -a 256 CCE-${VERSION_TAG}-*.dmg"
        echo "\`\`\`"
        echo ""

        # Links
        if [ -n "$REPO_URL" ]; then
            echo "## 🔗 Links"
            echo ""
            echo "- [Full Changelog](${REPO_URL}/compare/${PREVIOUS_TAG}...${VERSION_TAG})"
            echo "- [Download Release](${REPO_URL}/releases/tag/${VERSION_TAG})"
            echo ""
        fi

    else
        # Plain text format
        echo "RELEASE NOTES FOR ${VERSION_TAG}"
        echo "================================"
        echo "Released on $(date '+%B %d, %Y')"
        echo ""

        if [ "$INCLUDE_STATS" = true ]; then
            echo "STATISTICS"
            echo "----------"
            echo "Commits: ${TOTAL_COMMITS}"
            echo "Files Changed: ${FILES_CHANGED}"
            echo "Lines Added: ${INSERTIONS}"
            echo "Lines Removed: ${DELETIONS}"
            echo "Contributors: ${CONTRIBUTOR_COUNT}"
            echo ""
        fi

        if [ ${#breaking[@]} -gt 0 ]; then
            echo "BREAKING CHANGES"
            echo "----------------"
            printf '%s\n' "${breaking[@]}"
            echo ""
        fi

        if [ ${#features[@]} -gt 0 ]; then
            echo "NEW FEATURES"
            echo "------------"
            printf '%s\n' "${features[@]}"
            echo ""
        fi

        if [ ${#fixes[@]} -gt 0 ]; then
            echo "BUG FIXES"
            echo "---------"
            printf '%s\n' "${fixes[@]}"
            echo ""
        fi

        if [ ${#other[@]} -gt 0 ]; then
            echo "OTHER CHANGES"
            echo "-------------"
            printf '%s\n' "${other[@]}"
            echo ""
        fi
    fi
} > "$OUTPUT_FILE"

echo -e "${GREEN}✓ Release notes generated: ${OUTPUT_FILE}${NC}"

# Display summary
echo ""
echo "Summary:"
echo "  Features: ${#features[@]}"
echo "  Bug Fixes: ${#fixes[@]}"
echo "  Improvements: ${#improvements[@]}"
echo "  Breaking Changes: ${#breaking[@]}"
echo "  Other: ${#other[@]}"