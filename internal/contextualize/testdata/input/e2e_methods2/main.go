package main

type S struct{}

func (s *S) target() {}

func (s *S) mid1() {
	s.target()
}

func (s *S) mid2(_ context.Context) {
	s.mid1()
}

func top() {
	var s S
	s.mid1()
}

func main() {
	ctx := context.Background()
	top()
	var s S
	s.mid2(ctx)
}
