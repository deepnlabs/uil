# DeepN Universal Interface Layer (UIL) Specification & Go SDK

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Protocol Version](https://img.shields.io/badge/Protocol-UIP--1.0--CDID-blue.svg)](#)

The **Universal Interface Layer (UIL)** is an open specification and lightweight Go SDK for structuring, validating, and dispatching heterogeneous computing payloads across edge nodes, NPU hardware accelerators, and distributed intelligence daemons.

## Overview

In modern distributed edge systems, telemetry, prompts, and inference queries travel across heterogeneous acceleration targets (e.g., AMD Radeon Vulkan, NVIDIA CUDA, Hailo-8 NPU, MemryX NPU, and AMD XDNA 2). UIL provides a standardized, hardware-agnostic envelope protocol (**UIP-v1**) to encapsulate workloads alongside cryptographic proof commitments (`PoCog` / `PoX`).
