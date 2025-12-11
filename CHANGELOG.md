# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.16.1] - 2025-12-11

### Changed

- Bump Go version to `1.25.5`.

- Bumped all updatable Go dependencies as of this date.

## [0.16.0] - 2025-12-10

### Changed

- Move from [Mage](https://magefile.org/) to [Stave](https://github.com/yaklabco/stave) as build tool.

- Change `release.yml` workflow triggering to manual-only.

## [0.15.0] - 2025-12-09

### Added

- Added `-t`/`--tags` option to the CLI to control Go build tags affecting file visibility during analysis.

## [0.14.7] - 2025-12-08

### Changed

- Changelog catch-up.

## [0.14.6] - 2025-12-08

### Changed

- Re-release of `v0.14.5` due to tagging problem.

## [0.14.5] - 2025-12-08

### Changed

- Bumped all updatable Go dependencies as of this date.

## [0.14.4] - 2025-12-08

### Changed

- Upgraded `caarlos0/svu` to `v3`, and removed deprecated `--force-patch-increment` flag from all its invocations.

## [0.14.3] - 2025-12-08

### Changed

- Drop minimum Go version to `1.24.11` (was: `1.25.4`).

## [0.14.2] - 2025-12-02

### Changed

- Modernize testing code.

## [0.14.1] - 2025-12-01

### Changed

- Inhibit pre-push `svu next --force-patch-increment` CHANGELOG.md hook in release flow.

## [0.14.0] - 2025-12-01

### Added

- Add to pre-push hook a check that output of `svu next --force-patch-increment` is represented in CHANGELOG.md

## [0.13.0] - 2025-12-01

### Added

- Treat `func TestMain(m *testing.M)` as boundary, along the same lines as `func main()`.

## [0.12.0] - 2025-12-01

### Added

- Regression tests for the fixes contained in this release (see below).

### Fixed

- When you are in a subdir *below* where the `go.mod` is, and you run `goctx` with a TARGET argument that specifies a function in a source file in the current dir - but that function is called elsewhere in the module *outside* of this subdir - the app now adjusts all the relevant call sites within the module, not just the ones in the current subdir.

- Both functions defined in `*_test.go` files, and callsites contained in `*_test.go` files (even in cases where no signatures of functions defined in `*_test.go` were changed), are now correctly handled.

- App now correctly distinguishes between `MyFunc()`, `xyz.MyFunc()` (where `xyz` is a package name), `a.MyFunc()` (where `a` is an object of type `TypeA`), and `b.MyFunc()` (where `b` is an object of type `TypeB`). Changes to the function signature of one of these will not affect call sites where the others are called.

- When a function `MyFunc` contains calls to two functions `MyOtherFunc1` and `MyOtherFunc2`, and a context argument has been added to both `MyOtherFunc1` and `MyOtherFunc2`, only one context argument will be added to `MyFunc`'s signature (and that single argument will be used in the calls to both `MyOtherFunc1` and `MyOtherFunc2`).

### Changed

- Structure and dir- and file-naming under `pkg/goctx/testdata/{input,golden}` has been organized & rationalized. In particular, subdirs of `pkg/goctx/testdata/input` and `pkg/goctx/testdata/golden` are now named identically to the test in which they are used.

## [0.11.4] - 2025-12-01

### Added

- CHANGELOG.md catch-up.

## [0.11.3] - 2025-11-30

### Added

- CHANGELOG.md dates correction

## [0.11.2] - 2025-11-30

### Added

- CHANGELOG.md catch-up.

## [0.11.1] - 2025-11-30

### Changed

- Update documentation to reflect move to [mage](https://magefile.org/) (contributed by [James Ainslie](https://github.com/jamesainslie)).

## [0.11.0] - 2025-11-29

### Changed

- Changed build system from make to [mage](https://magefile.org/).

## [0.10.8] - 2025-11-29

### Changed

- Some light refactoring of cmd/goctx/version.go

## [0.10.7] - 2025-11-29

### Added

- Add `goreleaser check` step to Makefile `build` target.

## [0.10.6] - 2025-11-29

### Changed

- Propagate recent documentation changes to README.md

## [0.10.5] - 2025-11-29

### Changed

- Misc. documentation improvements.

## [0.10.4] - 2025-11-28

### Fixed

- Loop re-entry & termination logic in `fs.TruePath(...)` (used for path normalization, calculating absolute paths & resolving symlinks).

## [0.10.3] - 2025-11-28

### Removed

- Removed a misleading debug log message.

## [0.10.2] - 2025-11-28

### Fixed

- When calculating absolute paths for the purposes of paths-comparison, resolve symlinks as well.

## [0.10.1] - 2025-11-28

### Changed

- Removed colorization of dashes in version-string.

## [0.10.0] - 2025-11-28

### Added

- Colorization of version string in `goctx --version`.

## [0.9.4] - 2025-11-28

### Changed

- Maximize parallelism in build & CI flows.

## [0.9.3] - 2025-11-28

### Changed

- Bumped all updatable Go dependencies as of this date.

## [0.9.2] - 2025-11-28

### CI

- Make use of caching within GitHub workflows.

## [0.9.1] - 2025-11-28

### Developer infrastructure

- Added first "wave" of debug-logging.

## [0.9.0] - 2025-11-26

### New features

- Add basic logging infrastructure, and `-v`/`--verbose` flag to enable debug-logging.

### API changes

- Refactored `internal/contextualize` -> `pkg/goctx`.

### Testing

- Fixed initialization & handling of command-line flags to be compatible with current testing approach.

## [0.8.6] - 2025-11-17

### Changed

- Bumped all updatable Go dependencies as of this date.

## [0.8.5] - 2025-11-08

### Changed

- Bumped Go version to `1.25.4`.

## [0.8.4] - 2025-11-03

### Added

- Use `gotestsum` for human-readable `go test` output.

### Changed

- Makefile target for running Go tests renamed from `test-unit` to `test-go`.
- Added `-count 1` flag to Go test command-line to circumvent test caching.

## [0.8.3] - 2025-11-03

### Fixed

- Links at the bottom of the CHANGELOG now point to the right repo.

## [0.8.2] - 2025-11-03

### Added

- Documentation comment on `ExecuteWithFang(...)` function.

## [0.8.1] - 2025-11-03

### Changed

- Change `cobra` description of app to match GitHub & Homebrew descriptions.

## [0.8.0] - 2025-11-03

### Fixed

- When no TARGET function provided on command-line, print out help.

## [0.7.0] - 2025-11-03

### Added

- Pre-push hook to ensure that PRs include a change to the CHANGELOG.

## [0.6.0] - 2025-11-03

### Added

- Use `"github.com/charmbracelet/fang"` wrapping for `"github.com/spf13/cobra"` functionality.

## [0.5.6] - 2025-11-03

### Added

- The `$(SVU_BIN)` binary in the Makefile `release` target is explicitly run with `$(REPO_ROOT)` as the working directory.

## [0.5.5] - 2025-11-03

### Changed

- Makefile targets `test`, `build`, and `release` now depend on `init` target.

## [0.5.4] - 2025-11-02

### Fixed

- Invoke `svn` from Makefile using `$(shell ... )` notation.

## [0.5.3] - 2025-11-02

### Removed

- Printing of what next git tag *would* be in Makefile `test` target, and associated flags in `actions/checkout@v4` step in `checks` workflow.

## [0.5.2] - 2025-11-02

### Fixed

- Changelog fixes.

## [0.5.1] - 2025-11-02

### Fixed

- Swap inner & outer quotation marks in Homebrew description in .goreleaser.yaml

## [0.5.0] - 2025-11-02

### Added

- Print message about what next git tag *would* be in Makefile `test` target.

### Fixed

- Computation of next git tag in `release` workflow.

## [0.4.6] - 2025-11-02

### Changed

- Include the term "tag" in the message about next computed tag in Makefile `release` target.

## [0.4.5] - 2025-11-02

### Fixed

- Use explicit output of `go env` to determine `svu` binary path in Makefile.

## [0.4.4] - 2025-11-02

### Added

- Output result of `svu` call to stdout before using value in `git tag` command.

## [0.4.3] - 2025-11-02

### Changed

- Update the Homebrew cask description in .goreleaser.yaml so that it matches the current description of the GitHub repo at <https://github.com/preminger/goctx/>.

## [0.4.2] - 2025-11-01

### Changed

- Original contents of `func main()` now reside in a function `func actualMain() int`, which returns a value that `main()` then feeds to `os.Exit(...)`. This allows `defer` patterns to work correctly in main function.

## [0.4.1] - 2025-11-01

### Added

- This CHANGELOG!

## [0.4.0] - 2025-10-30

[unreleased]: https://github.com/preminger/goctx/compare/v0.16.1...HEAD
[0.16.1]: https://github.com/preminger/goctx/compare/v0.16.0...v0.16.1
[0.16.0]: https://github.com/preminger/goctx/compare/v0.15.0...v0.16.0
[0.15.0]: https://github.com/preminger/goctx/compare/v0.14.7...v0.15.0
[0.14.7]: https://github.com/preminger/goctx/compare/v0.14.6...v0.14.7
[0.14.6]: https://github.com/preminger/goctx/compare/v0.14.5...v0.14.6
[0.14.5]: https://github.com/preminger/goctx/compare/v0.14.4...v0.14.5
[0.14.4]: https://github.com/preminger/goctx/compare/v0.14.3...v0.14.4
[0.14.3]: https://github.com/preminger/goctx/compare/v0.14.2...v0.14.3
[0.14.2]: https://github.com/preminger/goctx/compare/v0.14.1...v0.14.2
[0.14.1]: https://github.com/preminger/goctx/compare/v0.14.0...v0.14.1
[0.14.0]: https://github.com/preminger/goctx/compare/v0.13.0...v0.14.0
[0.13.0]: https://github.com/preminger/goctx/compare/v0.12.0...v0.13.0
[0.12.0]: https://github.com/preminger/goctx/compare/v0.11.4...v0.12.0
[0.11.4]: https://github.com/preminger/goctx/compare/v0.11.3...v0.11.4
[0.11.3]: https://github.com/preminger/goctx/compare/v0.11.2...v0.11.3
[0.11.2]: https://github.com/preminger/goctx/compare/v0.11.1...v0.11.2
[0.11.1]: https://github.com/preminger/goctx/compare/v0.11.0...v0.11.1
[0.11.0]: https://github.com/preminger/goctx/compare/v0.10.8...v0.11.0
[0.10.8]: https://github.com/preminger/goctx/compare/v0.10.7...v0.10.8
[0.10.7]: https://github.com/preminger/goctx/compare/v0.10.6...v0.10.7
[0.10.6]: https://github.com/preminger/goctx/compare/v0.10.5...v0.10.6
[0.10.5]: https://github.com/preminger/goctx/compare/v0.10.4...v0.10.5
[0.10.4]: https://github.com/preminger/goctx/compare/v0.10.3...v0.10.4
[0.10.3]: https://github.com/preminger/goctx/compare/v0.10.2...v0.10.3
[0.10.2]: https://github.com/preminger/goctx/compare/v0.10.1...v0.10.2
[0.10.1]: https://github.com/preminger/goctx/compare/v0.10.0...v0.10.1
[0.10.0]: https://github.com/preminger/goctx/compare/v0.9.4...v0.10.0
[0.9.4]: https://github.com/preminger/goctx/compare/v0.9.3...v0.9.4
[0.9.3]: https://github.com/preminger/goctx/compare/v0.9.2...v0.9.3
[0.9.2]: https://github.com/preminger/goctx/compare/v0.9.1...v0.9.2
[0.9.1]: https://github.com/preminger/goctx/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/preminger/goctx/compare/v0.8.6...v0.9.0
[0.8.6]: https://github.com/preminger/goctx/compare/v0.8.5...v0.8.6
[0.8.5]: https://github.com/preminger/goctx/compare/v0.8.4...v0.8.5
[0.8.4]: https://github.com/preminger/goctx/compare/v0.8.3...v0.8.4
[0.8.3]: https://github.com/preminger/goctx/compare/v0.8.2...v0.8.3
[0.8.2]: https://github.com/preminger/goctx/compare/v0.8.1...v0.8.2
[0.8.1]: https://github.com/preminger/goctx/compare/v0.8.0...v0.8.1
[0.8.0]: https://github.com/preminger/goctx/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/preminger/goctx/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/preminger/goctx/compare/v0.5.6...v0.6.0
[0.5.6]: https://github.com/preminger/goctx/compare/v0.5.5...v0.5.6
[0.5.5]: https://github.com/preminger/goctx/compare/v0.5.4...v0.5.5
[0.5.4]: https://github.com/preminger/goctx/compare/v0.5.3...v0.5.4
[0.5.3]: https://github.com/preminger/goctx/compare/v0.5.2...v0.5.3
[0.5.2]: https://github.com/preminger/goctx/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/preminger/goctx/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/preminger/goctx/compare/v0.4.6...v0.5.0
[0.4.6]: https://github.com/preminger/goctx/compare/v0.4.5...v0.4.6
[0.4.5]: https://github.com/preminger/goctx/compare/v0.4.4...v0.4.5
[0.4.4]: https://github.com/preminger/goctx/compare/v0.4.3...v0.4.4
[0.4.3]: https://github.com/preminger/goctx/compare/v0.4.2...v0.4.3
[0.4.2]: https://github.com/preminger/goctx/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/preminger/goctx/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/preminger/goctx/releases/tag/v0.4.0
