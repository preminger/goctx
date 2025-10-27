SHELL := /bin/bash

.PHONY: init lint markdownlint test test-unit build release

# Install required developer tools via Homebrew Brewfile
init:
	brew bundle install --no-lock

# Run markdownlint on all Markdown files
markdownlint:
	@if ! command -v markdownlint-cli2 >/dev/null 2>&1; then \
		echo "markdownlint-cli2 not installed. Run: make init"; \
		exit 1; \
	fi; \
	markdownlint-cli2 "**/*.md"

# Run linters and auto-fix simple issues when possible
lint: markdownlint
	golangci-lint run --fix --allow-parallel-runners

# Aggregate test target: runs lint and unit tests with coverage
# (no commands of its own)
test: lint test-unit

# Run unit tests with coverage and generate HTML report
# Outputs:
#  - coverage.out   (coverage profile)
#  - coverage.html  (HTML report)
# Also prints the total coverage line to stdout
test-unit:
	set -euo pipefail; \
	go test ./... -coverprofile=coverage.out -covermode=atomic; \
	go tool cover -func=coverage.out | tail -n 1; \
	go tool cover -html=coverage.out -o coverage.html

# Build artifacts using GoReleaser
# Uses snapshot mode so it doesn't require a VCS tag or publish a release
build:
	goreleaser build --snapshot --clean

# Create and push a new git tag based on semantic version analysis by svu
# Requires a clean working tree and an "origin" remote.
release:
	set -euo pipefail; \
	if ! command -v svu >/dev/null 2>&1; then \
		echo "svu not installed. Install with: brew install caarlos0/tap/svu"; \
		exit 1; \
	fi; \
	if ! command -v git >/dev/null 2>&1; then \
		echo "git is required but not found"; \
		exit 1; \
	fi; \
	if [ -n "$$(git status --porcelain)" ]; then \
		echo "Working tree not clean. Commit or stash changes before releasing."; \
		exit 1; \
	fi; \
	if ! git ls-remote --exit-code origin >/dev/null 2>&1; then \
		echo "No 'origin' remote configured or not reachable. Configure it before releasing."; \
		exit 1; \
	fi; \
	NEXT_TAG=$$(svu next); \
	echo "Next version: $$NEXT_TAG"; \
	git tag -a "$$NEXT_TAG" -m "$$NEXT_TAG"; \
	git push origin "$$NEXT_TAG"
