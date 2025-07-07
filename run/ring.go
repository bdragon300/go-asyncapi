package run

import (
	"container/list"
	"sync"
)

// NewRing returns a new Ring with zero length.
func NewRing[T any]() *Ring[T] {
	return &Ring[T]{
		mu:    &sync.Mutex{},
		iter:  nil,
		items: list.New(),
	}
}

// Ring is simple thread-safe ring container with dynamic size.
//
// It provides the thread-safe methods to iterate over the items sequentially in loop and to append/remove
// the items for O(1) time complexity.
//
// Under the hood, it uses a mutex to ensure thread safety and a linked list.
type Ring[T any] struct {
	mu    *sync.Mutex
	iter  *list.Element
	items *list.List
}

// Next cyclically iterates over the items, returning a next item. Every call moves the iterator forward,
// wrapping around to the start when the end is reached. If the list is empty, it returns false.
func (r *Ring[T]) Next() (T, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	res := r.iter
	if res == nil {
		var zero T
		return zero, false
	}

	r.iter = r.iter.Next()
	if r.iter == nil {
		r.iter = r.items.Front()
	}

	return res.Value.(T), true
}

// Append adds a new item to the end of the list and returns the element. It doesn't affect the iterator
// position, unless the list was empty before the call, in which case it sets the iterator to the new element.
func (r *Ring[T]) Append(item T) *list.Element {
	r.mu.Lock()
	defer r.mu.Unlock()

	element := r.items.PushBack(item)
	if r.iter == nil {
		r.iter = element
	}

	return element
}

// Remove removes an element from the list. If the iterator is pointing to the removed element, it will
// be moved forward cyclically. If the list is empty or element is nil, the function does nothing.
func (r *Ring[T]) Remove(element *list.Element) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if element == nil || r.items.Len() == 0 {
		return
	}

	if r.iter == element {
		r.iter = element.Next()
	}

	r.items.Remove(element)
	if r.iter == nil {
		r.iter = r.items.Front() // Iterator remains nil if the list is empty
	}
}

// Len returns the number of items in the list.
func (r *Ring[T]) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.items.Len()
}
