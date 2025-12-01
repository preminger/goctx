package main

import (
	"fmt"

	"example.com/e2e/xyz"
)

type TypeA struct{}

func (TypeA) Func() { fmt.Println("TypeA.Func") }

type TypeB struct{ Func int }

func MyFunc() { fmt.Println("MyFunc") }

func main() {
	// unqualified function in current package
	MyFunc()

	// qualified function from another package
	xyz.MyFunc()

	// method call on a value
	var a TypeA
	a.Func()

	// field access, not a call
	var b TypeB
	_ = b.Func
}
