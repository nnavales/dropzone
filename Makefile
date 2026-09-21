BINARY_NAME  := dropzone
MAIN_PACKAGE := ./cmd/main.go
BUILD_DIR    := ./.local/bin

.PHONY: help test test-race vet build run clean 

## help: Show this help menu with available commands.
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	-@rg --no-filename '^##' $(MAKEFILE_LIST) | awk '{sub(/^## /, ""); match($$0, /:/); printf "  \033[36m%-12s\033[0m %s\n", substr($$0, 1, RSTART-1), substr($$0, RSTART+1)}'

## test: Run all standard unit tests.
test:
	go test -v ./...

## test-race: Run unit tests with data race detection enabled.
test-race:
	go test -v -race ./...

## vet: Run Go static analysis checks.
vet:
	go vet ./...

## build: Compile the Go binary into the local/bin directory.
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

## run: Compile and immediately run the dropzone binary (local config).
run: build
	DROPZONE_CONFIG=.local/cfg/config.yml $(BUILD_DIR)/$(BINARY_NAME)

## clean: Clear build artifacts and reset the Go test cache.
clean:
	rm -rf $(BUILD_DIR)
	go clean -testcache
