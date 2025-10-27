package main

import "context"

func target() {}

func caller(_ context.Context) {
	target() // should pass ctx after renaming from _
}

func main() {}
