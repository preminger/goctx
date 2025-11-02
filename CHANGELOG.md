# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[unreleased]: https://github.com/olivierlacan/keep-a-changelog/compare/v0.4.2...HEAD
[0.4.2]: https://github.com/olivierlacan/keep-a-changelog/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/olivierlacan/keep-a-changelog/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/olivierlacan/keep-a-changelog/releases/tag/v0.4.0
