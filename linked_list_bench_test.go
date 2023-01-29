package changroup

import (
	std "container/list"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type S struct {
	mu sync.Mutex
	l  *std.List
}

func NewS() *S {
	return &S{
		mu: sync.Mutex{},
		l:  std.New(),
	}
}

func BenchmarkInsertInt(b *testing.B) {
	if testing.Short() {
		benchmarkInsertGInt(b)
	} else {
		benchmarkInsertSInt(b)
	}
}

func benchmarkInsertGInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := newList[int]()
		for j := 0; j < 10000; j++ {
			l.Insert(j)
		}
	}
}

func benchmarkInsertSInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := NewS()
		for j := 0; j < 10000; j++ {
			l.mu.Lock()
			l.l.PushBack(j)
			l.mu.Unlock()
		}
	}
}

func BenchmarkInsertPtrStruct(b *testing.B) {
	if testing.Short() {
		benchmarkInsertGPtrStruct(b)
	} else {
		benchmarkInsertSPtrStruct(b)
	}
}

func benchmarkInsertGPtrStruct(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := newList[*channel[error]]()
		for j := 0; j < 10000; j++ {
			l.Insert(&channel[error]{
				ch:      make(chan error),
				done:    make(chan struct{}),
				send:    sync.WaitGroup{},
				release: func() {},
			})
		}
	}
}

func benchmarkInsertSPtrStruct(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := NewS()
		for j := 0; j < 10000; j++ {
			l.mu.Lock()
			l.l.PushBack(&channel[error]{
				ch:      make(chan error),
				done:    make(chan struct{}),
				send:    sync.WaitGroup{},
				release: func() {},
			})
			l.mu.Unlock()
		}
	}
}

func BenchmarkForEachInt(b *testing.B) {
	if testing.Short() {
		benchmarkForEachGInt(b)
	} else {
		benchmarkForEachSInt(b)
	}
}

var xxx int

func benchmarkForEachGInt(b *testing.B) {
	l := newList[int]()
	for j := 0; j < 10000; j++ {
		l.Insert(j)
	}
	var x int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.ForEach(func(i int) {
			x = i
		})
	}
	xxx = x
}

func benchmarkForEachSInt(b *testing.B) {
	l := NewS()
	for j := 0; j < 10000; j++ {
		l.mu.Lock()
		l.l.PushBack(j)
		l.mu.Unlock()
	}
	var x int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.mu.Lock()
		for e := l.l.Front(); e != nil; e = e.Next() {
			x = e.Value.(int)
		}
		l.mu.Unlock()
	}
	xxx = x
}

func BenchmarkForEachPtrStruct(b *testing.B) {
	if testing.Short() {
		benchmarkForEachGPtrStruct(b)
	} else {
		benchmarkForEachSPtrStruct(b)
	}
}

var yyy *channel[error]

func benchmarkForEachGPtrStruct(b *testing.B) {
	l := newList[*channel[error]]()
	for j := 0; j < 10000; j++ {
		l.Insert(&channel[error]{
			ch:      make(chan error),
			done:    make(chan struct{}),
			send:    sync.WaitGroup{},
			release: func() {},
		})
	}
	var x *channel[error]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.ForEach(func(i *channel[error]) {
			x = i
		})
	}
	yyy = x
}

func benchmarkForEachSPtrStruct(b *testing.B) {
	l := NewS()
	for j := 0; j < 10000; j++ {
		l.mu.Lock()
		l.l.PushBack(&channel[error]{
			ch:      make(chan error),
			done:    make(chan struct{}),
			send:    sync.WaitGroup{},
			release: func() {},
		})
		l.mu.Unlock()
	}
	var x *channel[error]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.mu.Lock()
		for e := l.l.Front(); e != nil; e = e.Next() {
			x = e.Value.(*channel[error])
		}
		l.mu.Unlock()
	}
	yyy = x
}

func BenchmarkDeleteInt(b *testing.B) {
	if testing.Short() {
		benchmarkDeleteGInt(b)
	} else {
		benchmarkDeleteSInt(b)
	}
}

func benchmarkDeleteGInt(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	l := newList[int]()
	d := make([]*node[int], 0, 1000)
	for j := 0; j < 10000; j++ {
		n := l.Insert(j)
		if r.Float32() < 0.1 {
			d = append(d, n)
		}
	}
	r.Shuffle(len(d), func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, n := range d {
			n.Delete()
		}
	}
}

func benchmarkDeleteSInt(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	l := NewS()
	d := make([]*std.Element, 0, 1000)
	for j := 0; j < 10000; j++ {
		l.mu.Lock()
		e := l.l.PushBack(j)
		l.mu.Unlock()
		if r.Float32() < 0.1 {
			d = append(d, e)
		}
	}
	r.Shuffle(len(d), func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, n := range d {
			l.mu.Lock()
			l.l.Remove(n)
			l.mu.Unlock()
		}
	}
}

func BenchmarkDeletePtrStruct(b *testing.B) {
	if testing.Short() {
		benchmarkDeleteGPtrStruct(b)
	} else {
		benchmarkDeleteSPtrStruct(b)
	}
}

func benchmarkDeleteGPtrStruct(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	l := newList[*channel[error]]()
	d := make([]*node[*channel[error]], 0, 1000)
	for j := 0; j < 10000; j++ {
		n := l.Insert(&channel[error]{
			ch:      make(chan error),
			done:    make(chan struct{}),
			send:    sync.WaitGroup{},
			release: func() {},
		})
		if r.Float32() < 0.1 {
			d = append(d, n)
		}
	}
	r.Shuffle(len(d), func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, n := range d {
			n.Delete()
		}
	}
}

func benchmarkDeleteSPtrStruct(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	l := NewS()
	d := make([]*std.Element, 0, 1000)
	for j := 0; j < 10000; j++ {
		l.mu.Lock()
		e := l.l.PushBack(&channel[error]{
			ch:      make(chan error),
			done:    make(chan struct{}),
			send:    sync.WaitGroup{},
			release: func() {},
		})
		l.mu.Unlock()
		if r.Float32() < 0.1 {
			d = append(d, e)
		}
	}
	r.Shuffle(len(d), func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, n := range d {
			l.mu.Lock()
			l.l.Remove(n)
			l.mu.Unlock()
		}
	}
}
