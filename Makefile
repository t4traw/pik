.PHONY: dev build install clean bindings release release-build release-tar release-sha

BIN_DIR  ?= $(HOME)/.local/bin
APP      := build/bin/pik.app
BIN      := $(APP)/Contents/MacOS/pik
DIST_DIR := dist

# Version: derive from the current git tag if on one, else "dev".
VERSION  ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "dev")
TARBALL  := $(DIST_DIR)/pik-$(VERSION)-darwin-universal.tar.gz

# Resolve `wails` from GOBIN if not already on PATH.
WAILS    := $(shell command -v wails || echo $(shell go env GOPATH)/bin/wails)

# ---- Dev / local install ----

dev:
	PIK_NO_RELAUNCH=1 $(WAILS) dev

build:
	$(WAILS) build

# Shim `pik` on PATH that exec's the signed binary inside the .app.
# Bare-copying the Mach-O breaks codesign (you'd see `zsh: killed`).
install: build
	@install -d $(BIN_DIR)
	@printf '#!/bin/sh\nexec "$(CURDIR)/$(BIN)" "$$@"\n' > $(BIN_DIR)/pik
	@chmod +x $(BIN_DIR)/pik
	@echo "installed → $(BIN_DIR)/pik (shim → $(CURDIR)/$(BIN))"

bindings:
	$(WAILS) generate module

# ---- Release ----

# Build a universal (arm64 + amd64) .app bundle.
release-build:
	$(WAILS) build -platform darwin/universal -trimpath -clean

# Tarball the .app so it can be uploaded as a GitHub Release asset.
# The archive unpacks to a single `pik.app/` directory.
release-tar: release-build
	@mkdir -p $(DIST_DIR)
	@rm -f $(TARBALL)
	tar -czf $(TARBALL) -C build/bin pik.app
	@echo "archive → $(TARBALL)"
	@shasum -a 256 $(TARBALL)

# Print the SHA256 — paste this into the Homebrew Cask formula.
release-sha: release-tar
	@shasum -a 256 $(TARBALL) | awk '{print $$1}'

# Full release pipeline (used locally before pushing a tag).
release: release-tar
	@echo "---"
	@echo "Upload $(TARBALL) to: https://github.com/t4traw/pik/releases/new"
	@echo "Then update the Homebrew cask with the sha256 above."

# ---- Housekeeping ----

clean:
	rm -rf build dist frontend/dist frontend/wailsjs
