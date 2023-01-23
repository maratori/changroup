package changroup_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/maratori/changroup"
	"github.com/stretchr/testify/require"
)

func TestGroup(t *testing.T) { //nolint:gocognit // yeah
	t.Parallel()
	t.Run("doesn't stuck if not acquired", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewGroup[int]()
		assertDoesNotStuck(t, group.Send, 1)
		assertDoesNotStuck(t, group.Send, 2)
		assertDoesNotStuck(t, group.Send, 3)
		assertDoesNotStuck(t, group.SendAsync, 4)
		assertDoesNotStuck(t, group.SendAsync, 5)
		assertDoesNotStuck(t, group.SendAsync, 6)
	})
	t.Run("release closes channel", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewGroup[int]()
		ch, release := group.Acquire()
		release()
		assertChanClosed(t, ch)
	})
	t.Run("release can be called twice", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewGroup[int]()
		_, release := group.Acquire()
		release()
		release()
	})
	t.Run("ReleaseAll closes all channels", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewGroup[int]()
		ch1, _ := group.Acquire()
		ch2, _ := group.Acquire()
		group.ReleaseAll()
		assertChanClosed(t, ch1)
		assertChanClosed(t, ch2)
	})
	t.Run("ReleaseAll can be called twice and in parallel with ReleaseFunc", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewGroup[int]()
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
		group := changroup.NewGroup[int]()
		ch1, _ := group.Acquire()
		ch2, release := group.Acquire()
		release()
		go group.Send(1)
		require.Equal(t, 1, waitChan(t, ch1))
		assertChanClosed(t, ch2)
		group.SendAsync(2)
		require.Equal(t, 2, waitChan(t, ch1))
		assertChanClosed(t, ch2)
	})
	t.Run("Send blocks", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewGroup[int]()
		ch, _ := group.Acquire()
		done1 := make(chan struct{})
		done2 := make(chan struct{})
		go func() {
			defer close(done1)
			group.Send(1)
		}()
		go func() {
			defer close(done2)
			waitChan(t, done1)
			group.Send(2)
		}()
		time.Sleep(time.Second)
		assertChanBlocked(t, done1)
		assertChanBlocked(t, done2)
		require.Equal(t, 1, waitChan(t, ch))
		waitChan(t, done1)
		assertChanBlocked(t, done2)
		require.Equal(t, 2, waitChan(t, ch))
		waitChan(t, done2)
	})
	t.Run("SendAsync doesn't block", func(t *testing.T) {
		t.Parallel()
		group := changroup.NewGroup[int]()
		ch1, _ := group.Acquire()
		ch2, _ := group.Acquire()
		assertDoesNotStuck(t, group.SendAsync, 1)
		assertDoesNotStuck(t, group.SendAsync, 2)
		assertDoesNotStuck(t, group.SendAsync, 3)
		// ch1 is not blocked even if no one is reading ch2
		require.ElementsMatch(t, []int{1, 2, 3}, []int{waitChan(t, ch1), waitChan(t, ch1), waitChan(t, ch1)})
		require.ElementsMatch(t, []int{1, 2, 3}, []int{waitChan(t, ch2), waitChan(t, ch2), waitChan(t, ch2)})
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
		group := changroup.NewGroup[struct{}]()
		done := make(chan struct{})
		stop := make(chan struct{})
		go func() {
			defer func() { done <- struct{}{} }()
			for {
				group.Send(struct{}{})
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
				group.SendAsync(struct{}{})
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

func waitChan[T any](t *testing.T, ch <-chan T) T {
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		require.Fail(t, "timeout")
		panic(nil)
	}
}

func assertChanClosed[T any](t *testing.T, ch <-chan T) {
	select {
	case v, ok := <-ch:
		if ok {
			require.Failf(t, "chan not closed", "%+v", v)
		}
	default:
		require.Fail(t, "chan blocked")
	}
}

func assertChanBlocked[T any](t *testing.T, ch <-chan T) {
	select {
	case v, ok := <-ch:
		if ok {
			require.Failf(t, "chan not blocked", "%+v", v)
		}
		require.Fail(t, "chan closed")
	default:
	}
}

func assertDoesNotStuck[T any](t *testing.T, fn func(T), value T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn(value)
	}()
	waitChan(t, done)
}
