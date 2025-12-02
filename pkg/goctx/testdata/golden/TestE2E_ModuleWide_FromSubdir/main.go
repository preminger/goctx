package main

import (
	"context"
	"example.com/e2e/sub"
)

func main() {
	ctx := context.Background()
	sub.FuncInNeedOfContext(ctx)
}
