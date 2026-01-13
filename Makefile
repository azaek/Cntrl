.PHONY: build clean install uninstall start stop status run tray installer

VERSION := 0.0.23-beta
BINARY_DIR := bin
BINARY := $(BINARY_DIR)\Cntrl.exe
LDFLAGS := -ldflags="-s -w -H windowsgui -X main.Version=$(VERSION)"

# Build the combined binary
build:
	if not exist $(BINARY_DIR) mkdir $(BINARY_DIR)
	go build $(LDFLAGS) -o $(BINARY) ./cmd/cntrl

# Build the installer (requires Inno Setup)
installer: build
	"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" Cntrl.iss

clean:
	if exist bin rmdir /s /q bin
	if exist dist rmdir /s /q dist

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
