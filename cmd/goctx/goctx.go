package goctx

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/preminger/goctx/pkg/goctx"
	"github.com/spf13/cobra"
)

const shortDescription = "Command-line Go utility that automatically adds missing 'plumbing' for `context.Context` parameters along the call-graph leading to a given function."

func NewRootCmd(ctx context.Context) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "goctx TARGET",
		Short: shortDescription,
		Long: shortDescription + `

TARGET is of the form:
  path/to/file.go:FuncName[:N]

Where N is the 1-based line number of the function/method declaration.
If you omit N and multiple functions with the same name exist in the file,
resolution is ambiguous and the tool will ask you to disambiguate by line number.`,
		Example: `  # Target a function by name
  goctx ./internal/foo/bar.go:FuncInNeedOfContext

  # Disambiguate by 1-based line number of the declaration
  goctx ./internal/foo/bar.go:FuncInNeedOfContext:42

  # Stop propagation at another function (also supports :N)
  goctx --stop-at ./internal/foo/baz.go:FuncServingAsBoundary ./internal/foo/bar.go:FuncInNeedOfContext

  NOTE: goctx will not work unless you have a 'go.mod' file.
  That's because it uses Go internals to parse your code into packages!`,
		Version: OverallVersionStringColorized(ctx),
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logHandler := log.NewWithOptions(
				cmd.OutOrStdout(),
				log.Options{
					Level:           log.WarnLevel, // Setting this to lowest possible value, since slog will handle the actual filtering.
					ReportTimestamp: true,
					ReportCaller:    true,
				},
			)
			logger := slog.New(logHandler)
			slog.SetDefault(logger)

			slog.Debug("logger initialized")

			stopAt, err := cmd.Flags().GetString(OptNameStopAt)
			if err != nil {
				return fmt.Errorf("parsing stop-at: %w", err)
			}

			tags, err := cmd.Flags().GetString(OptNameTags)
			if err != nil {
				return fmt.Errorf("parsing tags: %w", err)
			}

			httpMode, err := cmd.Flags().GetBool(OptNameHTTP)
			if err != nil {
				return fmt.Errorf("parsing html: %w", err)
			}

			verbose, err := cmd.PersistentFlags().GetBool(OptNameVerbose)
			if err != nil {
				return fmt.Errorf("parsing verbose: %w", err)
			}

			if verbose {
				logHandler.SetLevel(log.DebugLevel)
			}

			slog.Debug(
				"flags parsed",
				slog.String("stopAt", stopAt),
				slog.String("tags", tags),
				slog.Bool("html", httpMode),
				slog.Bool("verbose", verbose),
				slog.Int("argc", len(cmd.Flags().Args())),
			)

			if len(cmd.Flags().Args()) < 1 {
				return cmd.Help()
			}

			opts := goctx.Options{
				Target:  args[0],
				StopAt:  stopAt,
				Tags:    tags,
				HTML:    httpMode,
				WorkDir: ".",
			}

			slog.Debug(
				"invoking run",
				slog.String("target", opts.Target),
				slog.String("stopAt", opts.StopAt),
				slog.Bool("html", opts.HTML),
				slog.String("workDir", opts.WorkDir),
			)
			return goctx.Run(cmd.Context(), opts)
		},
	}

	rootCmd.Flags().String(OptNameStopAt, "", "Optional terminating function path of the form path/to/file.go:FuncName[:N]")
	rootCmd.Flags().Bool(OptNameHTTP, false, "Terminate at http.HandlerFunc boundaries and derive ctx from req.Context()")
	rootCmd.Flags().StringP(OptNameTags, "t", "", "List of build tags to consider during loading (same syntax as 'go build -tags', e.g. 'tag1,tag2' or '!exclude')")
	rootCmd.PersistentFlags().BoolP(OptNameVerbose, OptNameVerboseShortHand, false, "Verbose output")

	return rootCmd
}

// ExecuteWithFang runs the root Cobra command with Fang-specific options.
// It accepts a context and a root Cobra command as input parameters.
// Returns an error if the command execution fails.
func ExecuteWithFang(ctx context.Context, rootCmd *cobra.Command) error {
	return fang.Execute(ctx, rootCmd, fang.WithVersion(rootCmd.Version), fang.WithoutManpage()) //nolint:wrapcheck // This is the top-level error emitted from cobra, so it's okay.
}
