# goctx (contextualize)

Propagate context.Context through Go call graphs automatically.

This tool analyzes your Go source, ensures a `context.Context` parameter exists where needed, and propagates it through call chains by updating function signatures and call sites. It can also stop at well-defined boundaries such as `http.HandlerFunc`, deriving the context from `req.Context()`.

- Input: a target function specified as `path/to/file.go:FuncName[:N]` (with optional line number).
- Output: in-place source edits to add or thread a `context.Context` (named `ctx`) through your code.

## Why

- Standardize context propagation without tedious, error-prone manual edits.
- Enforce best practices around cancellation, deadlines, and request-scoped values.
- Make large refactors safer by relying on static analysis of call graphs.

## Installation

You can install the CLI with a regular `go install` or build from source.

- Latest version using `go install`:

  go install github.com/example/contextualize/app/goctx@latest

  This will install a `goctx` binary in your `GOBIN` (or `$GOPATH/bin`).

- Build from source (requires Go and optionally goreleaser):

  - Simple build:

    go build -o bin/goctx ./app/goctx

  - Using the provided Makefile and GoReleaser (snapshot build):

    make build

  Artifacts will be placed under the `dist/` directory when using GoReleaser.

## Quick start

- Ensure your working directory is the root of the module you want to modify (the tool defaults to `WorkDir: "."`).
- Run `goctx` with a target function path of the form `path/to/file.go:FuncName[:N]`.

Examples:

- Add/propagate ctx into a function and its callers:

  goctx ./internal/foo/bar.go:DoThing

- Stop propagation at an explicit boundary function:

  goctx --stop-at ./internal/foo/baz.go:BoundaryFunc ./internal/foo/bar.go:DoThing

- Stop at HTTP boundaries and derive ctx from `req.Context()` automatically:

  goctx --http ./internal/http/handler.go:ServeHTTP

Notes:
- The optional `:N` suffix can be used if there are multiple functions with the same name in the file and you want to disambiguate by line number (N is the 1-based starting line of the function).
- The tool will rename a blank `_` `context.Context` parameter to `ctx` in the target when needed.

## Usage

The CLI surface is intentionally small. Flags are also visible via `goctx --help`.

Usage:

  goctx [flags] "<path/to/file.go:FuncName[:N]>"

Flags:
- --stop-at string
  Optional terminating function path `path/to/file.go:FuncName[:N]` where propagation should stop.
- --http
  Treat HTTP handlers (`http.HandlerFunc`) as boundaries and derive `ctx` from `req.Context()`.

Behavior summary:
- If the target function has no `context.Context` parameter, one named `ctx` will be added.
- All call sites to the modified function will be updated to pass `ctx` (or a derived context if required by boundaries).
- Propagation continues along the call graph until a stopping point (explicit `--stop-at`, HTTP boundary when `--http` is set, the `main` function, or other analysis-defined limits).
- Only files actually modified are written back to disk.

## Examples in this repo

The repository contains end-to-end test inputs and golden outputs under:
- internal/contextualize/testdata/input/
- internal/contextualize/testdata/golden/

You can inspect pairs like these to understand before/after transformations:
- internal/contextualize/testdata/input/e2e_ctxparam_blank_to_ctx/main.go
- internal/contextualize/testdata/golden/e2e_ctxparam_blank_to_ctx_main_go.golden

There are additional scenarios covering propagation across multiple packages and HTTP boundaries.

## Development

Requirements:
- Go (a recent version per go.mod)
- Optional: Homebrew for developer tooling via Brewfile

Setup developer tools via Homebrew:

  make init

Lint:

  make lint

Run tests with coverage (also generates coverage.html):

  make test-unit

Or run both lint and unit tests:

  make test

Run a snapshot build using GoReleaser:

  make build

### Project layout

- app/goctx: CLI entrypoint (main package), produces the `goctx` binary
- cmd/goctx: Cobra command and flags
- internal/contextualize: Core analysis and rewrite logic
- internal/contextualize/testdata: End-to-end fixtures used by tests (input and golden files)

## Troubleshooting

- Nothing happens / no files change
  - Ensure your target function path is correct and relative to your module root, and that you run the tool from the module root (or specify the correct working directory before running).
  - Use the optional line suffix `:N` to pinpoint the exact function if there are duplicates.

- Build breaks after propagation
  - Run `go build ./...` and inspect errors to identify where additional manual adjustments might be needed, especially around interfaces or third-party APIs.

- Undoing changes
  - The tool writes in-place. Commit your work before running, or use your VCS to review/undo.

## Contributing

Contributions are welcome! Please:
- Open an issue describing the problem or enhancement.
- For code changes, include tests. You can look at existing tests in `internal/contextualize` as examples.
- Run `make test` before submitting a PR.

## License

This project is licensed under the terms of the MIT License. See the [LICENSE](LICENSE) file for details.
