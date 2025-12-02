package main

import "context"

func targetFunc() {}

func funcOne() {
	targetFunc()
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
