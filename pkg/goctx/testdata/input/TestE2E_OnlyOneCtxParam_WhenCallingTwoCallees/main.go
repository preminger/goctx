package main

import "fmt"

// MyOtherFunc1 needs a ctx param added
func MyOtherFunc1() { fmt.Println("one") }

// MyOtherFunc2 needs a ctx param added
func MyOtherFunc2() { fmt.Println("two") }

// MyFunc calls both; after adding ctx to both callees, MyFunc itself should have only one ctx param,
// and reuse it for both calls (not duplicate ctx params).
func MyFunc() {
	MyOtherFunc1()
	MyOtherFunc2()
}

func main() {
	MyFunc()
}
