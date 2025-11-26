package main

import (
	"context"
	"fmt"
	"os"

	"github.com/preminger/goctx/cmd/goctx"
)

func main() {
	os.Exit(actualMain())
}

func actualMain() int {
	ctx := context.Background()

	rootCmd := goctx.NewRootCmd(ctx)

	rootCmd.Flags().String(goctx.OptNameStopAt, "", "Optional terminating function path of the form path/to/file.go:FuncName[:N]")
	rootCmd.Flags().Bool(goctx.OptNameHTTP, false, "Terminate at http.HandlerFunc boundaries and derive ctx from req.Context()")
	rootCmd.Flags().BoolP(goctx.OptNameVerbose, goctx.OptNameVerboseShortHand, false, "Verbose output")

	if err := goctx.ExecuteWithFang(ctx, rootCmd); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
