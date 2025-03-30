package changroup_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/maratori/changroup"
)

func TestAckableChanGroup(t *testing.T) { //nolint:gocognit // yeah
	t.Parallel()
	t.Run("doesn't stuck if not acquired", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewAckableGroup[int]()
		assertDoesNotStuck(t, group.Send, changroup.NewAckable(1, func() {}))
		assertDoesNotStuck(t, group.Send, changroup.NewAckable(2, func() {}))
		assertDoesNotStuck(t, group.Send, changroup.NewAckable(3, func() {}))
		assertDoesNotStuck(t, group.SendAsync, changroup.NewAckable(4, func() {}))
		assertDoesNotStuck(t, group.SendAsync, changroup.NewAckable(5, func() {}))
		assertDoesNotStuck(t, group.SendAsync, changroup.NewAckable(6, func() {}))
	})
	t.Run("release closes channel", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewAckableGroup[int]()
		ch, release := group.Acquire()
		release()
		assertChanClosed(t, ch)
	})
	t.Run("release can be called twice", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewAckableGroup[int]()
		_, release := group.Acquire()
		release()
		release()
	})
	t.Run("ReleaseAll closes all channels", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewAckableGroup[int]()
		ch1, _ := group.Acquire()
		ch2, _ := group.Acquire()
		group.ReleaseAll()
		assertChanClosed(t, ch1)
		assertChanClosed(t, ch2)
	})
	t.Run("ReleaseAll can be called twice and in parallel with ReleaseFunc", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewAckableGroup[int]()
		ch1, release1 := group.Acquire()
		ch2, release2 := group.Acquire()

		done := make(chan struct{})
		for i := 0; i < 2; i++ {
			go func() {
				group.ReleaseAll()
				done <- struct{}{}
			}()
			go func() {
				release1()
				done <- struct{}{}
			}()
			go func() {
				release2()
				done <- struct{}{}
			}()
		}

		for i := 0; i < 2*3; i++ {
			waitChan(t, done)
		}

		assertChanClosed(t, ch1)
		assertChanClosed(t, ch2)
	})
	t.Run("doesn't send to released channel", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewAckableGroup[int]()
		ch1, _ := group.Acquire()
		ch2, release := group.Acquire()
		release()
		go group.Send(changroup.NewAckable(1, func() {}))
		require.Equal(t, 1, waitChan(t, ch1).Value)
		assertChanClosed(t, ch2)
		group.SendAsync(changroup.NewAckable(2, func() {}))
		require.Equal(t, 2, waitChan(t, ch1).Value)
		assertChanClosed(t, ch2)
	})
	// TODO: add more tests
	t.Run("ack happens once for SendAsync", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewAckableGroup[int]()
		ch1, _ := group.Acquire()
		ch2, _ := group.Acquire()
		done := make(chan struct{})
		group.SendAsync(changroup.NewAckable(1, func() { close(done) }))
		r1 := waitChan(t, ch1)
		r2 := waitChan(t, ch2)
		require.Equal(t, 1, r1.Value)
		require.Equal(t, 1, r2.Value)
		assertChanBlocked(t, done)
		r1.Ack()
		r1.Ack() // no effect
		r1.Ack() // no effect
		assertChanBlocked(t, done)
		r2.Ack()
		waitChan(t, done)
	})
	t.Run("ack happens once for Send", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewAckableGroup[int]()
		ch1, _ := group.Acquire()
		ch2, _ := group.Acquire()
		done := make(chan struct{})
		go group.Send(changroup.NewAckable(1, func() { close(done) }))
		r1 := waitChan(t, ch1)
		r2 := waitChan(t, ch2)
		require.Equal(t, 1, r1.Value)
		require.Equal(t, 1, r2.Value)
		assertChanBlocked(t, done)
		r1.Ack()
		r1.Ack() // no effect
		r1.Ack() // no effect
		assertChanBlocked(t, done)
		r2.Ack()
		waitChan(t, done)
	})
	t.Run("concurrency", func(t *testing.T) {
		t.Parallel()
		if testing.Short() {
			t.Skip("long test")
		}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		var mu sync.Mutex
		randDuration := func() time.Duration {
			mu.Lock()
			defer mu.Unlock()
			return time.Duration(r.Int63n(int64(time.Millisecond)))
		}
		group := changroup.NewAckableGroup[struct{}]()
		done := make(chan struct{})
		stop := make(chan struct{})
		go func() {
			defer func() { done <- struct{}{} }()
			for {
				group.Send(changroup.NewAckable(struct{}{}, func() {}))
				select {
				case <-time.After(randDuration()):
				case <-stop:
					return
				}
			}
		}()
		go func() {
			defer func() { done <- struct{}{} }()
			for {
				group.SendAsync(changroup.NewAckable(struct{}{}, func() {}))
				select {
				case <-time.After(randDuration()):
				case <-stop:
					return
				}
			}
		}()
		const n = 1000
		for i := 0; i < n; i++ {
			go func() {
				defer func() { done <- struct{}{} }()
				select {
				case <-time.After(randDuration()):
				case <-stop:
					return
				}
				ch, release := group.Acquire()
				defer release()
				for {
					select {
					case <-ch:
					case <-time.After(randDuration()):
						return
					case <-stop:
						return
					}
				}
			}()
		}
		time.Sleep(5 * time.Second)
		close(stop)
		for i := 0; i < n+2; i++ {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				require.Fail(t, "timeout", "i=%d", i)
			}
		}
	})
}
