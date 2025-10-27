package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/contextualize/internal/contextualize"
)

// single background context instance to satisfy guideline of only one Background in our codebase.
var rootBackground = context.Background()

func main() {
	rootCmd := &cobra.Command{
		Use:   "contextualize <path/to/file.go:FuncName[:N]>",
		Short: "Propagate context.Context through Go call graphs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetContext(rootBackground)

			stopAt, _ := cmd.Flags().GetString("stop-at")
			httpMode, _ := cmd.Flags().GetBool("html")

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

	rootCmd.Flags().String("stop-at", "", "Optional terminating function path of the form path/to/file.go:FuncName[:N]")
	rootCmd.Flags().Bool("html", false, "Terminate at http.HandlerFunc boundaries and derive ctx from req.Context()")

	if err := rootCmd.ExecuteContext(rootBackground); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
