# goctx

## tl;dr

"I'm in a function 10 calls deep from the closest `context.Context` object, and I suddenly find myself in need of a context; I know that calling `context.Background()` deep in my call stack is a no-no, and `context.TODO()` just kicks the can down the road. Is there a tool out there that can just automagically add all the necessary `context.Context` function/method parameter plumbing for me?"

Now there is.

(It also handles existing unused context parameters - i.e., `_ context.Context` - intelligently by just renaming them instead of adding a new parameter, and also reuses existing `context.Context` parameters that are not named `ctx`, if present.)

### Example

#### Command-line

```shell
goctx ./main.go:targetFunc
```

#### Before

```go
package main

import "context"

func targetFunc() {}

func funcOne() {
	targetFunc()
}

func funcTwo() {
	funcOne()
}

func funcThreeA(_ context.Context) {
	funcTwo()
}

func funcThreeB() {
	funcOne()
}

func funcThreeC(myWeirdlyNamedCtx context.Context) {
	funcTwo()
}

func funcFour() {
	funcThreeB()
}

func main() {
	ctx := context.Background()
	funcFour()
	funcThreeA(ctx)
	funcThreeC(ctx)
}
```

#### After

```go
package main

import "context"

func targetFunc(ctx context.Context) {}

func funcOne(ctx context.Context) {
	targetFunc(ctx)
}

func funcTwo(ctx context.Context) {
	funcOne(ctx)
}

func funcThreeA(ctx context.Context) {
	funcTwo(ctx)
}

func funcThreeB(ctx context.Context) {
	funcOne(ctx)
}

func funcThreeC(myWeirdlyNamedCtx context.Context) {
	funcTwo(myWeirdlyNamedCtx)
}

func funcFour(ctx context.Context) {
	funcThreeB(ctx)
}

func main() {
	ctx := context.Background()
	funcFour(ctx)
	funcThreeA(ctx)
	funcThreeC(ctx)
}
```

## Description

Propagate `context.Context` through Go call graphs automatically.

This tool analyzes your Go source, ensures a `context.Context` parameter exists where needed, and propagates it through call chains by updating function signatures and call sites. It can also stop at well-defined boundaries such as `http.HandlerFunc`, deriving the context from `req.Context()`.

- Input: a target function specified as `path/to/file.go:FuncName[:N]` (with optional line number).
- Output: in-place source edits to add or thread a `context.Context` (named `ctx`) through your code.

## Changelog

If you are looking for the CHANGELOG for the project, it can be found [here](./CHANGELOG.md).

>[!WARNING]
> Any use of `goctx` is analogous to passing the `-w` flag to tools like `gofmt`, `goimports`, etc. - in other words, `goctx` CHANGES YOUR SOURCE FILES, writing the modified files in place. (The changes are strictly additive; but they are changes nonetheless.)
> There is no built-in mechanism for undoing these changes. _Please make diligent use of version-control!_

## Why

- Standardize context propagation without tedious, error-prone manual edits.
- Enforce best practices around cancellation, deadlines, and request-scoped values.
- Make large refactors safer by relying on static analysis of call graphs.

## Installation

You can install the CLI via Homebrew, with `go install`, or build from source.

- Homebrew (recommended for macOS/Linux):

  Option A: tap once, then install/update normally:

  ```shell
  brew tap preminger/tap
  brew install goctx
  # later updates
  brew upgrade goctx
  ```

  Option B: install directly from the tap without adding it globally:

  ```shell
  brew install preminger/tap/goctx
  ```

- Latest version using `go install`:

  ```shell
  go install github.com/preminger/goctx/app/goctx@latest
  ```

  This will install a `goctx` binary in your `GOBIN` (or `$GOPATH/bin`).

- Build from source (requires Go and optionally goreleaser):

  - Simple build:

    ```shell
    go build -o bin/goctx ./app/goctx
    ```

  - Using the provided Makefile and GoReleaser (snapshot build):

    ```shell
    make build
    ```

  Artifacts will be placed under the `dist/` directory when using GoReleaser.

## Quick start

- Ensure your working directory is the root of the module you want to modify (the tool defaults to `WorkDir: "."`).
- Run `goctx` with a target function path of the form `path/to/file.go:FuncName[:N]`.

Examples:

- Add/propagate ctx into a function and its callers:

```shell
goctx ./internal/foo/bar.go:FuncInNeedOfContext
```

- Stop propagation at an explicit boundary function:

```shell
goctx --stop-at ./internal/foo/baz.go:FuncServingAsBoundary ./internal/foo/bar.go:FuncInNeedOfContext
```

- Stop at HTTP boundaries and derive ctx from `req.Context()` automatically:

```shell
goctx --http ./internal/http/handler.go:ServeHTTP
```

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

- pkg/goctx/testdata/input/
- pkg/goctx/testdata/golden/

You can inspect pairs like these to understand before/after transformations:

- pkg/goctx/testdata/input/e2e_ctxparam_blank_to_ctx/main.go
- pkg/goctx/testdata/golden/e2e_ctxparam_blank_to_ctx_main_go.golden

There are additional scenarios covering propagation across multiple packages and HTTP boundaries.

## Development

Requirements:

- Go (a recent version per go.mod)
- Optional: Homebrew for developer tooling via Brewfile

Setup developer tools via Homebrew:

```shell
make init
```

Lint:

```shell
make lint
```

Run tests with coverage (also generates coverage.html):

```shell
make test-unit
```

Or run both lint and unit tests:

```shell
make test
```

Run a snapshot build using GoReleaser:

```shell
make build
```

### Project layout

- app/goctx: CLI entrypoint (main package), produces the `goctx` binary
- cmd/goctx: Cobra command and flags
- pkg/goctx: Core analysis and rewrite logic
- pkg/goctx/testdata: End-to-end fixtures used by tests (input and golden files)

## Troubleshooting

- Nothing happens / no files change
  - Ensure your target function path is correct and relative to your module root, and that you run the tool from the module root (or specify the correct working directory before running).
  - Use the optional line suffix `:N` to pinpoint the exact function if there are duplicates.

- Build breaks after propagation
  - Run `go build ./...` and inspect errors to identify where additional manual adjustments might be needed, especially around interfaces or third-party APIs.

- Undoing changes
  - The tool writes in-place. Commit your work before running, or use your VCS to review/undo.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the terms of the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.
