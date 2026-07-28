package safemap

import (
	"slices"
	"sync"
	"testing"
	"time"
)

func TestSafeMap(t *testing.T) {

	t.Run("Basic Operations", func(t *testing.T) {
		m := NewSafeMap[string, int]()
		m.Set("a", 1)
		m.Set("b", 2)

		if v, ok := m.Get("a"); !ok || v != 1 {
			t.Errorf("Expected Get('a') to return 1, got %v", v)
		}

		if v, ok := m.Get("b"); !ok || v != 2 {
			t.Errorf("Expected Get('b') to return 2, got %v", v)
		}

		m.Delete("a")

		if _, ok := m.Get("a"); ok {
			t.Errorf("Expected Get('a') to return false after deletion")
		}

		if _, ok := m.Get("missing"); ok {
			t.Errorf("Expected Get('missing') to return false")
		}
	})

	t.Run("Update Logic", func(t *testing.T) {
		m := NewSafeMap[string, int]()

		m.Update("counter", func(v int) int {
			return v + 1
		})
		if v, ok := m.Get("counter"); !ok || v != 1 {
			t.Errorf("Expected counter to be initialized to 1, got %v", v)
		}

		m.Update("counter", func(v int) int {
			return v * 10
		})
		if v, _ := m.Get("counter"); v != 10 {
			t.Errorf("Expected counter to be updated to 10, got %v", v)
		}
	})

	t.Run("Atomic Update Concurrency", func(t *testing.T) {
		m := NewSafeMap[string, int]()
		var wg sync.WaitGroup
		workers := 10000

		wg.Add(workers)
		for range workers {
			go func() {
				defer wg.Done()
				m.Update("atomic_counter", func(v int) int {
					return v + 1
				})
			}()
		}
		wg.Wait()

		if v, _ := m.Get("atomic_counter"); v != workers {
			t.Errorf("Expected atomic_counter to be exactly %d, got %v", workers, v)
		}
	})

	t.Run("Mixed Concurrency Stress", func(t *testing.T) {
		m := NewSafeMap[string, int]()
		var wg sync.WaitGroup

		wg.Add(4)

		go func() { // Writer
			defer wg.Done()
			for i := range 1000 {
				m.Set("mix_key", i)
			}
		}()

		go func() { // Reader
			defer wg.Done()
			for range 1000 {
				m.Get("mix_key")
			}
		}()

		go func() { // Updater
			defer wg.Done()
			for range 1000 {
				m.Update("mix_key", func(v int) int { return v + 1 })
			}
		}()

		go func() { // Deleter
			defer wg.Done()
			for range 1000 {
				m.Delete("mix_key")
			}
		}()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// ok
		case <-time.After(3 * time.Second):
			t.Fatal("Test timed out, possible deadlock or race condition")
		}
	})

	t.Run("UpdateAndGet Logic", func(t *testing.T) {
		m := NewSafeMap[string, int]()

		got := m.UpdateAndGet("counter", func(v int) int {
			return v + 1
		})
		if got != 1 {
			t.Errorf("Expected UpdateAndGet to return 1, got %v", got)
		}

		if v, _ := m.Get("counter"); v != 1 {
			t.Errorf("Expected map value to be 1, got %v", v)
		}

		got = m.UpdateAndGet("counter", func(v int) int {
			return v * 10
		})
		if got != 10 {
			t.Errorf("Expected UpdateAndGet to return 10, got %v", got)
		}

		if v, _ := m.Get("counter"); v != 10 {
			t.Errorf("Expected map value to be 10, got %v", v)
		}
	})

	t.Run("Len and Has and Clear", func(t *testing.T) {
		m := NewSafeMap[string, int]()
		if m.Len() != 0 {
			t.Errorf("Expected Len to be 0, got %d", m.Len())
		}

		m.Set("a", 1)
		m.Set("b", 2)
		if m.Len() != 2 {
			t.Errorf("Expected Len to be 2, got %d", m.Len())
		}

		if !m.Has("a") {
			t.Error("Expected Has('a') to be true")
		}
		if m.Has("c") {
			t.Error("Expected Has('c') to be false")
		}

		m.Clear()
		if m.Len() != 0 {
			t.Errorf("Expected Len after Clear to be 0, got %d", m.Len())
		}
		if m.Has("a") {
			t.Error("Expected Has('a') to be false after Clear")
		}
	})

	t.Run("Range Callback", func(t *testing.T) {
		m := NewSafeMap[string, int]()
		m.Set("a", 1)
		m.Set("b", 2)

		keys := make([]string, 0)
		m.Range(func(k string, v int) bool {
			keys = append(keys, k)
			return true
		})

		if len(keys) != 2 || !slices.Contains(keys, "a") || !slices.Contains(keys, "b") {
			t.Errorf("Unexpected keys from Range: %v", keys)
		}

		// Break early test
		count := 0
		m.Range(func(k string, v int) bool {
			count++
			return false // break
		})
		if count != 1 {
			t.Errorf("Expected Range to break after 1 iteration, got %d", count)
		}
	})

	t.Run("All Iterator (Go 1.23)", func(t *testing.T) {
		m := NewSafeMap[string, int]()
		m.Set("a", 10)
		m.Set("b", 20)

		sum := 0
		keys := make([]string, 0)
		for k, v := range m.All() {
			keys = append(keys, k)
			sum += v
		}

		if sum != 30 {
			t.Errorf("Expected sum to be 30, got %d", sum)
		}
		if len(keys) != 2 || !slices.Contains(keys, "a") || !slices.Contains(keys, "b") {
			t.Errorf("Unexpected keys from All(): %v", keys)
		}
	})

	t.Run("Keys and Values", func(t *testing.T) {
		m := NewSafeMap[string, int]()
		m.Set("a", 1)
		m.Set("b", 2)

		keys := m.Keys()
		if len(keys) != 2 || !slices.Contains(keys, "a") || !slices.Contains(keys, "b") {
			t.Errorf("Unexpected Keys(): %v", keys)
		}

		values := m.Values()
		if len(values) != 2 || !slices.Contains(values, 1) || !slices.Contains(values, 2) {
			t.Errorf("Unexpected Values(): %v", values)
		}
	})

	t.Run("NewSafeMapWithSize and Negative Size", func(t *testing.T) {
		// make(map, negative_size) does not panic; it is normalized to 0.
		// Verify that it is initialized successfully and works as expected.
		m := NewSafeMapWithSize[string, int](-1)
		m.Set("a", 1)
		if m.Len() != 1 {
			t.Errorf("Expected Len to be 1, got %d", m.Len())
		}

		if v, ok := m.Get("a"); !ok || v != 1 {
			t.Errorf("Expected Get('a') to return 1, got %v", v)
		}
	})
}
