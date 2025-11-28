SHELL := /bin/bash
export HOMEBREW_NO_ANALYTICS := 1

# Repository root directory
REPO_ROOT := $(shell git rev-parse --show-toplevel)

NUM_PROCESSORS := $(shell bash $(REPO_ROOT)/nprocs.sh 2>/dev/null || echo "1")
ifeq ($(NUM_PROCESSORS),)
NUM_PROCESSORS := 1
endif

export GOMAXPROCS="$(NUM_PROCESSORS)"

.PHONY: init lint markdownlint test test-go build release

# Determine path to svu binary using go env (prefers GOBIN over GOPATH/bin)
SVU_BIN := $(shell bash -lc 'if [ -n "$$(/usr/bin/env go env GOBIN)" ]; then echo "$$(/usr/bin/env go env GOBIN)/svu"; else echo "$$(/usr/bin/env go env GOPATH)/bin/svu"; fi')

# Install required developer tools via Homebrew Brewfile and set up Husky git hooks
init:
	@if [ -n "$$CI" ]; then \
	  brew bundle --no-upgrade --file=Brewfile; \
	  echo "CI: skipping npm install; handled by workflow cache"; \
	else \
	  brew bundle --file=Brewfile; \
	  npm ci || npm install; \
	fi
	@git config core.hooksPath .husky
	@chmod +x .husky/pre-push || true
	@go mod tidy

# Run markdownlint on Markdown files known to Git (respects .gitignore)
markdownlint: init
	@if ! command -v markdownlint-cli2 >/dev/null 2>&1; then \
		echo "markdownlint-cli2 not installed. Run: make init"; \
		exit 1; \
	fi; \
	FILES=$$(git ls-files --cached --others --exclude-standard -- '*.md'); \
	if [ -z "$$FILES" ]; then echo "No Markdown files found."; exit 0; fi; \
	git ls-files -z --cached --others --exclude-standard -- '*.md' | xargs -0 markdownlint-cli2

# Run linters and auto-fix simple issues when possible
lint: markdownlint init
	golangci-lint run --fix --allow-parallel-runners

# Aggregate test target: runs lint and go tests with coverage
# (no commands of its own)
test: lint test-go init

# Run Go tests with coverage and generate HTML report
# Outputs:
#  - coverage.out   (coverage profile)
#  - coverage.html  (HTML report)
# Also prints the total coverage line to stdout
test-go: init
	@set -euo pipefail; \
	echo ""; \
	echo "RUNNING GO TESTS..."; \
	echo ""; \
	go tool gotestsum -f pkgname-and-test-fails -- -v -p $(NUM_PROCESSORS) -parallel $(NUM_PROCESSORS) ./... -count 1 -coverprofile=coverage.out -covermode=atomic; \
	go tool cover -html=coverage.out -o coverage.html

# Build artifacts using GoReleaser
# Uses snapshot mode so it doesn't require a VCS tag or publish a release
build: init
	goreleaser --parallelism $(NUM_PROCESSORS) build --snapshot --clean

# Create and push a new git tag based on semantic version analysis by svu
# Requires a clean working tree and an "origin" remote.
release: init
	@VERSION="$(shell cd $(REPO_ROOT) && $(SVU_BIN) next --force-patch-increment)"; \
	echo "Computed tag for next version: $$VERSION"; \
	git tag "$$VERSION"; \
	git push --tags; \
	goreleaser --parallelism $(NUM_PROCESSORS) release --clean
