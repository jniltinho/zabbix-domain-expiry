BINARY      := check_domain
CMD         := .
BUILD_DIR   := build
VERSION     ?= 2.0.0
LDFLAGS     := -s -w -X zabbix-domain-expiry/internal/output.Version=$(VERSION)
UPX         ?= upx
UPX_FLAGS   := --best --lzma

define compress
	@if command -v $(UPX) >/dev/null 2>&1; then \
		$(UPX) $(UPX_FLAGS) $(1); \
	else \
		echo "WARNING: upx not found, skipping compression for $(1)"; \
	fi
endef

.PHONY: all build build-nocompress clean test install run help

all: build

build: clean
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMD)
	$(call compress,$(BUILD_DIR)/$(BINARY))
	@echo "Built $(BUILD_DIR)/$(BINARY)"

build-nocompress:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMD)
	@echo "Built $(BUILD_DIR)/$(BINARY) (without UPX)"

build-linux-amd64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64 $(CMD)
	$(call compress,$(BUILD_DIR)/$(BINARY)-linux-amd64)

build-linux-arm64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-arm64 $(CMD)
	$(call compress,$(BUILD_DIR)/$(BINARY)-linux-arm64)

build-all: build-linux-amd64 build-linux-arm64

test:
	go test ./...

run: build
	./$(BUILD_DIR)/$(BINARY) $(ARGS)

install: build
	install -m 755 $(BUILD_DIR)/$(BINARY) /usr/lib/zabbix/externalscripts/$(BINARY)
	@echo "Installed to /usr/lib/zabbix/externalscripts/$(BINARY)"

clean:
	rm -f $(BUILD_DIR)/$(BINARY)

help:
	@echo "Targets:"
	@echo "  build              Build static binary compressed with UPX"
	@echo "  build-nocompress   Build static binary without UPX"
	@echo "  build-linux-amd64  Cross-compile for linux/amd64 (UPX)"
	@echo "  build-linux-arm64  Cross-compile for linux/arm64 (UPX)"
	@echo "  build-all          Build for all Linux targets (UPX)"
	@echo "  test               Run unit tests"
	@echo "  install            Install binary to Zabbix externalscripts dir"
	@echo "  run                Build and run (use ARGS='-d example.com')"
	@echo "  clean              Remove build artifacts"
