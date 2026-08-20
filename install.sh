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

# Require minisign and sha3sum
if ! command -v minisign >/dev/null 2>&1; then
  echo "❌ minisign not found. Please install minisign first."
  exit 1
fi

if ! command -v sha3sum >/dev/null 2>&1; then
  echo "❌ sha3sum not found. Please install sha3sum (e.g., from sha3sum or coreutils with SHA3 support)."
  exit 1
fi

echo "=> Resolving latest pre-compiled release binary for ${TARGET_ARCH} from GitHub..."

# Use /releases/latest to always get newest release
LATEST_JSON="$(
  curl -s https://api.github.com/repos/${REPO}/releases/latest
)"

DOWNLOAD_URL="$(
  echo "$LATEST_JSON" \
  | grep "browser_download_url" \
  | grep "linux-${TARGET_ARCH}.tar.gz" \
  | sed -E 's/.*"browser_download_url": "([^"]+)".*/\1/' \
  | head -n 1
)"

if [ -z "$DOWNLOAD_URL" ]; then
  echo "❌ Could not find a release asset matching linux-${TARGET_ARCH}.tar.gz on GitHub."
  exit 1
fi

SIG_URL="${DOWNLOAD_URL}.minisig"
SHA3_URL="${DOWNLOAD_URL}.sha3"

echo "=> Downloading: ${DOWNLOAD_URL}"
curl -sSL "$DOWNLOAD_URL" -o "${TMP_DIR}/archive.tar.gz"

echo "=> Downloading signature: ${SIG_URL}"
curl -sSL "$SIG_URL" -o "${TMP_DIR}/archive.minisig"

echo "=> Downloading SHA3-256 checksum: ${SHA3_URL}"
curl -sSL "$SHA3_URL" -o "${TMP_DIR}/archive.tar.gz.sha3"

# UIL Minisign public key (from uil-release.key.pub)
cat > "${TMP_DIR}/uil.pub" <<EOF
untrusted comment: minisign public key 2AA63A801C42F2FD
RWT98kIcgDqmKjZOGcy6zrD3SV9ckm37CxFudz7Rkef9aAOXGjsk6i4u
EOF

echo "=> Verifying Minisign signature..."
if ! minisign -V -p "${TMP_DIR}/uil.pub" \
              -m "${TMP_DIR}/archive.tar.gz" \
              -x "${TMP_DIR}/archive.minisig"; then
  echo "❌ Signature verification failed. Aborting."
  rm -rf "${TMP_DIR}"
  exit 1
fi

echo "=> Verifying SHA3-256 checksum..."
(
  cd "${TMP_DIR}"

  # Detect GNU/Coreutils SHA3 (supports -c)
  if sha3sum --help 2>&1 | grep -q "Usage: sha3sum"; then
      # GNU SHA3
      if ! sha3sum -c "archive.tar.gz.sha3"; then
          echo "❌ SHA3-256 checksum verification failed. Aborting."
          exit 1
      fi
  else
      # Rust or BusyBox SHA3 (no -c support)
      EXPECTED="$(awk '{print $1}' archive.tar.gz.sha3)"
      ACTUAL="$(sha3sum archive.tar.gz | awk '{print $1}')"

      if [[ "$EXPECTED" != "$ACTUAL" ]]; then
          echo "❌ SHA3-256 checksum mismatch. Aborting."
          exit 1
      fi
  fi
)

echo "=> Signature and checksum verified. Extracting and installing..."
tar -xzf "${TMP_DIR}/archive.tar.gz" -C "${TMP_DIR}"

sudo cp "${TMP_DIR}/uild" /usr/local/bin/uild
sudo chmod 0755 /usr/local/bin/uild

rm -rf "${TMP_DIR}"
echo "✅ UIL-X Hardware Governance Daemon (uild) installed successfully."

echo "=> Ensuring uild system user exists..."
if ! id "uild" >/dev/null 2>&1; then
    sudo useradd --system --no-create-home --shell /usr/sbin/nologin uild
    echo "   ✔ Created system user 'uild'"
else
    echo "   ✔ System user 'uild' already exists"
fi

echo "=> Installing hardened systemd service..."
sudo tee /etc/systemd/system/uild.service >/dev/null <<'EOF'
[Unit]
Description=UIL-X Hardware Governance Daemon
After=network.target

[Service]
Type=simple

# Run as a dedicated non-root user
User=uild
Group=uild

# Create secure runtime directory for IPC socket
RuntimeDirectory=uild
RuntimeDirectoryMode=0750

# Working directory (important for relative paths)
WorkingDirectory=/var/lib/uild

# Exec command
ExecStart=/usr/local/bin/uild run

# Restart policy
Restart=on-failure
RestartSec=3

# Capability drop (extra hardening)
NoNewPrivileges=true
ProtectSystem=full
ProtectHome=true
PrivateTmp=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

[Install]
WantedBy=multi-user.target
EOF

echo "=> Reloading systemd..."
sudo systemctl daemon-reload

echo "=> Enabling UIL-X daemon..."
sudo systemctl enable uild

echo "=> Starting UIL-X daemon..."
sudo systemctl restart uild

echo "✅ UIL-X daemon installed and running under locked-down 'uild' user."
