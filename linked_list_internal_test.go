package changroup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	t.Parallel()
	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		l := newList[int]()
		assertForEach(t, l, nil)
	})
	t.Run("append 1 element", func(t *testing.T) {
		t.Parallel()
		l := newList[int]()

		a := l.Append(10)

		assertForEach(t, l, []int{10})
		assert.Same(t, a, l.first)
		assert.Same(t, a, l.last)
		assert.Nil(t, a.prev)
		assert.Nil(t, a.next)
	})
	t.Run("delete 1 element", func(t *testing.T) {
		t.Parallel()
		l := newList[int]()
		a := l.Append(10)

		a.Delete()

		assertForEach(t, l, nil)
		assert.Nil(t, l.first)
		assert.Nil(t, l.last)
		assert.Nil(t, a.prev)
		assert.Nil(t, a.next)
	})
	t.Run("append 2 elements", func(t *testing.T) {
		t.Parallel()
		l := newList[int]()

		a := l.Append(10)
		b := l.Append(20)

		assertForEach(t, l, []int{10, 20})
		assert.Same(t, a, l.first)
		assert.Same(t, b, l.last)
		assert.Nil(t, a.prev)
		assert.Same(t, b, a.next)
		assert.Same(t, a, b.prev)
		assert.Nil(t, b.next)
	})
	t.Run("delete last from 2 elements", func(t *testing.T) {
		t.Parallel()
		l := newList[int]()
		a := l.Append(10)
		b := l.Append(20)

		b.Delete()

		assertForEach(t, l, []int{10})
		assert.Same(t, a, l.first)
		assert.Same(t, a, l.last)
		assert.Nil(t, a.prev)
		assert.Nil(t, a.next)
		assert.Nil(t, b.prev)
		assert.Nil(t, b.next)
	})
	t.Run("delete first from 2 elements", func(t *testing.T) {
		t.Parallel()
		l := newList[int]()
		a := l.Append(10)
		b := l.Append(20)

		a.Delete()

		assertForEach(t, l, []int{20})
		assert.Same(t, b, l.first)
		assert.Same(t, b, l.last)
		assert.Nil(t, a.prev)
		assert.Nil(t, a.next)
		assert.Nil(t, b.prev)
		assert.Nil(t, b.next)
	})
	t.Run("append 3 elements", func(t *testing.T) {
		t.Parallel()
		l := newList[int]()

		a := l.Append(10)
		b := l.Append(20)
		c := l.Append(30)

		assertForEach(t, l, []int{10, 20, 30})
		assert.Same(t, a, l.first)
		assert.Same(t, c, l.last)
		assert.Nil(t, a.prev)
		assert.Same(t, b, a.next)
		assert.Same(t, a, b.prev)
		assert.Same(t, c, b.next)
		assert.Same(t, b, c.prev)
		assert.Nil(t, c.next)
	})
	t.Run("delete last from 3 elements", func(t *testing.T) {
		t.Parallel()
		l := newList[int]()
		a := l.Append(10)
		b := l.Append(20)
		c := l.Append(30)

		c.Delete()

		assertForEach(t, l, []int{10, 20})
		assert.Same(t, a, l.first)
		assert.Same(t, b, l.last)
		assert.Nil(t, a.prev)
		assert.Same(t, b, a.next)
		assert.Same(t, a, b.prev)
		assert.Nil(t, b.next)
		assert.Nil(t, c.prev)
		assert.Nil(t, c.next)
	})
	t.Run("delete first from 3 elements", func(t *testing.T) {
		t.Parallel()
		l := newList[int]()
		a := l.Append(10)
		b := l.Append(20)
		c := l.Append(30)

		a.Delete()

		assertForEach(t, l, []int{20, 30})
		assert.Same(t, b, l.first)
		assert.Same(t, c, l.last)
		assert.Nil(t, a.prev)
		assert.Nil(t, a.next)
		assert.Nil(t, b.prev)
		assert.Same(t, c, b.next)
		assert.Same(t, b, c.prev)
		assert.Nil(t, c.next)
	})
	t.Run("delete middle from 3 elements", func(t *testing.T) {
		t.Parallel()
		l := newList[int]()
		a := l.Append(10)
		b := l.Append(20)
		c := l.Append(30)

		b.Delete()

		assertForEach(t, l, []int{10, 30})
		assert.Same(t, a, l.first)
		assert.Same(t, c, l.last)
		assert.Nil(t, a.prev)
		assert.Same(t, c, a.next)
		assert.Nil(t, b.prev)
		assert.Nil(t, b.next)
		assert.Same(t, a, c.prev)
		assert.Nil(t, c.next)
	})
}

func assertForEach[T any](t *testing.T, l *list[T], expected []T) {
	var actual []T
	l.ForEach(func(v T) {
		actual = append(actual, v)
	})
	assert.Equal(t, expected, actual)
}
