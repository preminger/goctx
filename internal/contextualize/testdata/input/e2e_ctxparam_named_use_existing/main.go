package main

import "context"

func target(_ context.Context) {}

func caller(myContext context.Context) {
	target() // should pass myContext, not add a new param
}

func main() {}
