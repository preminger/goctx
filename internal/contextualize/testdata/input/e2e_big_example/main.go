package main

import "context"

func target() {}

func funcOne() {
	target()
}

func funcTwo() {
	funcOne()
}

func funcThreeA(_ context.Context) {
	funcTwo()
}

func funcThreeB() {
	funcOne()
}

func funcThreeC(myWeirdlyNamedCtx context.Context) {
	funcTwo()
}

func funcFour() {
	funcThreeB()
}

func main() {
	ctx := context.Background()
	funcFour()
	funcThreeA(ctx)
	funcThreeC(ctx)
}
