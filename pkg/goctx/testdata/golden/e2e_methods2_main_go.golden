package main

import "context"

type S struct{}

func (s *S) target(ctx context.Context) {}

func (s *S) mid1(ctx context.Context) {
	s.target(ctx)
}

func (s *S) mid2(ctx context.Context) {
	s.mid1(ctx)
}

func top(ctx context.Context) {
	var s S
	s.mid1(ctx)
}

func main() {
	ctx := context.Background()
	top(ctx)
	var s S
	s.mid2(ctx)
}
