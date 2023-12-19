package types

type SimpleStack[T any] struct {
	stack []T
}

func (s *SimpleStack[T]) Top() T {
	if len(s.stack) == 0 {
		panic("Stack is empty")
	}
	return s.stack[len(s.stack)-1]
}

func (s *SimpleStack[T]) Pop() T {
	top := s.Top()
	s.stack = s.stack[:len(s.stack)-1]
	return top
}

func (s *SimpleStack[T]) Push(v T) {
	s.stack = append(s.stack, v)
}

func (s *SimpleStack[T]) Items() []T {
	return s.stack
}

func (s *SimpleStack[T]) ReplaceTop(v T) {
	if len(s.stack) == 0 {
		panic("Stack is empty")
	}
	s.stack[len(s.stack)-1] = v
}
