package main

import "context"

func target() {}

func caller(myContext context.Context) {
	target() // should pass myContext, not add a new param
}

func main() {
	caller(context.Background())
}
