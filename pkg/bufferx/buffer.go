package bufferx

import (
	"sync"
)

type (
	Buffer[T any] struct {
		mu    sync.Mutex
		size  int
		buf   []T
		idx   int
		async bool
	}

	Option[T any] func(*Buffer[T])
)

func WithAsync[T any]() Option[T] {
	return func(b *Buffer[T]) {
		b.async = true
	}
}

func NewBuffer[T any](sz int, options ...Option[T]) *Buffer[T] {
	b := &Buffer[T]{
		size: sz,
		buf:  make([]T, sz),
		idx:  0,
	}
	for _, opt := range options {
		opt(b)
	}
	return b
}

func (b *Buffer[T]) Push(v T, f func([]T)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf[b.idx] = v
	b.idx++
	if b.idx >= b.size {
		b.Flush(f)
	}
}

func (b *Buffer[T]) Flush(f func([]T)) {
	bak := make([]T, b.idx)
	copy(bak, b.buf[0:b.idx])
	if !b.async {
		f(bak)
	} else {
		go f(bak)
	}
	b.idx = 0
}
