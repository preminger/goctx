package main

import (
	"context"
	"fmt"
)

// target has a blank-named context parameter that should be renamed to ctx
func target(_ context.Context) {
	fmt.Println("hi")
}

func main() {}
