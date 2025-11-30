//go:build mage

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/samber/lo"

	"github.com/preminger/goctx/internal/ui"
)

// Default target to run when none is specified.
var Default = All

func All() error {
	mg.Deps(Init, Test)
	mg.Deps(Build)

	return nil
}

// Init installs required tools and sets up git hooks and modules.
func Init() error { // mage:help=Install dev tools (Brewfile), setup husky hooks, and tidy modules
	nCores, err := getNumberOfCores()
	if err != nil {
		return err
	}

	// Set GOMAXPROCS env var.
	if err := os.Setenv("GOMAXPROCS", strconv.Itoa(nCores)); err != nil {
		return err
	}

	// Install tools from Brewfile.
	if err := sh.Run("brew", "bundle", "--file=Brewfile"); err != nil {
		return err
	}

	// Install npm.
	if os.Getenv("CI") == "" {
		if err := sh.Run("npm", "ci"); err != nil {
			if err := sh.Run("npm", "install"); err != nil {
				return err
			}
		}
	} else {
		slog.Debug("in CI; skipping explicit npm installation")
	}

	// Set up husky git hooks.
	if err := sh.Run("git", "config", "core.hooksPath", ".husky"); err != nil {
		return err
	}
	if err := sh.Run("chmod", "+x", ".husky/pre-push"); err != nil {
		return err
	}

	return sh.Run("go", "mod", "tidy")
}

// Markdownlint runs markdownlint-cli2 on all tracked Markdown files.
func Markdownlint() error { // mage:help=Run markdownlint on Markdown files
	mg.Deps(Init)

	markdownFilesList, err := sh.Output("git", "ls-files", "--cached", "--others", "--exclude-standard", "--", "*.md")
	if err != nil {
		return err
	}

	markdownFilesList = strings.TrimSpace(markdownFilesList)
	if markdownFilesList == "" {
		slog.Info("No Markdown files found to lint. Skipping.")
		return nil
	}

	files := lo.Filter(strings.Split(markdownFilesList, "\n"), func(s string, _ int) bool {
		return !lo.IsEmpty(s)
	})

	return sh.Run("markdownlint-cli2", files...)
}

// Lint runs golangci-lint after markdownlint and init.
func Lint() error { // mage:help=Run linters and auto-fix issues
	mg.Deps(Markdownlint, Init)

	out, err := sh.Output("golangci-lint", "run", "--fix", "--allow-parallel-runners", "--build-tags=mage")
	if err != nil {
		titleStyle, blockStyle := ui.GetBlockStyles()
		_, _ = fmt.Println(titleStyle.Render("golangci-lint output"))
		_, _ = fmt.Println(blockStyle.Render(out))
		_, _ = fmt.Println()
		return err
	}

	return nil
}

// Test aggregate target runs Lint and TestGo.
func Test() { // mage:help=Run lint and Go tests with coverage
	mg.Deps(Init, Lint, TestGo)
}

// TestGo runs Go tests with coverage and produces coverage.out and coverage.html.
func TestGo() error { // mage:help=Run Go tests with coverage (coverage.out, coverage.html)
	mg.Deps(Init)

	nCores, err := getNumberOfCores()
	if err != nil {
		return err
	}
	nCoresStr := strconv.Itoa(nCores)

	if err := sh.RunV(
		"go", "tool", "gotestsum", "-f", "pkgname-and-test-fails",
		"--",
		"-v", "-p", nCoresStr, "-parallel", nCoresStr, "./...", "-count", "1",
		"-coverprofile=coverage.out", "-covermode=atomic",
	); err != nil {
		return err
	}

	return sh.Run("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
}

// Build artifacts via goreleaser snapshot build.
func Build() error { // mage:help=Build artifacts using goreleaser (snapshot)
	mg.Deps(Init)

	nCores, err := getNumberOfCores()
	if err != nil {
		return err
	}

	if err := sh.RunV("goreleaser", "check"); err != nil {
		return err
	}

	return sh.RunV("goreleaser", "--parallelism", strconv.Itoa(nCores), "build", "--snapshot", "--clean")
}

// Release tags the next version with svu and runs goreleaser release.
func Release() error { // mage:help=Create and push a new tag with svu, then goreleaser
	mg.Deps(Init)

	goBin, err := sh.Output("go", "env", "GOBIN")
	if err != nil {
		return err
	}
	goBin = strings.TrimSpace(goBin)

	if goBin == "" {
		goPath, err := sh.Output("go", "env", "GOPATH")
		if err != nil {
			return err
		}

		goBin = filepath.Join(strings.TrimSpace(goPath), "bin")
	}

	svuPath := filepath.Join(goBin, "svu")
	slog.Debug("svu binary path", slog.String("path", svuPath))
	nextVersion, err := sh.Output(svuPath, "next", "--force-patch-increment")
	if err != nil {
		return err
	}

	nextVersion = strings.TrimSpace(nextVersion)
	if nextVersion == "" {
		return errors.New("svu returned empty version")
	}

	slog.Info("computed next version", slog.String("version", nextVersion))

	if err := sh.Run("git", "tag", nextVersion); err != nil {
		return err
	}

	if err := sh.Run("git", "push", "--tags"); err != nil {
		return err
	}

	nCores, err := getNumberOfCores()
	if err != nil {
		return err
	}

	return sh.Run("goreleaser", "--parallelism", strconv.Itoa(nCores), "release", "--clean")
}

// getRepoRoot returns the absolute path to the repository root (git top-level).
func getRepoRoot() (string, error) {
	out, err := sh.Output("git", "rev-parse", "--show-toplevel")
	if err != nil {
		slog.Warn("error running `git rev-parse --show-toplevel`", slog.Any("error", err))

		// Fallback to current working dir on failure
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		return cwd, nil
	}

	return strings.TrimSpace(out), nil
}

// getNumberOfCores tries to detect number of processors using nprocs.sh. Falls back to 1.
func getNumberOfCores() (int, error) {
	root, err := getRepoRoot()
	if err != nil {
		return 1, err
	}

	utility := filepath.Join(root, "nprocs.sh")
	out, err := sh.Output("bash", utility)
	if err != nil {
		slog.Warn("error running nprocs utility", slog.String("path", utility), slog.Any("error", err))
		return 1, nil
	}
	out = strings.TrimSpace(out)
	if out == "" {
		slog.Warn("nprocs utility returned empty string", slog.String("path", utility))
		return 1, nil
	}

	intVal, err := strconv.Atoi(out)
	if err != nil {
		slog.Warn("nprocs utility returned invalid value", slog.String("path", utility), slog.Any("error", err))
		return 1, nil
	}

	return intVal, nil
}
