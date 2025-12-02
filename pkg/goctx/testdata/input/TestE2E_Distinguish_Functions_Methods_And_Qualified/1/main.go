package main

import (
	"fmt"

	"example.com/e2e/xyz"
)

type TypeA struct{}

func (_ *TypeA) MyFunc() int {
	fmt.Println("TypeA.Func")

	return -1
}

type TypeB struct{}

func (_ TypeB) MyFunc() int {
	fmt.Println("TypeB.Func")

	return -2
}

func MyFunc() {
	fmt.Println("MyFunc")
}

func main() {
	// unqualified function in current package
	MyFunc()

	// qualified function from another package
	xyz.MyFunc()

	// method call on a value
	var a TypeA
	a.MyFunc()

	// field access, not a call
	var b TypeB
	_ = b.MyFunc()
}
