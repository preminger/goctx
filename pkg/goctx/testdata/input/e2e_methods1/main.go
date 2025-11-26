package main

type S struct{}

func (s *S) target() {}

func (s *S) mid() {
	s.target()
}

func top() {
	var s S
	s.mid()
}

func main() {
	top()
}
