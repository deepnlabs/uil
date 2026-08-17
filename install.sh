#!/usr/bin/env bash
set -euo pipefail

REPO="deepnlabs/uil"
TARGET_ARCH="$(uname -m)"
TMP_DIR="$(mktemp -d)"

# Map uname -m to release arch names
case "$TARGET_ARCH" in
  x86_64) TARGET_ARCH="amd64" ;;
  aarch64) TARGET_ARCH="arm64" ;;
  armv6l|armv7l|arm) TARGET_ARCH="armv6" ;;
  *) echo "❌ Unsupported architecture: $TARGET_ARCH"; exit 1 ;;
esac

echo "=> Resolving pre-compiled release binary for ${TARGET_ARCH} from GitHub..."

DOWNLOAD_URL=$(
  curl -s https://api.github.com/repos/${REPO}/releases \
  | grep "browser_download_url" \
  | grep "linux-${TARGET_ARCH}.tar.gz" \
  | sed -E 's/.*"browser_download_url": "([^"]+)".*/\1/' \
  | head -n 1
)

if [ -z "$DOWNLOAD_URL" ]; then
  echo "❌ Could not find a release asset matching linux-${TARGET_ARCH}.tar.gz on GitHub."
  exit 1
fi

SIG_URL="${DOWNLOAD_URL}.minisig"

echo "=> Downloading: ${DOWNLOAD_URL}"
curl -sSL "$DOWNLOAD_URL" -o "${TMP_DIR}/archive.tar.gz"

echo "=> Downloading signature: ${SIG_URL}"
curl -sSL "$SIG_URL" -o "${TMP_DIR}/archive.minisig"

# UIL Minisign public key (from uil-release.pub)
UIL_PUBKEY="untrusted comment: minisign public key 55B82536E176CFEA
RWTqz3bhNiW4VV3XjzryJAU9CvAcFVpSe4+f9pc19R95zVCQaxJK0fV3"

echo "$UIL_PUBKEY" > "${TMP_DIR}/uil.pub"

echo "=> Verifying Minisign signature..."
if ! minisign -V -p "${TMP_DIR}/uil.pub" \
              -m "${TMP_DIR}/archive.tar.gz" \
              -x "${TMP_DIR}/archive.minisig"; then
  echo "❌ Signature verification failed. Aborting."
  rm -rf "${TMP_DIR}"
  exit 1
fi

echo "=> Signature verified. Extracting and installing..."
tar -xzf "${TMP_DIR}/archive.tar.gz" -C "${TMP_DIR}"

sudo cp "${TMP_DIR}/uild" /usr/local/bin/uild
sudo chmod 0755 /usr/local/bin/uild

# (your existing systemd + config setup goes here)

rm -rf "${TMP_DIR}"
echo "✅ UIL-X Hardware Governance Daemon (uild) installed successfully."
