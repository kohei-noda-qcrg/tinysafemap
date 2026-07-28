# tinysafemap

`tinysafemap` is a lightweight, thread-safe generic map implementation designed for Go 1.23+.

It is ideal for projects that want a zero-dependency, simple Mutex-wrapped map interface without the type-casting complexity of `sync.Map`.

## Features

- **Generics**: Supports Go generics, ensuring compile-time type validation for both keys and values.
- **Atomic Updates (`Update`/`UpdateAndGet`)**: Provides thread-safe, functional updates to prevent race conditions during read-modify-write operations.
- **For Range Iterator Support (`All()`)**: Implements the `iter.Seq2` signature, allowing clean `for range` loops over the concurrent map.
- **Other Features**: Includes helper methods such as `Len()`, `Has()`, `Clear()`, and slice conversions (`Keys()`/`Values()`).
- **Zero Dependencies**: Implemented purely with the Go standard library.

## Usage

### Basic Operations and Initialization

You can import the package using `import "github.com/kohei-noda-qcrg/tinysafemap"`.

```go
package main

import (
	"fmt"
	"github.com/kohei-noda-qcrg/tinysafemap"
)

func main() {
	// Initialize with default capacity
	m := safemap.NewSafeMap[string, int]()

	// Initialize with custom capacity to minimize memory allocations
	// m := safemap.NewSafeMapWithSize[string, int](100)

	// Set values
	m.Set("apple", 100)
	m.Set("banana", 200)

	// Get values
	if price, ok := m.Get("apple"); ok {
		fmt.Printf("Apple price: %d\n", price)
	}

	// Check existence
	if m.Has("banana") {
		fmt.Println("Banana exists!")
	}

	// Get size
	fmt.Printf("Total items: %d\n", m.Len()) // 2
}
```

### Atomic Updates (`Update` / `UpdateAndGet`)

Perform atomic modifications (read-modify-write) in a single locked transaction. This is useful for concurrent counters or state updates.

```go
// Atomically increment the value
m.Update("counter", func(v int) int {
	return v + 1
})

// Atomically update and return the new value
newVal := m.UpdateAndGet("counter", func(v int) int {
	return v * 10
})
```

### Map Iteration

Using the `All()` iterator, you can iterate over the map safely using standard `for range` loops. A read-lock (`RLock`) is automatically held during the entire iteration.

> [!CAUTION]
> **Avoid modifying the map inside the loop.**
> Since the read-lock is held during the iteration, calling any write operations (such as `Set`, `Delete`, or `Update`) on the same map within the loop will cause a deadlock.
> If you need to modify the map during iteration, collect the keys first using `Keys()` and iterate over that slice instead.

```go
for key, val := range m.All() {
	fmt.Printf("%s: %d\n", key, val)
}
```
