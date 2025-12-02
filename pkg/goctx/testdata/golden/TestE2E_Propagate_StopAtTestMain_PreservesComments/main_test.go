package main

import (
	"context"

	"example.com/e2e/a"
	"testing"
)

// TestMain is the boundary; comment should stay here.
func TestMain(m *testing.M) {
	ctx := context.Background()
	a.Caller(ctx)
}
