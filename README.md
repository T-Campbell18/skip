# skipmap

[![Build Status](https://img.shields.io/github/actions/workflow/status/T-Campbell18/skip/go.yml?branch=main)](https://github.com/T-Campbell18/skip/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/T-Campbell18/skip)](https://goreportcard.com/report/github.com/T-Campbell18/skip)
[![Go Reference](https://pkg.go.dev/badge/github.com/T-Campbell18/skip/skipmap.svg)](https://pkg.go.dev/github.com/T-Campbell18/skip/skipmap)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance, thread-safe skip list implementation in Go with generic support.

## Overview

`skipmap` provides a concurrent, ordered map data structure based on a skip list. It offers O(log n) average time complexity for search, insertion, and deletion operations. It's designed from the ground up to be thread-safe for concurrent use and supports generic key-value pairs, making it a flexible replacement for scenarios where a sorted, concurrent map is needed.

### Why use `skipmap`?

Go's built-in `map` is highly optimized but is not safe for concurrent access. `sync.Map` is concurrent but is optimized for "write-once, read-many" workloads and, crucially, does not maintain key order.

`skipmap` is ideal when you need **both** concurrent read/write access **and** the ability to efficiently query sorted data. Common use cases include:

- Implementing priority queues.
- Storing time-series data that needs to be queried by range.
- Building leaderboards or any system requiring ordered ranking.

## Features

- **Thread-Safe**: Designed for high-concurrency with reader-writer mutex protection.
- **Generic**: Supports any key and value types. Keys can be any `cmp.Ordered` type or use a custom comparator.
- **Ordered Operations**: Keys are always sorted, allowing for efficient range queries.
- **Min/Max Access**: O(1) access to the minimum element and O(log n) access to the maximum.
- **Memory Efficient**: Reuses internal buffers to reduce GC pressure during write operations.

## Installation

To use `skipmap`, import it in your Go code:

```go
import "github.com/T-Campbell18/skip/skipmap"
```

Then, run `go mod tidy` or `go get` to add the dependency to your project.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/T-Campbell18/skip/skipmap"
)

func main() {
    // Create a new skipmap with integer keys and string values
    sm := skipmap.New[int, string]()

    // Add some key-value pairs
    sm.Put(1, "one")
    sm.Put(3, "three")
    sm.Put(2, "two")

    // Get a value
    if value, exists := sm.Get(2); exists {
        fmt.Printf("Value for key 2: %s\n", value)
    }

    // Check if empty
    fmt.Printf("Is empty: %t\n", sm.IsEmpty())
    fmt.Printf("Length: %d\n", sm.Len())

    // Get min and max
    if minKey, minValue, exists := sm.Min(); exists {
        fmt.Printf("Min: key=%d, value=%s\n", minKey, minValue)
    }

    if maxKey, maxValue, exists := sm.Max(); exists {
        fmt.Printf("Max: key=%d, value=%s\n", maxKey, maxValue)
    }
}

/* Expected Output:
Value for key 2: two
Is empty: false
Length: 3
Min: key=1, value=one
Max: key=3, value=three
*/
```

## API Reference

### Constructor Functions

#### `New[K cmp.Ordered, V any]() *SkipMap[K, V]`

Creates a new skipmap with default settings for standard ordered key types (e.g., `int`, `string`, `float64`).

#### `NewWithComparator[K any, V any](comparator func(a, b K) int) *SkipMap[K, V]`

Creates a new skipmap with a custom comparator function for complex key types.

### Core Methods

#### `Put(key K, value V)`

Inserts a key-value pair. If the key already exists, its value is updated. Time complexity: O(log n) average.

#### `Get(key K) (V, bool)`

Retrieves a value by key. Returns the value and a boolean indicating if the key exists. Time complexity: O(log n) average.

#### `Remove(key K) bool`

Removes a key-value pair. Returns `true` if the key was found and removed. Time complexity: O(log n) average.

#### `Len() int`

Returns the number of key-value pairs in the skipmap.

#### `IsEmpty() bool`

Returns `true` if the skipmap contains no elements.

### Range Operations

#### `Range(start, end K) []V`

Returns a slice of all values for keys in the inclusive range `[start, end]`. Time complexity: O(log n + k) where k is the number of elements in the range.

#### `RangeFunc(start, end K, f func(key K, value V) bool)`

Iterates over elements in the range `[start, end]`, calling function `f` for each. If `f` returns `false`, iteration stops. This is more memory-efficient than `Range` for large result sets.

### Min/Max Operations

#### `Min() (K, V, bool)`

Returns the minimum key-value pair. Returns zero values and `false` if the map is empty. Time complexity: O(1).

#### `Max() (K, V, bool)`

Returns the maximum key-value pair. Returns zero values and `false` if the map is empty. Time complexity: O(log n).

## Advanced Usage

### Custom Comparators

For complex key types that don't implement `cmp.Ordered`, provide a custom comparator.

```go
import "strings"

type Person struct {
    Name string
    Age  int
}

// comparePerson sorts by Age, then by Name.
func comparePerson(a, b Person) int {
    if a.Age < b.Age {
        return -1
    }
    if a.Age > b.Age {
        return 1
    }
    return strings.Compare(a.Name, b.Name)
}

// Create skipmap with the custom comparator
sm := skipmap.NewWithComparator[Person, string](comparePerson)
```

### Concurrent Usage

The skipmap is thread-safe and can be used from multiple goroutines.

```go
var wg sync.WaitGroup
sm := skipmap.New[int, string]()

// Concurrent writes
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(i int) {
        defer wg.Done()
        sm.Put(i, fmt.Sprintf("value_%d", i))
    }(i)
}
wg.Wait() // Ensure all writes are complete

// Concurrent reads
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(i int) {
        defer wg.Done()
        if value, exists := sm.Get(i); exists {
            // Do something with the value
        }
    }(i)
}
wg.Wait()
```

### Range Queries

```go
sm := skipmap.New[int, string]()
for i := 0; i < 100; i++ {
    sm.Put(i, fmt.Sprintf("val_%d", i))
}

// Get all values in range [10, 20]
values := sm.Range(10, 20)
fmt.Printf("Found %d values in range\n", len(values))

// Iterate over a range with a function
sm.RangeFunc(5, 15, func(key int, value string) bool {
    fmt.Printf("Key: %d, Value: %s\n", key, value)
    if key == 10 {
        return false // Stop iteration
    }
    return true // Continue iteration
})
```

## Performance

- **Average Time Complexity**:
  - Search: O(log n)
  - Insert: O(log n)
  - Delete: O(log n)
  - Min: O(1)
  - Max: O(log n)
  - Range: O(log n + k)
- **Space Complexity**: O(n)
- **Concurrency Model**: A single `sync.RWMutex` protects the entire data structure, providing thread safety.

## Configuration

The skipmap uses these default constants, which are not currently configurable at runtime:

```go
const (
    DefaultMaxLevel    = 32  // Maximum number of levels in the skip list
    DefaultProbability = 0.5 // Probability used for random level generation
)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
