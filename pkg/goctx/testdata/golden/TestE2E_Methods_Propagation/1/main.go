package main

import "context"

type S struct{}

func (s *S) target(ctx context.Context) {}

func (s *S) mid(ctx context.Context) {
	s.target(ctx)
}

func top(ctx context.Context) {
	var s S
	s.mid(ctx)
}

func main() {
	ctx := context.Background()
	top(ctx)
}
