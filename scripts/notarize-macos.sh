#!/bin/bash
# Notarize macOS .app bundle for Universal-Ops
# Requires the following environment variables:
#   APPLE_SIGNING_IDENTITY      — e.g. "Developer ID Application: Your Name (TEAMID)"
#   APPLE_CERTIFICATE           — base64-encoded .p12 signing certificate
#   APPLE_CERTIFICATE_PASSWORD  — password for the .p12 certificate
#   APPLE_ID                    — Apple ID email
#   APPLE_TEAM_ID               — Apple Team ID (10-char alphanumeric)
#   APPLE_NOTARIZATION_PASSWORD — App-specific password for notary service
#
# Usage: ./notarize-macos.sh <path/to/universal-ops.app>

set -euo pipefail

APP_PATH="${1:?Usage: $0 <path/to/app>}"
APP_DIR="$(dirname "$APP_PATH")"
APP_NAME="$(basename "$APP_PATH")"
ARCHIVE="/tmp/notarize-${APP_NAME}.zip"

echo "[notarize] Signing certificate available: ${APPLE_SIGNING_IDENTITY:+yes}${APPLE_SIGNING_IDENTITY:-no}"

# ── 1. Import signing certificate ──────────────────────────────────
if [ -n "${APPLE_SIGNING_IDENTITY:-}" ]; then
  echo "$APPLE_CERTIFICATE" | base64 -d > /tmp/cert.p12 2>/dev/null
  KEYCHAIN="notarize-$$.keychain"
  security create-keychain -p temp "$KEYCHAIN"
  security default-keychain -s "$KEYCHAIN"
  security unlock-keychain -p temp "$KEYCHAIN"
  security import /tmp/cert.p12 -k "$KEYCHAIN" -P "${APPLE_CERTIFICATE_PASSWORD:-}" -T /usr/bin/codesign 2>/dev/null
  security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k temp "$KEYCHAIN"
  rm -f /tmp/cert.p12
fi

# ── 2. Sign the .app bundle ────────────────────────────────────────
if [ -n "${APPLE_SIGNING_IDENTITY:-}" ]; then
  echo "[notarize] Signing ${APP_NAME} …"
  codesign --force --options runtime --sign "$APPLE_SIGNING_IDENTITY" --deep "$APP_PATH"
else
  echo "[notarize] APPLE_SIGNING_IDENTITY not set — skipping code signing (notarization will fail)"
fi

# ── 3. Submit to Apple Notary Service ──────────────────────────────
if [ -n "${APPLE_ID:-}" ] && [ -n "${APPLE_TEAM_ID:-}" ] && [ -n "${APPLE_NOTARIZATION_PASSWORD:-}" ]; then
  echo "[notarize] Creating archive …"
  ditto -c -k --sequesterRsrc --keepParent "$APP_PATH" "$ARCHIVE"

  echo "[notarize] Submitting to notary service …"
  xcrun notarytool submit "$ARCHIVE" \
    --apple-id "$APPLE_ID" \
    --team-id "$APPLE_TEAM_ID" \
    --password "$APPLE_NOTARIZATION_PASSWORD" \
    --wait

  rm -f "$ARCHIVE"

  # ── 4. Staple the ticket ─────────────────────────────────────────
  echo "[notarize] Stapling ticket …"
  xcrun stapler staple "$APP_PATH"
else
  echo "[notarize] APPLE_ID / APPLE_TEAM_ID / APPLE_NOTARIZATION_PASSWORD not set — skipping notarization"
fi

echo "[notarize] Done."
