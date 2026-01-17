.PHONY: build clean install uninstall start stop status run tray installer
.PHONY: build-macos-intel build-macos-arm build-macos-universal
.PHONY: dmg-intel dmg-arm dmg-universal dmg-all clean-macos

VERSION := 0.0.23-beta
BINARY_DIR := bin
BINARY := $(BINARY_DIR)/Cntrl.exe
LDFLAGS := -ldflags="-s -w -H windowsgui -X main.Version=$(VERSION)"

# ============================================================================
# Windows Builds
# ============================================================================

# Build the combined binary (Windows)
build:
ifeq ($(OS),Windows_NT)
	if not exist $(BINARY_DIR) mkdir $(BINARY_DIR)
	go build $(LDFLAGS) -o $(BINARY) ./cmd/cntrl
else
	@echo "Windows build target. Use 'make build-macos-*' on macOS."
endif

# Build the installer (requires Inno Setup)
installer: build
	"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" Cntrl.iss

# ============================================================================
# macOS Builds
# ============================================================================

# Build Intel (amd64) binary
build-macos-intel:
	./scripts/build-macos.sh $(VERSION) intel

# Build Apple Silicon (arm64) binary
build-macos-arm:
	./scripts/build-macos.sh $(VERSION) arm

# Build Universal binary (both architectures)
build-macos-universal:
	./scripts/build-macos-universal.sh $(VERSION)

# Create Intel DMG
dmg-intel: build-macos-intel
	./scripts/create-dmg.sh $(VERSION) intel

# Create Apple Silicon DMG
dmg-arm: build-macos-arm
	./scripts/create-dmg.sh $(VERSION) arm

# Create Universal DMG
dmg-universal: build-macos-universal
	./scripts/create-dmg.sh $(VERSION) universal

# Build all DMGs (Intel, Apple Silicon, and Universal)
dmg-all: build-macos-universal
	./scripts/create-dmg.sh $(VERSION) intel
	./scripts/create-dmg.sh $(VERSION) arm
	./scripts/create-dmg.sh $(VERSION) universal

# Clean macOS build artifacts
clean-macos:
	rm -rf macos/Cntrl-Intel.app macos/Cntrl-AppleSilicon.app
	rm -rf dist/*.dmg

# ============================================================================
# Common
# ============================================================================

clean:
ifeq ($(OS),Windows_NT)
	if exist bin rmdir /s /q bin
	if exist dist rmdir /s /q dist
else
	rm -rf bin dist
	rm -rf macos/Cntrl-Intel.app macos/Cntrl-AppleSilicon.app
endif

install: build
	$(BINARY) install

uninstall:
	$(BINARY) uninstall

start:
	$(BINARY) start

stop:
	$(BINARY) stop

status:
	$(BINARY) status

run: build
	$(BINARY) run

# Launch tray mode
tray: build
	start $(BINARY)

# Development helpers
dev:
	go run ./cmd/cntrl run

tidy:
	go mod tidy

test:
	go test ./...
