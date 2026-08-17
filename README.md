# UIL-X (`uild`)

[![Open Collective](https://opencollective.com/uil-x/tiers/badge.svg)](https://opencollective.com/uil-x)
[![License](https://img.shields.io/badge/License-AGPL_v3.0-blue.svg)](LICENSE)
[![Patent](https://img.shields.io/badge/Patent-Pending-amber.svg)](#patent--intellectual-property)

**Real-time hardware governance runtime and sub-millisecond safety interlocks for edge computing and autonomous systems.**

> *"Put WE back in PoWEr."*

---

## ⚡ What is UIL-X?

As autonomous AI workloads and cyber-physical systems deploy into edge hardware, physical execution bounds, power limits, and thermal guardrails are increasingly trapped inside closed, proprietary platforms. 

**UIL-X** flips this dynamic by providing a lightweight, low-latency background daemon (`uild`) written in Go. It enforces strict execution safety, constraint monitoring ($CSS\text{-}X$ / $CaP\text{-}X$), and hardware interlocks locally at the edge—keeping physical computing safe, transparent, and developer-owned.

---

## 🚀 Key Features

* **Sub-Millisecond Interlocks:** Real-time interception and signal termination for hardware safety breaches.
* **Local Governance Daemon:** Operates independently at the edge without cloud dependencies or black-box lock-in.
* **Cryptographic Event Logging:** Standardized $UPFo$ proof headers for auditability and state verification.
* **Edge-Native Support:** Built for Linux environments (x86_64 / ARM64) with support for hardware monitoring (NPUs, SBCs, single-board controllers).

---

## 🏁 Quickstart (v0.8 Alpha Preview)

```bash
# Initialize local configuration
uild init

# Run the safety governance daemon in background
uild run --config /etc/uild/config.yml

# Check daemon health and active constraint loops
uild status
