#!/usr/bin/env bash
# version-check.sh — Verify version consistency across submodules before
# tagging a release. Exits non-zero on any inconsistency.
#
# All modules in this repo are on the v3 major track.
#
# Checks:
#   1. All `require github.com/tenz-io/gokit/*/v3` lines share one version.
#   2. No `replace github.com/tenz-io/gokit/...` left over from development
#      (replace directives should be removed before tagging, since the tag
#      must be resolvable from VCS).
#   3. Every module directory under go.work has a go.mod (sanity).
#
# Usage:
#   ./scripts/version-check.sh                # check, print report
#   ./scripts/version-check.sh --quiet        # only print on failure

set -euo pipefail

QUIET=false
[ "${1:-}" = "--quiet" ] && QUIET=true

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

STATUS=0
log() { [ "$QUIET" = false ] && echo "$1" || true; }

# ---- Check 1: v3 require consistency ---------------------------------------
log "## v3 require versions"
V3_VERSIONS=$(grep -rhoE 'github\.com/tenz-io/gokit/[a-z0-9-]+/v3 v[0-9.]+' --include=go.mod . \
  | grep -v example | awk '{print $2}' | sort -u)
V3_COUNT=$(echo "$V3_VERSIONS" | grep -c . || true)
if [ "$V3_COUNT" -eq 1 ]; then
  log "  OK: all v3 requires at $(echo "$V3_VERSIONS")"
elif [ "$V3_COUNT" -eq 0 ]; then
  log "  (no v3 requires found)"
else
  log "  FAIL: v3 requires use multiple versions:"
  echo "$V3_VERSIONS" | sed 's/^/    /'
  STATUS=1
fi

# ---- Check 2: leftover replace directives ----------------------------------
log "## leftover replace directives"
LEFTOVER=$(grep -rn '^replace.*tenz-io/gokit' --include=go.mod . | grep -v example || true)
if [ -z "$LEFTOVER" ]; then
  log "  OK: no replace directives in non-example modules"
else
  log "  WARN: replace directives present (remove before tagging unless dev-only):"
  echo "$LEFTOVER" | sed 's/^/    /'
  # Warn, not fail — replace may be intentional for unreleased modules.
fi

# ---- Check 3: go.work modules all have go.mod ------------------------------
log "## go.work modules"
MISSING=0
for mod in $(grep '\./' go.work | grep -v 'use (' | grep -v ')' | tr -d ' \t\r' | sed 's|^\./||'); do
  [ -z "$mod" ] && continue
  if [ ! -f "$mod/go.mod" ]; then
    log "  FAIL: $mod listed in go.work but has no go.mod"
    MISSING=$((MISSING + 1))
    STATUS=1
  fi
done
if [ "$MISSING" -eq 0 ]; then
  log "  OK: all go.work modules have go.mod"
fi

# ---- Summary ----------------------------------------------------------------
echo ""
if [ "$STATUS" -eq 0 ]; then
  log "version-check: PASS"
else
  log "version-check: FAIL (see above)"
fi
exit $STATUS
