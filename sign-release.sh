#!/usr/bin/env bash
set -e

KEY="uil-release.key"
PUB="uil-release.key.pub"
DIST_DIR="dist"

echo "=== UIL Release Signing Script ==="

# 1. Check keypair
if [[ ! -f "$KEY" || ! -f "$PUB" ]]; then
    echo "ERROR: Minisign keypair not found."
    echo "Expected: $KEY and $PUB"
    echo "Generate with:"
    echo "  minisign -G -s uil-release.key -p uil-release.key.pub"
    exit 1
fi

echo "✔ Minisign keypair found."

# 2. Ensure dist/ exists
if [[ ! -d "$DIST_DIR" ]]; then
    echo "ERROR: dist/ directory not found."
    exit 1
fi

# 3. Fix ownership (avoids root-owned build artifacts)
echo "✔ Ensuring dist/ is owned by current user..."
sudo chown -R "$USER:$USER" "$DIST_DIR"

# 4. Sign all tarballs
echo "✔ Signing release tarballs..."
for file in "$DIST_DIR"/*.tar.gz; do
    if [[ -f "$file" ]]; then
        echo "→ Signing $file"
        minisign -S -s "$KEY" -m "$file"
        echo "   ✔ Signature created: $file.minisig"
    fi
done

echo "=== Signing Complete ==="
echo "Upload BOTH the .tar.gz and .minisig files to GitHub Releases."
