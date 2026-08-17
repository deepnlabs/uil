#!/usr/bin/env bash
set -e

# UIL-X Automated Installer Script (v0.8-alpha)
echo "--------------------------------------------------------"
echo "  Installing UIL-X Hardware Governance Daemon (uild)   "
echo "--------------------------------------------------------"

# 1. Detect System Architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)      TARGET_ARCH="amd64" ;;
    aarch64|arm64) TARGET_ARCH="arm64" ;;
    armv6l|armv7l) TARGET_ARCH="armv6" ;;
    *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "=> Architecture detected: $TARGET_ARCH"

# 2. Prepare Directories
sudo mkdir -p /etc/uild
sudo mkdir -p /var/lib/uild/plugins
sudo mkdir -p /usr/local/bin

# 3. Build/Copy Local Binary
if [ -f "./bin/uild" ]; then
    echo "=> Installing locally compiled binary..."
    sudo cp ./bin/uild /usr/local/bin/uild
else
    echo "=> Compiling binary..."
    make all
    sudo cp ./bin/uild /usr/local/bin/uild
fi

sudo chmod +x /usr/local/bin/uild

# 4. Install Default Configuration
if [ ! -f "/etc/uild/config.json" ]; then
    echo "=> Writing default configuration to /etc/uild/config.json..."
    sudo cp ./config/config.json /etc/uild/config.json
fi

# 5. Install Systemd Service
if [ -d "/etc/systemd/system" ]; then
    echo "=> Installing uild.service systemd unit..."
    sudo cp ./packaging/systemd/uild.service /etc/systemd/system/uild.service
    sudo systemctl daemon-reload
    echo "=> Enabling and starting uild.service..."
    sudo systemctl enable --now uild.service
fi

echo ""
echo "✅ UIL-X Installation Complete!"
echo "   Status: sudo systemctl status uild"
echo "   Logs:   sudo journalctl -u uild -f"

