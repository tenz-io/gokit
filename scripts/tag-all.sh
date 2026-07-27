#!/usr/bin/env bash
# tag-all.sh — One-shot batch tag + GitHub Release for all (or specified) v3
# submodules. All modules in this repo are on the v3 major track.
#
# Unlike a plain `git tag`/`git push` flow, this script talks to GitHub through
# the gh CLI: `gh release create <tag>` creates the git tag AND the GitHub
# Release in a single call (the tag is auto-created on the repo's default branch
# if it does not already exist). This avoids depending on a git remote pointing
# at github.com — `origin` can stay pointed at a self-hosted server.
#
# Usage:
#   ./scripts/tag-all.sh v3.0.1                          # preview (dry-run)
#   ./scripts/tag-all.sh v3.0.1 --release                # one-shot: tag + release for all modules
#   ./scripts/tag-all.sh v3.0.1 --push                   # tag only, no releases
#   ./scripts/tag-all.sh v3.0.1 --release --dry-run      # preview the release commands
#   ./scripts/tag-all.sh v3.0.1 --release tracer,logger  # specific modules only
#   ./scripts/tag-all.sh v3.0.1 --release --repo tenz-io/gokit
#   ./scripts/tag-all.sh v3.0.1 --release --notes-from-file NOTES.md
#   ./scripts/tag-all.sh v3.0.1 --release --no-overwrite # skip modules whose tag exists
#
# Tag format: <prefix>/<version>
#   e.g. tag annotation/v3.0.1 for module github.com/tenz-io/gokit/annotation/v3
#        tag logger/v3.0.1    for module github.com/tenz-io/gokit/logger/v3
# The tag prefix is the module path with the trailing /vN removed (Go's rule:
# module github.com/tenz-io/gokit/annotation/v3 is tagged annotation/v3.0.1,
# NOT annotation/v3/v3.0.1).
#
# Conflict policy: by default an existing tag+release is DELETED then recreated
# (overwrite). Use --no-overwrite to instead skip modules whose tag exists.
#
# Prerequisites:
#   - gh CLI (https://cli.github.com) installed and authenticated.
#       brew install gh && gh auth login
#   - push access to the target GitHub repo.
#   - `gh auth setup-git` is run by this script so git operations against
#     github.com authenticate via the gh token (no manual remote needed).

set -euo pipefail

VERSION=""
REPO="tenz-io/gokit"
DRY_RUN=false
PUSH=false
RELEASE=false
OVERWRITE=true
NOTES_FILE=""
MODULES=""

# --- arg parsing ----------------------------------------------------------
while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run)        DRY_RUN=true; shift ;;
    --push)           PUSH=true; shift ;;
    --release)        PUSH=true; RELEASE=true; shift ;;
    --no-overwrite)   OVERWRITE=false; shift ;;
    --repo)           REPO="${2:?--repo requires a value}"; shift 2 ;;
    --repo=*)         REPO="${1#--repo=}"; shift ;;
    --notes-from-file) NOTES_FILE="${2:?--notes-from-file requires a value}"; shift 2 ;;
    --notes-from-file=*) NOTES_FILE="${1#--notes-from-file=}"; shift ;;
    --help|-h)
      sed -n '2,/^$/p' "$0" | sed 's/^# \{0,1\}//'
      exit 0 ;;
    --*)
      echo "ERROR: unknown option: $1" >&2
      exit 2 ;;
    *)
      # A version number is recognized by its vN.M.P prefix. Any other
      # non-flag arg is the optional comma-separated module list.
      case "$1" in
        v[0-9]*) VERSION="$1" ;;
        *)       MODULES="$1" ;;
      esac
      shift ;;
  esac
done

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version> [--release|--push] [--dry-run] [--repo OWNER/REPO]"
  echo "                    [--notes-from-file FILE] [--no-overwrite] [mod1,mod2,...]"
  echo ""
  echo "Arguments:"
  echo "  <version>                  e.g. v3.0.1 (required)"
  echo "  mod1,mod2,...              optional subset of modules (default: all in go.work)"
  echo ""
  echo "Options:"
  echo "  --release                  Create git tags AND GitHub Releases (one-shot)"
  echo "  --push                     Create git tags only (no releases)"
  echo "  --dry-run                   Preview without creating anything"
  echo "  --repo OWNER/REPO           Target GitHub repo (default: $REPO)"
  echo "  --notes-from-file FILE      Use FILE contents as release notes for every module"
  echo "  --no-overwrite              Skip modules whose tag already exists (default: delete+recreate)"
  echo ""
  echo "Prerequisites:"
  echo "  gh auth login               # authenticate gh CLI to github.com"
  echo ""
  echo "Examples:"
  echo "  $0 v3.0.1 --release                  # one-shot all modules"
  echo "  $0 v3.0.1 --release --dry-run        # preview"
  echo "  $0 v3.0.1 --release tracer,logger    # specific modules"
  exit 1
fi

# --- preflight -----------------------------------------------------------
if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z-]+)?$ ]]; then
  echo "ERROR: version must look like v3.0.1 (got: $VERSION)" >&2
  exit 2
fi

if [ "$RELEASE" = false ] && [ "$PUSH" = false ]; then
  echo "ERROR: pass --release (tag+release) or --push (tag only). Without either,"
  echo "       nothing is created. Use --dry-run to preview without flags."
  exit 2
fi

if ! command -v gh &> /dev/null; then
  echo "ERROR: gh CLI not found. Install from https://cli.github.com"
  echo "  brew install gh && gh auth login"
  exit 1
fi

if [ "$DRY_RUN" != true ]; then
  if ! gh auth status &> /dev/null; then
    echo "ERROR: gh CLI not authenticated. Run: gh auth login" >&2
    exit 1
  fi
  # Let git operations against github.com use the gh token (idempotent).
  gh auth setup-git >/dev/null 2>&1 || true
  # Confirm we can reach the target repo.
  if ! gh repo view --repo "$REPO" >/dev/null 2>&1; then
    echo "ERROR: cannot access repo '$REPO'. Check --repo and gh auth scope." >&2
    exit 1
  fi
fi

# Optional custom release notes (read once, reused for every module).
CUSTOM_NOTES=""
if [ -n "$NOTES_FILE" ]; then
  if [ ! -f "$NOTES_FILE" ]; then
    echo "ERROR: notes file not found: $NOTES_FILE" >&2
    exit 1
  fi
  CUSTOM_NOTES="$(cat "$NOTES_FILE")"
fi

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

# --- build module list ----------------------------------------------------
# Each entry is the full module path suffix, e.g. "annotation/v3".
if [ -n "$MODULES" ]; then
  IFS=',' read -ra MODS <<< "$MODULES"
  printf '%s\n' "${MODS[@]}"
else
  grep '\./' go.work | grep -v 'use (' | grep -v '^)' \
    | tr -d $' \t\r' | grep -v '/example' | grep -v '/example-' \
    | sed 's|^\./||'
fi > /tmp/tag_mods.txt

# Plan lines: "<module> <tag_prefix> <version>".
: > /tmp/tag_plan.txt
while IFS= read -r mod; do
  [ -z "$mod" ] && continue
  prefix=$(echo "$mod" | sed -E 's|/v[0-9]+$||')
  echo "$mod $prefix $VERSION" >> /tmp/tag_plan.txt
done < /tmp/tag_mods.txt

TOTAL=$(wc -l < /tmp/tag_plan.txt | tr -d ' ')
echo "Repo:    $REPO"
ACTION=""
if [ "$RELEASE" = true ]; then ACTION="tag + release"; else ACTION="tag only"; fi
OVERWRITE_STR=$( [ "$OVERWRITE" = true ] && echo yes || echo no )
echo "Action:  $ACTION  (overwrite=$OVERWRITE_STR)"
echo "Version: $VERSION"
echo "Modules to tag ($TOTAL):"
awk '{printf "  %-20s ->  %s/%s\n", $1, $2, $3}' /tmp/tag_plan.txt
echo "---"

if [ "$TOTAL" -eq 0 ]; then
  echo "Nothing to tag."
  rm -f /tmp/tag_mods.txt /tmp/tag_plan.txt
  exit 0
fi

if [ "$DRY_RUN" = true ]; then
  while IFS=' ' read -r mod prefix ver; do
    [ -z "$mod" ] && continue
    TAG="$prefix/$ver"
    echo "[DRY RUN] check tag: gh api repos/$REPO/git/refs/tags/$TAG"
    if [ "$RELEASE" = true ]; then
      echo "[DRY RUN] gh release create $TAG --repo $REPO --title '$mod $ver' --latest=false"
    else
      echo "[DRY RUN] gh release create $TAG --repo $REPO --title '$mod $ver (tag)' --notes '' --latest=false"
    fi
  done < /tmp/tag_plan.txt
  rm -f /tmp/tag_mods.txt /tmp/tag_plan.txt
  echo ""
  echo "Dry run only. To execute: ./scripts/tag-all.sh $VERSION --release"
  exit 0
fi

# --- execute --------------------------------------------------------------
# tag_exists <tag> -> 0 if the tag exists on GitHub, 1 otherwise.
tag_exists() {
  gh api "repos/$REPO/git/refs/tags/$1" >/dev/null 2>&1
}

# release_exists <tag> -> 0 if a release is associated with the tag.
release_exists() {
  gh api "repos/$REPO/releases/tags/$1" >/dev/null 2>&1
}

CREATED=0; OVERWRITTEN=0; SKIPPED=0; FAILED=0

while IFS=' ' read -r mod prefix ver; do
  [ -z "$mod" ] && continue
  TAG="$prefix/$ver"

  # Default release notes: per-module install template.
  if [ -n "$CUSTOM_NOTES" ]; then
    NOTES="$CUSTOM_NOTES"
  else
    NOTES=$(cat <<EOF
## ${mod} ${ver}

Go module: \`github.com/tenz-io/gokit/${mod}\`

\`\`\`
go get github.com/tenz-io/gokit/${mod}@${ver}
\`\`\`
EOF
)
  fi

  # Overwrite path: delete existing tag + release first.
  did_overwrite=false
  if [ "$OVERWRITE" = true ] && tag_exists "$TAG"; then
    if release_exists "$TAG"; then
      gh release delete "$TAG" --cleanup-tag -y --repo "$REPO" >/dev/null 2>&1 \
        || gh release delete "$TAG" -y --repo "$REPO" >/dev/null 2>&1 || true
    fi
    # --cleanup-tag should have removed the tag ref; if not, delete it directly.
    if tag_exists "$TAG"; then
      gh api --method DELETE "repos/$REPO/git/refs/tags/$TAG" >/dev/null 2>&1 || true
    fi
    # Drop any local copy so we don't carry a stale ref.
    git tag -d "$TAG" 2>/dev/null || true
    did_overwrite=true
  fi

  # Skip path: tag already exists and user asked not to overwrite.
  if [ "$OVERWRITE" = false ] && tag_exists "$TAG"; then
    echo "  Skip     $TAG (already exists, --no-overwrite)"
    SKIPPED=$((SKIPPED + 1))
    continue
  fi

  # Create tag + release in one gh call. gh auto-creates the git tag on the
  # repo's default branch when it does not yet exist. For --push (tag only),
  # we still go through gh release create (gh has no tag-without-release
  # subcommand); the resulting release carries an empty body. Pass --notes
  # even for --push so --push later upgraded to --release needs no extra work.
  if [ "$RELEASE" = true ]; then
    TITLE="$mod $ver"
  else
    TITLE="$mod $ver (tag)"
  fi

  if gh release create "$TAG" \
      --repo "$REPO" \
      --title "$TITLE" \
      --notes "$NOTES" \
      --latest=false >/dev/null 2>&1; then
    if [ "$did_overwrite" = true ]; then
      echo "  Overwrite $TAG"
      OVERWRITTEN=$((OVERWRITTEN + 1))
    else
      echo "  Create   $TAG"
      CREATED=$((CREATED + 1))
    fi
  else
    echo "  ERROR    $TAG — gh release create failed" >&2
    FAILED=$((FAILED + 1))
  fi
done < /tmp/tag_plan.txt

rm -f /tmp/tag_mods.txt /tmp/tag_plan.txt

# --- summary --------------------------------------------------------------
echo ""
echo "=== Summary ==="
echo "  created:     $CREATED"
echo "  overwritten: $OVERWRITTEN"
echo "  skipped:     $SKIPPED"
echo "  failed:      $FAILED"
echo "  repo:        $REPO"
if [ "$RELEASE" = true ]; then
  echo "  releases:    https://github.com/$REPO/releases"
fi
echo "Done."

[ "$FAILED" -eq 0 ]
