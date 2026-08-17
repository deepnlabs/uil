# ⚡ UIL-X (`uild`): Universal Interface Layer & Hardware Governance Daemon

[![Release](https://img.shields.io/badge/release-v0.8.0--alpha-blue.svg)](https://github.com/deepnlabs/uil/releases)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE)
[![Open Source Collective](https://img.shields.io/badge/Open_Source_Collective-fiscally_hosted-orange.svg)](https://opencollective.com/uil)

**UIL-X (`uild`)** is a lightweight, low-latency hardware governance daemon and execution safety framework designed for edge AI accelerators, robotics stacks, and heterogeneous Linux nodes. It continuously evaluates physical thermal/power interlocks, computes post-quantum SHA3 payload commitments ($UPFo$), and streams real-time state alerts across a peer-to-peer (P2P) UDP mesh.

---

## 🏛️ System Architecture

                     ┌──────────────────────────────┐
                                │   ROS2 Autonomous Systems    │
                                │   • /safety/emergency_stop   │
                                │   • /cmd_vel (Twist Overrides)│
                                └──────────────▲───────────────┘
                                               │ IPC Unix Socket (/tmp/uild.sock)
┌─────────────────────────┐          ┌─────────────┴───────────────┐          ┌─────────────────────────┐│  deepn-node-1 (x86_64)  │ <──────> │   UIL-X Daemon (uild)     │ <──────> │  node-pi0-1 (ARMv6)     ││  • P2P Mesh Gossip      │ UDP 9090 │   • sysfs Thermal Reader    │ UDP 9090 │  • Edge Sensors         ││  • Substrate Routing    │          │   • Post-Quantum SHA3 $UPFo$│          │  • Low-Power Governance │└─────────────────────────┘          └─────────────────────────────┘          └─────────────────────────┘
---

## ⚡ Key Features in v0.8.0 Alpha

* **Post-Quantum Cryptographic Commitments:** Standardized SHA3-256 state hashing (`pkg/crypto`) for verifiable audit trails.
* **P2P Mesh Gossip Network (`pkg/mesh`):** Zero-configuration UDP discovery and sub-second interlock event propagation across cluster nodes.
* **Dynamic `.so` Substrate Drivers (`pkg/substrate`):** Load proprietary thermal and physical safety interlocks at runtime via constructor function symbols (`NewDriver()`).
* **ROS2 Emergency Stop Bridge (`ros2/`):** Unix socket IPC streaming (`/tmp/uild.sock`) triggering instant ROS2 `$E\text{-}Stop$` publishes and zero-velocity overrides.
* **Kernel Thermal Telemetry (`pkg/substrate/gpio_linux.go`):** Reads hardware CPU thermal zones directly from `/sys/class/thermal/`.
* **Multi-Architecture Support:** Pre-compiled static release binaries for `amd64`, `arm64`, and `armv6` (Raspberry Pi Zero W).

---

## 🚀 5-Minute Quickstart

### 1. Automated Installation (Linux `amd64` / `arm64` / `armv6`)
Run the one-line installer on your Linux host:

```bash
curl -sSL [https://raw.githubusercontent.com/deepnlabs/uil/main/install.sh](https://raw.githubusercontent.com/deepnlabs/uil/main/install.sh) | bash
Or install locally from source:Bashgit clone [https://github.com/deepnlabs/uil.git](https://github.com/deepnlabs/uil.git)
cd uil
make all
sudo ./install.sh
2. Service Managementuild runs as an auto-restarting systemd background daemon:Bash# Check service status
sudo systemctl status uild

# Stream live governance & interlock logs
sudo journalctl -u uild -f
3. Developer CLI UsageBash# Manual runtime execution
uild run

# Status check
uild status
🤖 ROS2 IntegrationTo connect your ROS2 robot controllers to UIL-X safety interlocks:Bash# Run the ROS2 IPC Emergency Stop Bridge Node
python3 ros2/ros2_uil_estop.py
🛡️ Intellectual Property & LicenseLicense: GNU Affero General Public License (AGPL-3.0).Commercial Exception: Closed-source commercial driver exception licenses are available through DeePN Labs.Patent Notice: Protected under pending United States Intellectual Property filings (U.S. Patent Application No. 19/753,918).🤝 Community & SupportUIL-X is fiscally hosted through Open Source Collective. Help support independent hardware governance infrastructure via Open Collective.

