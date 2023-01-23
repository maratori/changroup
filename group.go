package changroup

import (
	"sync"
)

// ReleaseFunc is called to remove channel from group and close it.
type ReleaseFunc func()

type channel[T any] struct {
	ch      chan T
	done    chan struct{}
	send    sync.WaitGroup
	release ReleaseFunc
}

// Group provides pub-sub model working with channels.
//
// Each acquired channel will receive a copy of a value provided to [Group.Send].
type Group[T any] struct {
	channels *list[*channel[T]]
}

func NewGroup[T any]() *Group[T] {
	return &Group[T]{
		channels: newList[*channel[T]](),
	}
}

// ReleaseAll releases all acquired channels and closes them.
// It's safe to call [Group.ReleaseAll] several times as well as in parallel with [ReleaseFunc].
func (g *Group[T]) ReleaseAll() {
	var all []*channel[T]
	g.channels.ForEach(func(ch *channel[T]) {
		all = append(all, ch)
	})
	for _, ch := range all {
		ch.release() // there will be deadlock if call it inside ForEach.
	}
}

// Acquire creates new channel and adds it to group.
//
// [ReleaseFunc] is returned as the second value.
// It should be called to remove the channel from the group and close it.
// It's safe to call [ReleaseFunc] several times as well as in parallel with [Group.ReleaseAll].
func (g *Group[T]) Acquire() (<-chan T, ReleaseFunc) {
	ch := g.channels.Insert(&channel[T]{
		ch:      make(chan T),
		done:    make(chan struct{}),
		send:    sync.WaitGroup{},
		release: nil, // is filled below
	})

	once := sync.Once{}
	ch.elem.release = func() {
		once.Do(func() {
			ch.Delete()
			close(ch.elem.done)
			ch.elem.send.Wait()
			close(ch.elem.ch)
		})
	}

	return ch.elem.ch, ch.elem.release
}

// Send sends a value to each acquired channel.
//
// It guarantees that all channels receive the values in the same order.
// And that the order is the same as [Group.Send] calls.
//
// It waits for all channels to receive the value or to be released.
func (g *Group[T]) Send(value T) {
	wg := sync.WaitGroup{}
	g.channels.ForEach(func(ch *channel[T]) {
		// select is an optimisation to not create goroutine if someone reads the channel (should cover 90% cases)
		select {
		case ch.ch <- value:
		default:
			wg.Add(1)
			ch.send.Add(1)
			go func() {
				defer wg.Done()
				defer ch.send.Done()
				select {
				case ch.ch <- value:
				case <-ch.done:
				}
			}()
		}
	})
	wg.Wait()
}

// SendAsync sends a value to each acquired channel, but unlike [Group.Send] doesn't block.
// Also, it doesn't preserve the order of values!
func (g *Group[T]) SendAsync(value T) {
	g.channels.ForEach(func(ch *channel[T]) {
		// select is an optimisation to not create goroutine if someone reads the channel (should cover 90% cases)
		select {
		case ch.ch <- value:
		default:
			ch.send.Add(1)
			go func() {
				defer ch.send.Done()
				select {
				case ch.ch <- value:
				case <-ch.done:
				}
			}()
		}
	})
}
