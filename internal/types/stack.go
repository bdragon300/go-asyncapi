package types

// SimpleStack is a simple stack implementation, representing the LIFO data structure.
type SimpleStack[T any] struct {
	stack []T
}

// Top returns the top element of the stack. Panics if the stack is empty.
func (s *SimpleStack[T]) Top() T {
	if len(s.stack) == 0 {
		panic("Stack is empty")
	}
	return s.stack[len(s.stack)-1]
}

// Pop removes the top element of the stack and returns it. Panics if the stack is empty.
func (s *SimpleStack[T]) Pop() T {
	top := s.Top()
	s.stack = s.stack[:len(s.stack)-1]
	return top
}

// Push adds a new element to the top of the stack.
func (s *SimpleStack[T]) Push(v T) {
	s.stack = append(s.stack, v)
}

// Items returns all elements of the stack starting from the bottom.
func (s *SimpleStack[T]) Items() []T {
	return s.stack
}

// ReplaceTop replaces the top element with the given value. Panics if the stack is empty.
//
// This is an equivalent for:
//
//	stack.Pop()
//	stack.Push(v)
func (s *SimpleStack[T]) ReplaceTop(v T) {
	if len(s.stack) == 0 {
		panic("Stack is empty")
	}
	s.stack[len(s.stack)-1] = v
}
