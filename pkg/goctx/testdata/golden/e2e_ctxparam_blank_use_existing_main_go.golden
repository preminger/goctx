package main

import "context"

func target(ctx context.Context) {}

func caller(ctx context.Context) {
	target(ctx) // should pass ctx after renaming from _
}

func main() {}
