package main

import (
	"context"
	"fmt"
)

// MyOtherFunc1 needs a ctx param added
func MyOtherFunc1(ctx context.Context) { fmt.Println("one") }

// MyOtherFunc2 needs a ctx param added
func MyOtherFunc2(ctx context.Context) { fmt.Println("two") }

// MyFunc calls both; after adding ctx to both callees, MyFunc itself should have only one ctx param,
// and reuse it for both calls (not duplicate ctx params).
func MyFunc(ctx context.Context) {
	MyOtherFunc1(ctx)
	MyOtherFunc2(ctx)
}

func main() {
	ctx := context.Background()
	MyFunc(ctx)
}
