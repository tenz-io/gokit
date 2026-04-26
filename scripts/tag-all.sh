#!/usr/bin/env bash
# tag-all.sh — Batch tag all (or specified) submodules with a single version.
#
# Usage:
#   ./scripts/tag-all.sh v2.0.1                  # tag all non-example submodules
#   ./scripts/tag-all.sh v2.0.1 --dry-run        # preview what would happen
#   ./scripts/tag-all.sh v2.0.1 tracer,logger    # tag specific modules only
#   ./scripts/tag-all.sh v2.0.1 --push           # tag AND push to origin
#
# For Go module resolution, the tag format is: <module>/v<version>
# e.g. tag annotation/v2.0.1 for module github.com/tenz-io/gokit/annotation

set -euo pipefail

VERSION="${1:-}"
shift 2>/dev/null || true

DRY_RUN=false
PUSH=false
MODULES=""

while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run) DRY_RUN=true; shift ;;
    --push)    PUSH=true; shift ;;
    *)         MODULES="$1"; shift ;;
  esac
done

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version> [--dry-run] [--push] [mod1,mod2,...]"
  echo "Example: $0 v2.0.1 --dry-run"
  echo "Example: $0 v2.0.1 tracer,logger,async --push"
  exit 1
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

while IFS= read -r mod; do
  [ -z "$mod" ] && continue

  # Normalize: if module dir is "logger/v2", tag is "logger/v2.0.1" (not "logger/v2/v2.0.1")
  # Go convention: tag prefix = module path minus /vN suffix
  TAG_MOD=$(echo "$mod" | sed 's|/v[0-9][0-9]*$||')
  TAG="${TAG_MOD}/${VERSION}"

  if [ "$DRY_RUN" = true ]; then
    echo "[DRY RUN] git tag -a $TAG -m 'Release ${TAG_MOD} ${VERSION}'"
  else
    git tag -a "$TAG" -m "Release ${TAG_MOD} ${VERSION}"
    echo "Created: $TAG"
  fi
done < /tmp/tag_mods.txt

if [ "$PUSH" = true ] && [ "$DRY_RUN" != true ]; then
  git push origin --tags
  echo "Tags pushed to origin"
fi
