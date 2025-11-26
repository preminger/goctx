package main

import "context"

// target has no ctx param; its immediate caller has a blank-named context param
// at a higher level, a caller already passes context; we should not modify higher callers
func target(ctx context.Context) {}

func mid(ctx context.Context) {
	target(ctx) // mid should rename '_' to ctx and pass it here
}

func high(myContext context.Context) {
	mid(myContext) // high already passes a context; should remain untouched
}

func main() {
	high(context.Background())
}
