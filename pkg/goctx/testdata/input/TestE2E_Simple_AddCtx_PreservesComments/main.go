package main

// Package-level comment stays here

import (
	"fmt" // say hello
)

// FuncInNeedOfContext does something.
// It should get a ctx parameter inserted, but comments must remain.
func FuncInNeedOfContext() {
	fmt.Println("hi") // inline comment
}

func main() {
	FuncInNeedOfContext()
}
