package main

import "context"

func target(ctx context.Context) {}

func caller(myContext context.Context) {
	target(myContext) // should pass myContext, not add a new param
}

func main() {
	caller(context.Background())
}
