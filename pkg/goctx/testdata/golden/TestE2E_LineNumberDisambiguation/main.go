package main

import "context"

type A struct{}
type B struct{}

func (A) target(ctx context.Context) {}
func (B) target()                    {}

func main() {}
