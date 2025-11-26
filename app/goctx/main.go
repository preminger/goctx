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

	if err := goctx.ExecuteWithFang(ctx, rootCmd); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
