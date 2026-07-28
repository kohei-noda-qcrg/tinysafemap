package safemap

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
)

// read-only performance under parallel workload (100% Reads).
// RWMutex should scale well here as RLock does not block other RLocks.
func BenchmarkSafeMap_Get_Parallel(b *testing.B) {
	m := NewSafeMap[string, int]()
	m.Set("key", 100)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = m.Get("key")
		}
	})
}

// write-only performance under parallel workload (100% Writes).
// This is expected to be slow compared with read-only tests due to severe lock contention on a single RWMutex.
func BenchmarkSafeMap_Set_Parallel(b *testing.B) {
	m := NewSafeMap[string, int]()

	var id int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			val := atomic.AddInt64(&id, 1)
			m.Set(strconv.FormatInt(val, 10), int(val))
		}
	})
}

// mixed workload (90% Reads, 10% Writes).
// Represents a realistic server cache workload.
func BenchmarkSafeMap_Mixed_Parallel(b *testing.B) {
	m := NewSafeMap[string, int]()
	m.Set("key", 100)

	var id int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var count int
		for pb.Next() {
			count++
			if count%10 == 0 {
				// 10% Writes
				val := atomic.AddInt64(&id, 1)
				m.Set(strconv.FormatInt(val, 10), int(val))
			} else {
				// 90% Reads
				_, _ = m.Get("key")
			}
		}
	})
}

// It compares memory allocations and speed between Go 1.23 iterator All() and Keys() slice copy.
func BenchmarkIteration(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		m := NewSafeMapWithSize[string, int](size)
		for i := range size {
			m.Set(strconv.Itoa(i), i)
		}

		b.Run(fmt.Sprintf("All/Size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				for k, v := range m.All() {
					_, _ = k, v
				}
			}
		})

		b.Run(fmt.Sprintf("Keys/Size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				keys := m.Keys()
				for _, k := range keys {
					_ = k
				}
			}
		})
	}
}
