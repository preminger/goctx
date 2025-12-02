package a

import "context"

// Caller calls callee.
func Caller(ctx context.Context) {
	Callee(ctx) // call down
}
