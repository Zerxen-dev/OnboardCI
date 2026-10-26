.PHONY: all build test test-unit lint install assets clean check

BIN_NAME := onboardci
BUILD_CACHE_DIR := /root/build/Onboardci
CUR_DIR := $(shell pwd)

# Detect if current filesystem supports POSIX file locks needed by Go modules
LOCK_UNSUPPORTED := $(shell python3 -c "import fcntl; f = open('go.mod', 'r'); fcntl.flock(f, fcntl.LOCK_SH)" 2>/dev/null && echo 0 || echo 1)

# On Android/ARM64 39-bit VMA kernels, ThreadSanitizer (-race) is unsupported by hardware
ifeq ($(LOCK_UNSUPPORTED),1)
RACE_FLAG :=
else
RACE_FLAG := -race
endif

all: build

sync-build-cache:
ifeq ($(LOCK_UNSUPPORTED),1)
	@mkdir -p $(BUILD_CACHE_DIR)
	@cp -r $(CUR_DIR)/internal $(BUILD_CACHE_DIR)/ 2>/dev/null || true
	@cp -r $(CUR_DIR)/cmd $(BUILD_CACHE_DIR)/ 2>/dev/null || true
	@cp -r $(CUR_DIR)/testdata $(BUILD_CACHE_DIR)/ 2>/dev/null || true
	@cp $(CUR_DIR)/go.mod $(CUR_DIR)/go.sum $(BUILD_CACHE_DIR)/ 2>/dev/null || true
endif

build: sync-build-cache
ifeq ($(LOCK_UNSUPPORTED),1)
	@cd $(BUILD_CACHE_DIR) && go build -o $(BUILD_CACHE_DIR)/bin/$(BIN_NAME) ./cmd/onboardci
	@mkdir -p $(CUR_DIR)/bin
	@cp $(BUILD_CACHE_DIR)/bin/$(BIN_NAME) $(CUR_DIR)/bin/$(BIN_NAME)
else
	@mkdir -p bin
	go build -o bin/$(BIN_NAME) ./cmd/onboardci
endif
	@echo "Build successful: bin/$(BIN_NAME)"

test-unit: sync-build-cache
ifeq ($(LOCK_UNSUPPORTED),1)
	cd $(BUILD_CACHE_DIR) && go test -v ./internal/...
else
	go test -v ./internal/...
endif

test: sync-build-cache
ifeq ($(LOCK_UNSUPPORTED),1)
	cd $(BUILD_CACHE_DIR) && go test -v $(RACE_FLAG) ./internal/...
	cd $(BUILD_CACHE_DIR) && go test -v -tags integration ./internal/runner
else
	go test -v $(RACE_FLAG) ./...
	go test -v -tags integration ./internal/runner
endif

lint: sync-build-cache
ifeq ($(LOCK_UNSUPPORTED),1)
	cd $(BUILD_CACHE_DIR) && go vet ./internal/... ./cmd/...
else
	go vet ./...
endif

install: build
	cp bin/$(BIN_NAME) /usr/local/bin/$(BIN_NAME) || cp bin/$(BIN_NAME) $(HOME)/go/bin/$(BIN_NAME)
	@echo "Installed $(BIN_NAME) successfully."

assets:
	@python3 scripts/validate_assets.py

check: lint test assets

clean:
	rm -rf bin/ $(BUILD_CACHE_DIR)/bin/
