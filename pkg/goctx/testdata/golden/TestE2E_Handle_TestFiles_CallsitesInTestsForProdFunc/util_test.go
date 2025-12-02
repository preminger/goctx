package main

import "testing"

func TestProdFunc(t *testing.T) {
	ctx := t.Context()
	_ = ProdFunc(ctx)
}
