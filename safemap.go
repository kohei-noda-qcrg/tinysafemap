package safemap

import (
	"iter"
	"sync"
)

type SafeMap[K comparable, V any] struct {
	mtx  sync.RWMutex
	data map[K]V
}

// create a SafeMap instance without size (size=0)
func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return NewSafeMapWithSize[K, V](0)
}

// create a SafeMap instance with size
func NewSafeMapWithSize[K comparable, V any](size int) *SafeMap[K, V] {
	return &SafeMap[K, V]{
		data: make(map[K]V, size),
	}
}

func (m *SafeMap[K, V]) Set(key K, value V) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.data[key] = value
}

func (m *SafeMap[K, V]) Get(key K) (V, bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	v, ok := m.data[key]
	return v, ok
}

func (m *SafeMap[K, V]) Update(key K, f func(V) V) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.data[key] = f(m.data[key])
}

func (m *SafeMap[K, V]) UpdateAndGet(key K, f func(V) V) V {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	newVal := f(m.data[key])
	m.data[key] = newVal
	return newVal
}

func (m *SafeMap[K, V]) Delete(key K) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	delete(m.data, key)
}

func (m *SafeMap[K, V]) Keys() []K {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *SafeMap[K, V]) Values() []V {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	values := make([]V, 0, len(m.data))
	for _, v := range m.data {
		values = append(values, v)
	}
	return values
}

func (m *SafeMap[K, V]) Len() int {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	return len(m.data)
}

func (m *SafeMap[K, V]) Has(key K) bool {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	_, ok := m.data[key]
	return ok
}

// erase all elements in the safemap
func (m *SafeMap[K, V]) Clear() {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	clear(m.data)
}

func (m *SafeMap[K, V]) Range(f func(key K, value V) bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	for k, v := range m.data {
		if !f(k, v) {
			break
		}
	}
}

// All returns a Go 1.23 compatible iterator (iter.Seq2).
// Note: A read-lock is held during the entire iteration. Calling write operations
// (e.g., Set, Delete, Update) on the same map inside the loop will cause a deadlock.
func (m *SafeMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.mtx.RLock()
		defer m.mtx.RUnlock()
		for k, v := range m.data {
			if !yield(k, v) {
				return
			}
		}
	}
}
