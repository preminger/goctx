SHELL := /bin/bash

# Repository root directory
REPO_ROOT := $(shell git rev-parse --show-toplevel)

.PHONY: init lint markdownlint test test-unit build release

# Determine path to svu binary using go env (prefers GOBIN over GOPATH/bin)
SVU_BIN := $(shell bash -lc 'if [ -n "$$(/usr/bin/env go env GOBIN)" ]; then echo "$$(/usr/bin/env go env GOBIN)/svu"; else echo "$$(/usr/bin/env go env GOPATH)/bin/svu"; fi')

# Install required developer tools via Homebrew Brewfile
init:
	brew bundle install

# Run markdownlint on all Markdown files
markdownlint: init
	@if ! command -v markdownlint-cli2 >/dev/null 2>&1; then \
		echo "markdownlint-cli2 not installed. Run: make init"; \
		exit 1; \
	fi; \
	markdownlint-cli2 "**/*.md"

# Run linters and auto-fix simple issues when possible
lint: markdownlint init
	golangci-lint run --fix --allow-parallel-runners

# Aggregate test target: runs lint and unit tests with coverage
# (no commands of its own)
test: lint test-unit init

# Run unit tests with coverage and generate HTML report
# Outputs:
#  - coverage.out   (coverage profile)
#  - coverage.html  (HTML report)
# Also prints the total coverage line to stdout
test-unit: init
	set -euo pipefail; \
	go test ./... -coverprofile=coverage.out -covermode=atomic; \
	go tool cover -func=coverage.out | tail -n 1; \
	go tool cover -html=coverage.out -o coverage.html

# Build artifacts using GoReleaser
# Uses snapshot mode so it doesn't require a VCS tag or publish a release
build: init
	goreleaser build --snapshot --clean

# Create and push a new git tag based on semantic version analysis by svu
# Requires a clean working tree and an "origin" remote.
release: init
	@VERSION="$(shell cd $(REPO_ROOT) && $(SVU_BIN) next --force-patch-increment)"; \
	echo "Computed tag for next version: $$VERSION"; \
	git tag "$$VERSION"; \
	git push --tags; \
	goreleaser release --clean
