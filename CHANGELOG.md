# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[unreleased]: https://github.com/preminger/goctx/compare/v0.8.4...HEAD
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
