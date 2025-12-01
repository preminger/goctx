# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.11.2] - 2025-11-29

### Added

- CHANGELOG.md catch-up.

## [0.11.1] - 2025-11-29

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

- Printing of what next git tag _would_ be in Makefile `test` target, and associated flags in `actions/checkout@v4` step in `checks` workflow.

## [0.5.2] - 2025-11-02

### Fixed

- Changelog fixes.

## [0.5.1] - 2025-11-02

### Fixed

- Swap inner & outer quotation marks in Homebrew description in .goreleaser.yaml

## [0.5.0] - 2025-11-02

### Added

- Print message about what next git tag _would_ be in Makefile `test` target.

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

[unreleased]: https://github.com/preminger/goctx/compare/v0.11.2...HEAD
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
