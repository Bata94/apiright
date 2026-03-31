#!/usr/bin/env bash
set -euo pipefail

TYPE="${1:-patch}"
VERSION_FILE="VERSION"

# Read current version
CURRENT_VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')
if [[ ! "$CURRENT_VERSION" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    echo "ERROR: Invalid version format in $VERSION_FILE: $CURRENT_VERSION"
    exit 1
fi

MAJOR="${BASH_REMATCH[1]}"
MINOR="${BASH_REMATCH[2]}"
PATCH="${BASH_REMATCH[3]}"

# Calculate new version
case "$TYPE" in
    patch)
        PATCH=$((PATCH + 1))
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    *)
        echo "ERROR: Invalid release type: $TYPE"
        echo "Valid types: patch, minor, major"
        exit 1
        ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

echo "Current version: $CURRENT_VERSION"
echo "New version: $NEW_VERSION"

# Update VERSION file
echo "$NEW_VERSION" > "$VERSION_FILE"
echo "Updated $VERSION_FILE"

# Commit version change
git add "$VERSION_FILE"
git commit -m "chore: release $NEW_VERSION"
echo "Created commit: $(git rev-parse --short HEAD)"

# Generate release notes from commits since last tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -n "$LAST_TAG" ]; then
    RELEASE_NOTES=$(git log "$LAST_TAG"..HEAD --format="- %s" 2>/dev/null || echo "Release $NEW_VERSION")
else
    # No previous tag, use all commits
    RELEASE_NOTES=$(git log --format="- %s" | head -20)
fi

# Create annotated tag
git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION

$RELEASE_NOTES
"
echo "Created tag: $NEW_VERSION"

# Push
echo "Pushing commits and tags..."
git push
git push origin "$NEW_VERSION"
echo "✅ Release $NEW_VERSION created and pushed"
