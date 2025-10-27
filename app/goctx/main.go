package main

import (
	"context"
	"fmt"
	"os"

	"github.com/example/contextualize/cmd/goctx"
)

func main() {
	ctx := context.Background()

	rootCmd := goctx.NewRootCmd()

	rootCmd.Flags().String(goctx.OptNameStopAt, "", "Optional terminating function path of the form path/to/file.go:FuncName[:N]")
	rootCmd.Flags().Bool(goctx.OptNameHTTP, false, "Terminate at http.HandlerFunc boundaries and derive ctx from req.Context()")

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
