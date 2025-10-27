package main

// Package-level comment stays here

import (
	"fmt" // say hello
)

// DoThing does something.
// It should get a ctx parameter inserted, but comments must remain.
func DoThing() {
	fmt.Println("hi") // inline comment
}

func main() {
	DoThing()
}
