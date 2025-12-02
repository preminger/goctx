package main

import "context"

// ProdFunc is a production function used by tests.
func ProdFunc(ctx context.Context) int {
	return 7
}
