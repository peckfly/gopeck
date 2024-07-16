package collection

import "sync"

type (
	I64Type interface {
		int64 | string | int | int32
	}
	I64Map[T I64Type] struct {
		sync.RWMutex `json:"-"`
		m            map[T]int64
	}
)

func NewI64Map[T I64Type]() I64Map[T] {
	return I64Map[T]{
		m: make(map[T]int64),
	}
}

func (m *I64Map[T]) Increment(key T) {
	m.Lock()
	defer m.Unlock()
	m.m[key]++
}

func (m *I64Map[T]) Get(key T) int64 {
	m.RLock()
	defer m.RUnlock()
	return m.m[key]
}

func (m *I64Map[T]) Range(f func(key T, val int64)) {
	m.RLock()
	defer m.RUnlock()
	for k, v := range m.m {
		f(k, v)
	}
}

func (m *I64Map[T]) Map() map[T]int64 {
	m.RLock()
	defer m.RUnlock()
	return m.m
}
