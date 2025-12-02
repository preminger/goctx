package main

import "context"

// HelperTarget is defined in a _test.go file and is called by other test code.
func HelperTarget(ctx context.Context) int {
	return 42
}
