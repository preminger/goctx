package main

// Package-level comment stays here

import (
	"context"
	"fmt" // say hello
)

// FuncInNeedOfContext does something.
// It should get a ctx parameter inserted, but comments must remain.
func FuncInNeedOfContext(ctx context.Context) {
	fmt.Println("hi") // inline comment
}

func main() {
	ctx := context.Background()
	FuncInNeedOfContext(ctx)
}
