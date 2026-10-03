#!/usr/bin/env bash
# Post-deploy smoke for the Go API tier (immortalvibes-api on Fly).
#
# Deliberately NOT just /health. `api/middleware/auth.go` exempts /health and
# /api/webhooks/stripe, so /health returns 200 even when PROXY_SECRET is
# completely wrong — it cannot distinguish a healthy API from an unreachable one.
# Each assertion below can fail on its own:
#
#   1. /health 200                      — the process is up at all
#   2. direct /api/products -> 403      — the proxy-secret middleware IS enforcing
#   3. storefront /api/products -> 200  — the real client path (Pages -> Fly) agrees
#   4. that payload parses and is non-empty — 200 with a broken body is not a pass
#
# Assertion 3 is the one that catches a secret mismatch between the two sides;
# assertion 2 is the one that catches auth being accidentally disabled.
set -uo pipefail

# Overridable so the FAILURE path can be positive-controlled; a gate that can only
# pass is not a gate. CI leaves both unset and hits production.
API="${SMOKE_API:-https://immortalvibes-api.fly.dev}"
LIVE="${SMOKE_LIVE:-https://theimmortalvibes.com}"
FAILED=0

code() { curl -s -o /dev/null -w '%{http_code}' --max-time 20 "$1"; }

# Machines may still be starting right after a deploy — retry, but only on the
# way UP. A wrong code that is stable is a failure, not a slow start.
expect() {
  local label="$1" url="$2" want="$3" got=""
  for _ in 1 2 3 4 5 6; do
    got="$(code "$url")"
    [ "$got" = "$want" ] && break
    sleep 5
  done
  if [ "$got" = "$want" ]; then
    printf '  PASS  %-34s %s\n' "$label" "$got"
  else
    printf '  FAIL  %-34s got %s, want %s\n' "$label" "$got" "$want"
    FAILED=1
  fi
}

echo "API smoke:"
expect "/health"                      "$API/health"          200
expect "direct /api/products (no hdr)" "$API/api/products"    403
expect "storefront /api/products"      "$LIVE/api/products"   200

# 200 with an unusable body is not a pass.
BODY="$(curl -s --max-time 20 "$LIVE/api/products")"
if N="$(printf '%s' "$BODY" | jq 'if type=="array" then length elif type=="object" then (.products // [] | length) else 0 end' 2>/dev/null)" \
   && [ -n "$N" ] && [ "$N" -ge 1 ]; then
  printf '  PASS  %-34s %s products\n' "payload parses, non-empty" "$N"
else
  printf '  FAIL  %-34s body did not parse as a non-empty product list\n' "payload parses, non-empty"
  FAILED=1
fi

[ "$FAILED" -eq 0 ] && echo "API smoke: PASS" || echo "API smoke: FAIL"
exit "$FAILED"
