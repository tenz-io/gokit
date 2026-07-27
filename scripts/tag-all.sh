#!/usr/bin/env bash
# tag-all.sh — Batch tag all (or specified) submodules.
#
# All modules in this repo are on the v3 major track. This script tags every
# non-example module listed in go.work with a single version number.
#
# Usage:
#   ./scripts/tag-all.sh v3.0.1                        # tag all modules
#   ./scripts/tag-all.sh v3.0.1 --dry-run              # preview what would happen
#   ./scripts/tag-all.sh v3.0.1 --push                 # tag AND push to origin
#   ./scripts/tag-all.sh v3.0.1 --release              # push AND create GitHub Releases
#   ./scripts/tag-all.sh v3.0.1 tracer,logger,async    # tag specific modules only
#
# Tag format: <prefix>/<version>
#   e.g. tag annotation/v3.0.1 for module github.com/tenz-io/gokit/annotation/v3
#        tag logger/v3.0.1    for module github.com/tenz-io/gokit/logger/v3
# The tag prefix is the module path with the trailing /vN removed (Go's rule:
# module github.com/tenz-io/gokit/annotation/v3 is tagged annotation/v3.0.1,
# NOT annotation/v3/v3.0.1).
#
# Prerequisites:
#   - git push access to the repository
#   - gh CLI (https://cli.github.com) installed and authenticated (for --release)

set -euo pipefail

VERSION=""
DRY_RUN=false
PUSH=false
RELEASE=false
MODULES=""

# Parse the single version arg, then flags / module list.
while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run) DRY_RUN=true; shift ;;
    --push)    PUSH=true; shift ;;
    --release) PUSH=true; RELEASE=true; shift ;;
    --help|-h)
      sed -n '2,/^$/p' "$0" | sed 's/^# \{0,1\}//'
      exit 0 ;;
    *)
      # A version number is recognized by its v. prefix. Any other non-flag
      # arg is the optional comma-separated module list.
      case "$1" in
        v*.*) VERSION="$1" ;;
        *)    MODULES="$1" ;;
      esac
      shift ;;
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
  echo "  $0 v3.0.1 --dry-run"
  echo "  $0 v3.0.1 --push"
  echo "  $0 v3.0.1 tracer,logger,async --push"
  exit 1
fi

# Check gh CLI if --release is requested
if [ "$RELEASE" = true ] && [ "$DRY_RUN" != true ]; then
  if ! command -v gh &> /dev/null; then
    echo "ERROR: gh CLI not found. Install from https://cli.github.com"
    echo "  brew install gh && gh auth login"
    exit 1
  fi
  if ! gh auth status &> /dev/null; then
    echo "ERROR: gh CLI not authenticated. Run: gh auth login"
    exit 1
  fi
fi

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

# Build the module list from go.work (excluding example modules) when no explicit
# list is given. Each entry is the full module path suffix, e.g. "annotation/v3".
# The /vN suffix is stripped below to form each tag prefix.
if [ -n "$MODULES" ]; then
  IFS=',' read -ra MODS <<< "$MODULES"
  printf '%s\n' "${MODS[@]}" > /tmp/tag_mods.txt
else
  grep '\./' go.work | grep -v 'use (' | grep -v '^)' \
    | tr -d $' \t\r' | grep -v '/example' | grep -v '/example-' \
    | sed 's|^\./||' > /tmp/tag_mods.txt
fi

# For each module, strip the trailing /vN to form the tag prefix (Go's rule:
# module github.com/tenz-io/gokit/annotation/v3 is tagged annotation/v3.0.1,
# NOT annotation/v3/v3.0.1). Writes "<module_path> <tag_prefix> <version>".
: > /tmp/tag_plan.txt
while IFS= read -r mod; do
  [ -z "$mod" ] && continue
  prefix=$(echo "$mod" | sed -E 's|/v[0-9]+$||')
  echo "$mod $prefix $VERSION" >> /tmp/tag_plan.txt
done < /tmp/tag_mods.txt

TOTAL=$(wc -l < /tmp/tag_plan.txt | tr -d ' ')
echo "Version: $VERSION"
echo "Submodules to tag ($TOTAL):"
awk '{printf "  %s  ->  %s/%s\n", $1, $2, $3}' /tmp/tag_plan.txt
echo "---"

if [ "$TOTAL" -eq 0 ]; then
  echo "Nothing to tag."
  rm -f /tmp/tag_mods.txt /tmp/tag_plan.txt
  exit 0
fi

TAG_FILE=$(mktemp /tmp/created_tags.XXXXXX)

if [ "$DRY_RUN" = true ]; then
  while IFS=' ' read -r mod prefix ver; do
    [ -z "$mod" ] && continue
    echo "[DRY RUN] git tag -a $prefix/$ver -m 'Release $mod $ver'"
    if [ "$RELEASE" = true ]; then
      echo "[DRY RUN] gh release create $prefix/$ver --title '$mod $ver' --notes ''"
    fi
  done < /tmp/tag_plan.txt
else
  while IFS=' ' read -r mod prefix ver; do
    [ -z "$mod" ] && continue
    TAG="$prefix/$ver"
    git tag -d "$TAG" 2>/dev/null || true
    git tag -a "$TAG" -m "Release $mod $ver" -f
    echo "Created: $TAG"
    echo "$TAG" >> "$TAG_FILE"
  done < /tmp/tag_plan.txt

  # Push tags
  if [ "$PUSH" = true ]; then
    echo ""
    echo "Pushing tags to origin..."
    if git push origin --force $(tr '\n' ' ' < "$TAG_FILE"); then
      echo "Tags pushed ($(wc -l < "$TAG_FILE" | tr -d ' ') tags)"
    else
      echo "ERROR: Failed to push tags. Check your git remote and authentication."
      rm -f "$TAG_FILE" /tmp/tag_mods.txt /tmp/tag_plan.txt
      exit 1
    fi

    # Create GitHub Releases
    if [ "$RELEASE" = true ]; then
      echo ""
      echo "Creating GitHub Releases..."
      while IFS=' ' read -r mod prefix ver; do
        [ -z "$mod" ] && continue
        TAG="$prefix/$ver"
        RELEASE_NOTES=$(cat <<EOF
## ${mod} ${ver}

Go module: \`github.com/tenz-io/gokit/${mod}\`

\`\`\`
go get github.com/tenz-io/gokit/${mod}@${ver}
\`\`\`
EOF
)
        if gh release create "$TAG" \
          --title "${mod} ${ver}" \
          --notes "$RELEASE_NOTES" \
          --latest=false 2>/dev/null; then
          echo "  Release: $TAG"
        else
          echo "  Skipped (may already exist): $TAG"
        fi
      done < /tmp/tag_plan.txt
      echo ""
      echo "Releases created. View at: https://github.com/tenz-io/gokit/releases"
    fi
  fi
fi

rm -f "$TAG_FILE" /tmp/tag_mods.txt /tmp/tag_plan.txt

echo ""
echo "Done."
if [ "$DRY_RUN" != true ] && [ "$PUSH" = false ]; then
  echo "Tags created locally. To push: ./scripts/tag-all.sh $VERSION --push"
  [ "$RELEASE" = true ] && echo "To create GitHub Releases too: add --release"
fi
