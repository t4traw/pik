.PHONY: dev build install clean bindings

BIN_DIR ?= $(HOME)/.local/bin
APP     := build/bin/pik.app
BIN     := $(APP)/Contents/MacOS/pik

# Resolve `wails` from GOBIN (go install's target) if not already on PATH.
WAILS   := $(shell command -v wails || echo $(shell go env GOPATH)/bin/wails)

# Interactive development with hot reload on both frontend & Go.
dev:
	$(WAILS) dev

# Production build (frontend bundled, self-signed .app).
build:
	$(WAILS) build

# Install a shim at $(BIN_DIR)/pik that execs the Mach-O binary from inside
# the .app bundle. Running the binary bare outside the bundle breaks macOS
# code signing (you'd get `zsh: killed`), so the shim keeps the bundle
# context while still giving you a `pik` command on PATH.
install: build
	@install -d $(BIN_DIR)
	@printf '#!/bin/sh\nexec "$(CURDIR)/$(BIN)" "$$@"\n' > $(BIN_DIR)/pik
	@chmod +x $(BIN_DIR)/pik
	@echo "installed → $(BIN_DIR)/pik (shim → $(CURDIR)/$(BIN))"

# Regenerate frontend/wailsjs/* bindings without rebuilding.
bindings:
	$(WAILS) generate module

clean:
	rm -rf build frontend/dist frontend/wailsjs
