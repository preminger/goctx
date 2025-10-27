SHELL := /bin/bash

.PHONY: init lint test test-unit build

# Install required developer tools via Homebrew Brewfile
init:
	brew bundle install --no-lock

# Run linters and auto-fix simple issues when possible
lint:
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
