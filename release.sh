#!/usr/bin/env bash
set -e

VERSION="$1"
KEY="uil-release.key"
PUB="uil-release.key.pub"
DIST="dist"

if [[ -z "$VERSION" ]]; then
    echo "Usage: ./release.sh v0.8.1.1-alpha"
    exit 1
fi

echo "=== UIL Release Builder & Signer ==="
echo "Target version: $VERSION"
echo

# 1. Check keypair
if [[ ! -f "$KEY" || ! -f "$PUB" ]]; then
    echo "ERROR: Minisign keypair not found."
    echo "Expected: $KEY and $PUB"
    echo "Generate with:"
    echo "  minisign -G -s uil-release.key -p uil-release.key.pub"
    exit 1
fi
echo "✔ Minisign keypair found."

# 2. Build release tarballs
echo "✔ Building release artifacts..."
make dist VERSION="$VERSION"

# 3. Fix ownership
echo "✔ Ensuring dist/ is owned by current user..."
sudo chown -R "$USER:$USER" "$DIST"

# 4. Sign all tarballs
echo "✔ Signing release tarballs..."
for file in "$DIST"/*.tar.gz; do
    if [[ -f "$file" ]]; then
        echo "→ Signing $file"
        minisign -S -s "$KEY" -m "$file"
        echo "   ✔ Signature created: $file.minisig"
    fi
done

# 5. Verify signatures
echo "✔ Verifying signatures..."
for file in "$DIST"/*.tar.gz; do
    sig="$file.minisig"
    echo "→ Verifying $file"
    minisign -V -p "$PUB" -m "$file" -x "$sig"
    echo "   ✔ Verified"
done

# 6. Create GitHub release
echo "✔ Creating GitHub release $VERSION"
gh release create "$VERSION" "$DIST"/* \
    --title "UIL-X $VERSION" \
    --notes "UIL-X Hardware Governance Daemon release $VERSION" 

echo
echo "=== Release Complete ==="
echo "Version: $VERSION"
echo "Uploaded files:"
ls -1 "$DIST"
echo
echo "GitHub Release created successfully."
