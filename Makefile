BINARY_NAME  := dropzone
MAIN_PACKAGE := ./cmd
BUILD_DIR    := ./.local/bin
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo 0.0.1)
ARGS         ?= --dev

.PHONY: help build run test test-race vet clean version release

## help: Show this help menu with available commands.
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	-@rg --no-filename '^##' $(MAKEFILE_LIST) | awk '{sub(/^## /, ""); match($$0, /:/); printf "  \033[36m%-12s\033[0m %s\n", substr($$0, 1, RSTART-1), substr($$0, RSTART+1)}'

## build: Compile the Go binary into the local/bin directory.
build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

## run: Compile and immediately run the dropzone binary (local config). Override with ARGS, e.g. make run ARGS="init --dev".
run: build
	$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## test: Run all standard unit tests.
test:
	go test -v ./...

## test-race: Run unit tests with data race detection enabled.
test-race:
	go test -v -race ./...

## vet: Run Go static analysis checks.
vet:
	go vet ./...

## clean: Clear build artifacts and reset the Go test cache.
clean:
	rm -rf $(BUILD_DIR)
	go clean -testcache

## version: Show the current release tag.
version:
	@last="$$(git tag --list 'v*' --sort=-v:refname | head -n1)"; \
	if [ -n "$$last" ]; then echo "current version: $$last"; else echo "no release tags yet"; fi

## release: Tag and push a release, e.g. make release version=v0.2.0.
release:
	@version=$(version); \
	if [ -z "$$version" ]; then \
		echo "usage: make release version=v0.2.0"; \
		last="$$(git tag --list 'v*' --sort=-v:refname | head -n1)"; \
		if [ -n "$$last" ]; then \
			next="$$(echo "$$last" | awk -F. '{print $$1"."$$2"."($$3+1)}')"; \
			echo "current version: $$last"; \
			echo "suggested next:  $$next"; \
		else \
			echo "no release tags yet - suggested first: v0.1.0"; \
		fi; \
		exit 1; \
	fi; \
	case "$$version" in v*) ;; *) echo "version must start with 'v' (got: $$version)" >&2; exit 1;; esac; \
	if ! git diff --quiet || ! git diff --cached --quiet; then echo "working tree is dirty: commit changes first" >&2; exit 1; fi; \
	if git rev-parse "$$version" >/dev/null 2>&1; then echo "tag $$version already exists" >&2; exit 1; fi; \
	git push origin master; \
	git tag "$$version"; \
	git push origin "$$version"
