# GhostScan — Makefile
# Targets: build, release, checksums, clean, install

PROJECT   := ghostscan
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
LDFLAGS   := -ldflags="-s -w -X main.version=$(VERSION)"
BUILD_DIR := dist

# Platforms to build for
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: all build release checksums clean install

all: build

# ── Build for the local OS/arch ─────────────────────────────────
build:
	@echo "==> Building $(PROJECT) $(VERSION) for $(shell go env GOOS)/$(shell go env GOARCH)"
	go build $(LDFLAGS) -o $(PROJECT)$(shell go env GOEXE) .

# ── Cross-platform release archives ─────────────────────────────
release: clean
	@mkdir -p $(BUILD_DIR)
	@set -e; \
	for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		ext=$$( [ "$$os" = "windows" ] && echo ".exe" || echo "" ); \
		archive="$(PROJECT)_$(VERSION)_$${os}_$${arch}.zip"; \
		echo "==> Building $$archive"; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o "$(BUILD_DIR)/$(PROJECT)$$ext" .; \
		cd $(BUILD_DIR); \
		if [ "$$os" != "windows" ]; then chmod +x "$(PROJECT)$$ext"; fi; \
		zip "$$archive" "$(PROJECT)$$ext"; \
		rm "$(PROJECT)$$ext"; \
		cd ..; \
	done
	@echo ""
	@ls -lh $(BUILD_DIR)/

# ── SHA256 checksums ────────────────────────────────────────────
checksums: release
	@echo "==> Generating SHA256 checksums"
	cd $(BUILD_DIR) && sha256sum *.zip > "$(PROJECT)_$(VERSION)_checksums.txt"
	@cat $(BUILD_DIR)/$(PROJECT)_$(VERSION)_checksums.txt

# ── Clean ───────────────────────────────────────────────────────
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f $(PROJECT) $(PROJECT).exe

# ── Install to GOPATH/bin ──────────────────────────────────────
install:
	go install $(LDFLAGS) .

# ── Test ────────────────────────────────────────────────────────
test:
	go test ./...
	go vet ./...