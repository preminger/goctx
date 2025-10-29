package goctx

import (
	"context"
	"fmt"

	"runtime/debug"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Version is the CLI version. It can be overridden at build time via:
//
//	-ldflags "-X github.com/preminger/goctx/cmd/goctx.Version=v0.0.0"
//
// If left as "dev", we will attempt to detect the version from Git metadata
// at runtime (git describe) or, as a fallback, from Go build info.
var Version = "dev" //nolint:gochecknoglobals // Populated by goreleaser ldflags.

// Commit is the git commit hash. It can be overridden at build time via:
//
//	-ldflags "-X github.com/preminger/goctx/cmd/goctx.Commit=<commit>"
var Commit = "" //nolint:gochecknoglobals // Populated by goreleaser ldflags.

// BuildDate is the RFC3339 timestamp of the build. It can be overridden via:
//
//	-ldflags "-X github.com/preminger/goctx/cmd/goctx.BuildDate=<RFC3339>"
var BuildDate = "" //nolint:gochecknoglobals // Populated by goreleaser ldflags.

// EffectiveVersion returns the best-effort version string for the binary.
// Precedence:
//  1. If Version was set via -ldflags and is not "dev"/empty, use it as-is.
//  2. If built via `go install module@version`, use Go build info `Main.Version`.
//  3. Fallback to Go build info `vcs.revision` (+ "-dirty" if `vcs.modified=true`).
//  4. Finally, return "dev".
func EffectiveVersion(_ context.Context) string {
	v := strings.TrimSpace(Version)
	if v != "" && v != "dev" {
		// Caller injected a version via ldflags
		return v
	}

	// Prefer the module version embedded by the Go toolchain when installed via
	// `go install module@version` (e.g., v0.2.0). When built from source it is usually
	// "(devel)" and thus ignored.
	if bi, ok := debug.ReadBuildInfo(); ok && bi != nil {
		if mv := strings.TrimSpace(bi.Main.Version); mv != "" && mv != "(devel)" {
			return mv
		}
		var rev string
		var dirty string
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				rev = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					dirty = "-dirty"
				}
			}
		}
		if rev != "" {
			return rev + dirty
		}
	}

	return "dev"
}

// EffectiveCommit returns the preferred commit hash for the build.
// Precedence:
// 1) Commit from ldflags, if provided.
// 2) Go build info `vcs.revision` (if available).
func EffectiveCommit(_ context.Context) string {
	c := strings.TrimSpace(Commit)
	if c != "" {
		return c
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi != nil {
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" && s.Value != "" {
				return s.Value
			}
		}
	}
	return ""
}

// EffectiveBuildTime returns the build time as RFC3339 string when available.
// Precedence:
// 1) BuildDate from ldflags if provided.
// 2) Go build info `vcs.time`.
func EffectiveBuildTime() string {
	if t, ok := EffectiveBuildTimeParsed(); ok {
		return t.UTC().Format(time.RFC3339)
	}
	return ""
}

// EffectiveBuildTimeParsed attempts to parse the build time into time.Time.
// It tries RFC3339 and RFC3339Nano layouts.
// Returns the parsed time and true on success; zero time and false otherwise.
func EffectiveBuildTimeParsed() (time.Time, bool) {
	bd := strings.TrimSpace(BuildDate)
	if bd != "" {
		if t, ok := parseRFC3339MaybeNano(bd); ok {
			return t, true
		}
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi != nil {
		for _, s := range bi.Settings {
			if s.Key == "vcs.time" && s.Value != "" {
				if t, ok := parseRFC3339MaybeNano(s.Value); ok {
					return t, true
				}
			}
		}
	}
	return time.Time{}, false
}

// parseRFC3339MaybeNano tries RFC3339 and RFC3339Nano.
func parseRFC3339MaybeNano(v string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func NewVersionCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information and exit",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Version: %s\n", EffectiveVersion(ctx))
			if c := EffectiveCommit(ctx); c != "" {
				fmt.Fprintf(out, "Commit: %s\n", c)
			}
			if t, ok := EffectiveBuildTimeParsed(); ok {
				local := t.In(time.Local)
				// Use a readable local format including zone name
				fmt.Fprintf(out, "Built:  %s\n", local.Format(time.RFC1123))
			} else if raw := EffectiveBuildTime(); raw != "" {
				// Fallback to raw string if parsing failed but value exists
				fmt.Fprintf(out, "Built:  %s\n", raw)
			}
		},
	}
}
