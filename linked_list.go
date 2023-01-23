package changroup

import "sync"

// list is a doubly linked list with limited number of methods.
type list[T any] struct {
	mu    sync.RWMutex
	first *node[T]
	last  *node[T]
}

func newList[T any]() *list[T] {
	return &list[T]{
		mu:    sync.RWMutex{},
		first: nil,
		last:  nil,
	}
}

// node is an element of doubly linked list.
type node[T any] struct {
	elem T
	prev *node[T]
	next *node[T]
	list *list[T]
}

// Insert inserts new node to the end of the list.
func (l *list[T]) Insert(elem T) *node[T] {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := &node[T]{
		elem: elem,
		prev: l.last,
		next: nil,
		list: l,
	}
	if l.last == nil {
		l.first = n
	} else {
		l.last.next = n
	}
	l.last = n
	return n
}

// Delete removes node from the list.
func (n *node[T]) Delete() {
	n.list.mu.Lock()
	defer n.list.mu.Unlock()
	if n.list.first == n {
		n.list.first = n.next
	}
	if n.list.last == n {
		n.list.last = n.prev
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
	if n.prev != nil {
		n.prev.next = n.next
	}
	n.next, n.prev = nil, nil
}

// ForEach loops over all elements and calls f with each of them.
func (l *list[T]) ForEach(f func(T)) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	n := l.first
	for n != nil {
		f(n.elem)
		n = n.next
	}
}
