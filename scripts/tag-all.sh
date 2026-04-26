#!/usr/bin/env bash
# tag-all.sh — Batch tag all (or specified) submodules with a single version.
#
# Usage:
#   ./scripts/tag-all.sh v2.0.1                  # tag all non-example submodules
#   ./scripts/tag-all.sh v2.0.1 --dry-run        # preview what would happen
#   ./scripts/tag-all.sh v2.0.1 tracer,logger    # tag specific modules only
#   ./scripts/tag-all.sh v2.0.1 --push           # tag AND push to origin
#   ./scripts/tag-all.sh v2.0.1 --release        # push AND create GitHub Releases
#
# For Go module resolution, the tag format is: <module>/v<version>
# e.g. tag annotation/v2.0.1 for module github.com/tenz-io/gokit/annotation/v2
#
# Prerequisites:
#   - git push access to the repository
#   - gh CLI (https://cli.github.com) installed and authenticated (for --release)

set -euo pipefail

VERSION="${1:-}"
shift 2>/dev/null || true

DRY_RUN=false
PUSH=false
RELEASE=false
MODULES=""

while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run) DRY_RUN=true; shift ;;
    --push)    PUSH=true; shift ;;
    --release) PUSH=true; RELEASE=true; shift ;;
    *)         MODULES="$1"; shift ;;
  esac
done

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version> [--dry-run] [--push] [--release] [mod1,mod2,...]"
  echo ""
  echo "Options:"
  echo "  --dry-run   Preview tags without creating them"
  echo "  --push      Create tags and push to origin"
  echo "  --release   Create tags, push, and create GitHub Releases (requires gh CLI)"
  echo ""
  echo "Examples:"
  echo "  $0 v2.0.1 --dry-run"
  echo "  $0 v2.0.1 tracer,logger,async --push"
  echo "  $0 v2.0.1 --release"
  exit 1
fi

# Check gh CLI if --release is requested
if [ "$RELEASE" = true ] && [ "$DRY_RUN" != true ]; then
  if ! command -v gh &> /dev/null; then
    echo "ERROR: gh CLI not found. Install from https://cli.github.com"
    echo "  brew install gh    # macOS"
    echo "  gh auth login      # authenticate"
    exit 1
  fi
  if ! gh auth status &> /dev/null; then
    echo "ERROR: gh CLI not authenticated. Run: gh auth login"
    exit 1
  fi
fi

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

# Choose submodule source
if [ -n "$MODULES" ]; then
  IFS=',' read -ra MODS <<< "$MODULES"
  printf '%s\n' "${MODS[@]}" > /tmp/tag_mods.txt
else
  grep '\./' go.work | grep -v 'use (' | grep -v '^)' \
    | tr -d $' \t\r' | grep -v '/example' \
    | sed 's|^\./||' > /tmp/tag_mods.txt
fi

echo "Version: $VERSION"
echo "Submodules ($(wc -l < /tmp/tag_mods.txt | tr -d ' ')):"
cat /tmp/tag_mods.txt
echo "---"

TAG_COUNT=0
TAG_FILE=$(mktemp /tmp/created_tags.XXXXXX)

while IFS= read -r mod; do
  [ -z "$mod" ] && continue

  # Normalize: if module dir is "logger/v2", tag is "logger/v2.0.1" (not "logger/v2/v2.0.1")
  # Go convention: tag prefix = module path minus /vN suffix
  TAG_MOD=$(echo "$mod" | sed 's|/v[0-9][0-9]*$||')
  TAG="${TAG_MOD}/${VERSION}"

  if [ "$DRY_RUN" = true ]; then
    echo "[DRY RUN] git tag -a $TAG -m 'Release ${TAG_MOD} ${VERSION}'"
    if [ "$RELEASE" = true ]; then
      echo "[DRY RUN] gh release create $TAG --title '${TAG_MOD} ${VERSION}' --notes ''"
    fi
  else
    # Force-overwrite local tag to point to current HEAD
    git tag -d "$TAG" 2>/dev/null || true
    git tag -a "$TAG" -m "Release ${TAG_MOD} ${VERSION}" -f
    echo "Created: $TAG"
    echo "$TAG" >> "$TAG_FILE"
    TAG_COUNT=$((TAG_COUNT + 1))
  fi
done < /tmp/tag_mods.txt

# Push tags
if [ "$PUSH" = true ] && [ "$DRY_RUN" != true ]; then
  echo ""
  echo "Pushing tags to origin..."
  # Push only the tags we just created (avoids "already exists" errors from old tags)
  if git push origin --force $(tr '\n' ' ' < "$TAG_FILE"); then
    echo "Tags pushed ($TAG_COUNT tags)"
  else
    echo "ERROR: Failed to push tags. Check your git remote and authentication."
    rm -f "$TAG_FILE"
    exit 1
  fi

  # Create GitHub Releases
  if [ "$RELEASE" = true ]; then
    echo ""
    echo "Creating GitHub Releases..."

    while IFS= read -r mod; do
      [ -z "$mod" ] && continue
      TAG_MOD=$(echo "$mod" | sed 's|/v[0-9][0-9]*$||')
      TAG="${TAG_MOD}/${VERSION}"

      RELEASE_NOTES=$(cat <<EOF
## ${TAG_MOD} ${VERSION}

Go module: \`github.com/tenz-io/gokit/${mod}\`

\`\`\`
go get github.com/tenz-io/gokit/${mod}@${VERSION}
\`\`\`
EOF
)

      if gh release create "$TAG" \
        --title "${TAG_MOD} ${VERSION}" \
        --notes "$RELEASE_NOTES" \
        --latest=false 2>/dev/null; then
        echo "  Release: $TAG"
      else
        echo "  Skipped (may already exist): $TAG"
      fi
    done < /tmp/tag_mods.txt

    echo ""
    echo "Releases created. View at: https://github.com/tenz-io/gokit/releases"
  fi
fi

rm -f "$TAG_FILE"

echo ""
echo "Done."
if [ "$DRY_RUN" != true ] && [ "$PUSH" = false ]; then
  echo "Tags created locally. To push: ./scripts/tag-all.sh $VERSION --push"
  echo "To create GitHub Releases too: ./scripts/tag-all.sh $VERSION --release"
fi
