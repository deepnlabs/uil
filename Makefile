# UIL-X Hardware Governance Daemon Build System
BINARY_NAME=uild
BUILD_DIR=bin
DIST_DIR=dist
MAIN_SRC=cmd/uild/main.go
UILCTL_SRC=cmd/uilctl/main.go
PLUGIN_SRC=plugins_src/custom_interlock/main.go
PLUGIN_OUT=plugins/custom_interlock.so

VERSION=v0.8.1.2-alpha
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(COMMIT) -s -w"

.PHONY: all clean build build-plugin linux-amd64 linux-arm64 linux-armv6 dist help

all: clean build build-plugin

## Build local host binary
build:
	@echo "==> Building local host binary ($(BINARY_NAME))..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_SRC)

## Build local dynamic .so plugin
build-plugin:
	@echo "==> Building local dynamic plugin..."
	@mkdir -p plugins
	CGO_ENABLED=1 go build -buildmode=plugin -o $(PLUGIN_OUT) $(PLUGIN_SRC)

## Cross-Compile Target: x86_64 / Linux AMD64 (e.g., deepn-node-1)
linux-amd64:
	@echo "==> Cross-compiling for linux/amd64..."
	@mkdir -p $(BUILD_DIR)/amd64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/amd64/$(BINARY_NAME) $(MAIN_SRC)

## Cross-Compile Target: 64-bit ARM / Linux ARM64 (e.g., deepn-node-2, NanoPi, Jetson)
linux-arm64:
	@echo "==> Cross-compiling for linux/arm64..."
	@mkdir -p $(BUILD_DIR)/arm64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/arm64/$(BINARY_NAME) $(MAIN_SRC)

## Cross-Compile Target: 32-bit ARM / Linux ARMv6 (e.g., Raspberry Pi Zero W)
linux-armv6:
	@echo "==> Cross-compiling for linux/armv6 (Pi Zero W)..."
	@mkdir -p $(BUILD_DIR)/armv6
	GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/armv6/$(BINARY_NAME) $(MAIN_SRC)

## Build all cross-compilation binaries and create release tarballs in dist/
dist: clean linux-amd64 linux-arm64 linux-armv6 uilctl-all
	@echo "==> Packaging release distribution tarballs in $(DIST_DIR)/..."
	@mkdir -p $(DIST_DIR)
	@tar -czf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR)/amd64 $(BINARY_NAME) uilctl
	@tar -czf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(BUILD_DIR)/arm64 $(BINARY_NAME) uilctl
	@tar -czf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-armv6.tar.gz -C $(BUILD_DIR)/armv6 $(BINARY_NAME) uilctl
	@echo "✅ Distribution release packages created in ./$(DIST_DIR):"
	@ls -lh $(DIST_DIR)

## Build uilctl for all architectures
uilctl-all: $(BUILD_DIR)/amd64/uilctl $(BUILD_DIR)/arm64/uilctl $(BUILD_DIR)/armv6/uilctl

$(BUILD_DIR)/amd64/uilctl: ./cmd/uilctl/*.go
	@mkdir -p $(BUILD_DIR)/amd64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o $(BUILD_DIR)/amd64/uilctl ./cmd/uilctl

$(BUILD_DIR)/arm64/uilctl: ./cmd/uilctl/*.go
	@mkdir -p $(BUILD_DIR)/arm64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -o $(BUILD_DIR)/arm64/uilctl ./cmd/uilctl

$(BUILD_DIR)/armv6/uilctl: ./cmd/uilctl/*.go
	@mkdir -p $(BUILD_DIR)/armv6
	GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build -ldflags "-s -w" -o $(BUILD_DIR)/armv6/uilctl ./cmd/uilctl

## Remove build artifacts and temporary files
clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)

help:
	@echo "UIL-X Build System Targets:"
	@echo "  make all           - Clean and build local native binary & plugin"
	@echo "  make linux-amd64   - Cross-compile static binary for x86_64"
	@echo "  make linux-arm64   - Cross-compile static binary for ARM64"
	@echo "  make linux-armv6   - Cross-compile static binary for Raspberry Pi Zero W"
	@echo "  make dist          - Build all architectures and generate release tarballs"
	@echo "  make clean         - Delete bin/, dist/, and plugin binaries"

runtime-clean:
	@echo "==> Cleaning runtime artifacts..."
	@sudo rm -f /tmp/uild.sock 2>/dev/null || true
	@sudo systemctl stop uild 2>/dev/null || true

