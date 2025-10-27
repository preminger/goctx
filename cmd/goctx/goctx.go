package goctx

import (
	"fmt"

	"github.com/example/contextualize/internal/contextualize"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "contextualize TARGET",
		Short: "Propagate context.Context through Go call graphs",
		Long: `Propagate context.Context through Go call graphs.

TARGET is of the form:
  path/to/file.go:FuncName[:N]

Where N is the 1-based line number of the function/method declaration.
If you omit N and multiple functions with the same name exist in the file,
resolution is ambiguous and the tool will ask you to disambiguate by line number.`,
		Example: `  # Target a function by name
  contextualize ./pkg/foo.go:DoThing

  # Disambiguate by 1-based line number of the declaration
  contextualize ./pkg/foo.go:DoThing:42

  # Stop propagation at another function (also supports :N)
  contextualize --stop-at ./pkg/stop.go:Boundary ./pkg/foo.go:DoThing`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			stopAt, err := cmd.Flags().GetString(OptNameStopAt)
			if err != nil {
				return fmt.Errorf("parsing stop-at: %w", err)
			}

			httpMode, err := cmd.Flags().GetBool(OptNameHTTP)
			if err != nil {
				return fmt.Errorf("parsing html: %w", err)
			}

			opts := contextualize.Options{
				Target:  args[0],
				StopAt:  stopAt,
				HTML:    httpMode,
				WorkDir: ".",
			}
			if err := contextualize.Run(cmd.Context(), opts); err != nil {
				return fmt.Errorf("running contextualize: %w", err)
			}
			return nil
		},
	}

	return rootCmd
}
