package main

import (
	"testing"

	"example.com/e2e/a"
)

// TestMain is the boundary; comment should stay here.
func TestMain(m *testing.M) {
	a.Caller()
}
