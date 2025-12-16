package main

import (
	"context"
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
		return 1
	}

	return 0
}
