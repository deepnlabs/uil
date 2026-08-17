### `ros2/README.md`

```markdown
# 🤖 UIL-X ROS2 Hardware Safety & Emergency Stop Bridge

This directory contains the native ROS2 integration package for **UIL-X (`uild`)**, providing real-time hardware governance, thermal/power interlock monitoring, and automated emergency stop ($E\text{-}Stop$) actuation across ROS2-managed autonomous systems.

---

## 📐 Architecture Overview

`uild` operates as a low-latency background daemon reading kernel sysfs telemetry, NPU hardware execution metrics, and physical GPIO state. When an interlock breach is detected (e.g., thermal limit overrun or power anomaly), `uild` broadcasts the breach payload over an IPC Unix Domain Socket (`/var/run/uild.sock`).

The ROS2 bridge node (`ros2_uil_estop.py`) listens to this socket and immediately publishes safety overrides to ROS2 topics:


```

┌────────────────────────────────┐                 /var/run/uil.sock                ┌───────────────────────────────────┐
│ UIL-X Daemon (`uild`)          │ ───────────────────────────────────────────────> │ ROS2 Safety Bridge Node           │
│ • Kernel sysfs thermal monitoring│  {"interlock_id": "linux-sysfs-v1",            │ (`ros2_uil_estop.py`)             │
│ • Sub-millisecond interlocks   │   "breach": true, "cpu_temp_celsius": 85.2}     └─────────────────┬─────────────────┘
└────────────────────────────────┘                                                                    │
▼
┌───────────────────────────────────┐
│ ROS2 Execution Bus                │
│ • /safety/emergency_stop (Bool)   │
│ • /cmd_vel (Twist Zero Vectors)   │
└───────────────────────────────────┘

```

---

## ⚡ Topics Published

| Topic | Message Type | Description |
| :--- | :--- | :--- |
| `/safety/emergency_stop` | `std_msgs/msg/Bool` | Publishes `True` instantly upon any UIL-X safety interlock breach. |
| `/cmd_vel` | `geometry_msgs/msg/Twist` | Forces zero linear/angular velocity vectors to halt motor controllers immediately. |

---

## 🚀 Quickstart & Usage

### 1. Prerequisites
* **ROS2 Install:** Humble Hawksbill or newer (`rclpy`, `std_msgs`, `geometry_msgs`).
* **UIL-X Daemon:** `uild` running locally on the host node (`./bin/uild run`).

### 2. Manual Execution
Run the ROS2 bridge node inside your active ROS2 environment:

```bash
source /opt/ros/humble/setup.bash
python3 ros2/ros2_uil_estop.py

```

### 3. Adding to a `colcon` Workspace

To include `ros2_uil_estop` in your robot's launch stack, symlink or copy this directory into your ROS2 workspace `src/` folder:

```bash
cd ~/ros2_ws/src
ln -s ~/go/deepn-uil/ros2 uil_safety_bridge
cd ~/ros2_ws
colcon build --packages-select uil_safety_bridge
source install/setup.bash

```

---

## 🧪 Testing Safety Interlocks

To test E-Stop actuation without overheating physical hardware:

1. Launch the `uild` daemon on `deepn-node-2` (or your target ARM/x86 node):
```bash
./bin/uild run

```


2. Start the ROS2 safety node in a separate terminal:
```bash
ros2 run uil_safety_bridge ros2_uil_estop.py

```


3. Echo the emergency stop topic to observe breach events:
```bash
ros2 topic echo /safety/emergency_stop

```



---

## 🛡️ Intellectual Property & Licensing

* **License:** GNU Affero General Public License v3.0 ([AGPL-3.0](https://www.google.com/search?q=../LICENSE)). Commercial enterprise licenses and closed-source driver exceptions are available through DeePN Labs.
* **Patent Notice:** Protected under pending United States Intellectual Property filings (**U.S. Patent Application No. 19/753,918**).

---

*Fiscally hosted via [Open Source Collective](https://www.google.com/search?q=https://opencollective.com/uil-x).*

```

```
