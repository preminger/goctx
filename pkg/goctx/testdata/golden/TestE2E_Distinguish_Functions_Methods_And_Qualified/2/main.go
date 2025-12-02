package main

import (
	"context"

	"example.com/e2e/xyz"
	"fmt"
)

type TypeA struct{}

func (_ *TypeA) MyFunc() int {
	fmt.Println("TypeA.Func")

	return -1
}

type TypeB struct{}

func (_ TypeB) MyFunc(ctx context.Context) int {
	fmt.Println("TypeB.Func")

	return -2
}

func MyFunc() {
	fmt.Println("MyFunc")
}

func main() {
	// unqualified function in current package
	ctx := context.Background()
	MyFunc()

	// qualified function from another package
	xyz.MyFunc()

	// method call on a value
	var a TypeA
	a.MyFunc()

	// field access, not a call
	var b TypeB
	_ = b.MyFunc(ctx)
}
