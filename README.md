# UIL‑X Hardware Governance Daemon

UIL‑X is a secure, cross‑platform hardware governance daemon designed for distributed mesh environments, embedded devices, and heterogeneous compute nodes. It provides:

- a hardened, non‑root system daemon (`uild`)
- a companion control‑plane CLI (`uilctl`)
- secure plugin execution
- Minisign‑verified releases
- SHA3‑256 integrity verification
- automatic updates
- multi‑architecture builds (amd64, arm64, armv6)
- systemd sandboxing and privilege isolation

UIL‑X is built for environments where hardware control, telemetry, and distributed coordination must be **secure by default**.

---

## Features

### ✔ Secure Daemon (`uild`)
- Runs under a locked‑down system user (`uild`)
- No root privileges required after install
- Strict systemd sandboxing
- No kernel module access
- No device access
- No privilege escalation
- Restricted network families

### ✔ Control‑Plane CLI (`uilctl`)
- `uilctl update` — secure auto‑update
- `uilctl status` — daemon health
- Future: mesh commands, plugin commands, thermal controls

### ✔ Secure Release Pipeline
- Cross‑compiled binaries for:
  - `linux-amd64`
  - `linux-arm64`
  - `linux-armv6` (Pi Zero W)
- Minisign signatures for all release artifacts
- SHA3‑256 checksums for integrity verification
- GitHub Releases auto‑publishing

### ✔ Secure Installer
- Auto‑detects latest release
- Verifies Minisign signature
- Verifies SHA3‑256 checksum
- Installs hardened systemd service
- Creates `uild` system user if missing
- Starts daemon safely

### ✔ Plugin System
- Dynamic `.so` plugins
- Loaded under non‑root sandbox
- Safe for distributed environments

UIL v0.9 Rewrite (In Progress)

The current version of UIL lives in the root directory.

The next major version (v0.9) is being developed in /v0.9/ as a clean rewrite with a minimal core and plugin ecosystem.

v0.8.1.8 will remain available for legacy users.

---

## Installation

To install the latest verified release:

```bash
curl -sSL https://raw.githubusercontent.com/deepnlabs/uil/main/install.sh | bash



