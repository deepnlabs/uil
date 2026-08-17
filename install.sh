#!/usr/bin/env bash
set -e

REPO="deepnlabs/uil"

echo "--------------------------------------------------------"
echo "  Installing UIL-X Hardware Governance Daemon (uild)   "
echo "--------------------------------------------------------"

# 1. Detect System Architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)        TARGET_ARCH="amd64" ;;
    aarch64|arm64) TARGET_ARCH="arm64" ;;
    armv6l|armv7l) TARGET_ARCH="armv6" ;;
    *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "=> Architecture detected: $TARGET_ARCH"

# 2. Prepare System Directories
sudo mkdir -p /etc/uil
sudo mkdir -p /var/lib/uild/plugins
sudo mkdir -p /usr/local/bin

# 3. Obtain Binary
if [ -f "./bin/uild" ]; then
    echo "=> Installing locally compiled binary from ./bin/uild..."
    sudo cp ./bin/uild /usr/local/bin/uild
elif [ -f "./Makefile" ]; then
    echo "=> Compiling from local source..."
    make all
    sudo cp ./bin/uild /usr/local/bin/uild
else
    echo "=> Resolving pre-compiled release binary for ${TARGET_ARCH} from GitHub..."
    TMP_DIR=$(mktemp -d)
    
    # Query GitHub API for the direct asset URL matching our architecture
    DOWNLOAD_URL=$(curl -s https://api.github.com/repos/${REPO}/releases | \
      grep "browser_download_url" | \
      grep -E "uild-v[0-9]+\.[0-9]+\.[0-9]+.*-linux-${TARGET_ARCH}\.tar\.gz"
      head -n 1 | \
      cut -d '"' -f 4)

    if [ -z "$DOWNLOAD_URL" ]; then
        echo "❌ Could not find a release asset matching linux-${TARGET_ARCH}.tar.gz on GitHub."
        exit 1
    fi

    echo "=> Downloading: ${DOWNLOAD_URL}"
    curl -sSL "$DOWNLOAD_URL" -o "${TMP_DIR}/archive.tar.gz"
    tar -xzf "${TMP_DIR}/archive.tar.gz" -C "${TMP_DIR}"
    sudo cp "${TMP_DIR}/uild" /usr/local/bin/uild
    rm -rf "${TMP_DIR}"
fi

sudo chmod +x /usr/local/bin/uild

# 4. Install Default Configuration
if [ ! -f "/etc/uil/config.json" ]; then
    echo "=> Writing default configuration to /etc/uil/config.json..."
    sudo tee /etc/uil/config.json > /dev/null << 'EOF'
{
  "node_id": "auto",
  "log_level": "info",
  "thermal_limit_celsius": 80.0,
  "ipc": {
    "enabled": true,
    "socket_path": "/tmp/uild.sock"
  },
  "mesh": {
    "enabled": true,
    "port": 9090
  }
}
EOF
fi

# 5. Install Systemd Service Unit
if [ -d "/etc/systemd/system" ]; then
    echo "=> Installing uild.service systemd unit..."
    sudo tee /etc/systemd/system/uild.service > /dev/null << 'EOF'
[Unit]
Description=UIL-X Hardware Governance Daemon (uild)
Documentation=https://github.com/deepnlabs/uil
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/uild run
Restart=always
RestartSec=3s
WorkingDirectory=/var/lib/uild

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    echo "=> Enabling and starting uild.service..."
    sudo systemctl enable --now uild.service
fi

echo ""
echo "✅ UIL-X Installation Complete on $(hostname)!"
echo "   Status: sudo systemctl status uild"
echo "   Logs:   sudo journalctl -u uild -f"
