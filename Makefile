.PHONY: build clean install uninstall start stop status run tray

VERSION := 1.0.0
BINARY := go-pc-rem.exe
LDFLAGS := -ldflags="-s -w -H windowsgui -X main.Version=$(VERSION)"

# Build the combined binary
build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/go-pc-rem

clean:
	del /f $(BINARY) 2>nul || true

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
	go run ./cmd/go-pc-rem run

tidy:
	go mod tidy

test:
	go test ./...
