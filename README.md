# mcache

A simple, lightweight in-memory cache library for Go with automatic TTL-based expiration.

## Features

- Generic — type-safe keys and values via Go generics (Go 1.18+)
- Thread-safe operations
- Configurable cleanup interval independent of TTL
- Graceful shutdown via `Close()`
- No external dependencies

## Installation

```bash
go get github.com/moorzeen/mcache
```

## Usage

```go
package main

import (
    "fmt"
    "time"

    "github.com/moorzeen/mcache"
)

func main() {
    // Create a cache with a 5-minute TTL and 1-minute cleanup interval
    c := mcache.NewCache[string, int](5*time.Minute, time.Minute)
    defer c.Close()

    // Store a value
    c.Set("requests", 42)

    // Retrieve a value
    if val, ok := c.Get("requests"); ok {
        fmt.Println(val) // 42
    }

    // Retrieve and delete a value in one step
    if val, ok := c.Release("requests"); ok {
        fmt.Println(val) // 42
    }

    // Delete a value without retrieving it
    c.Delete("requests")

    // Count non-expired items
    fmt.Println(c.Count())

    // Get all non-expired items
    fmt.Println(c.GetAll())
}
```

## API

### `NewCache[K comparable, V any](ttl, cleanupInterval time.Duration) *Cache[K, V]`

Creates a new cache instance. `ttl` sets the lifetime for stored items. `cleanupInterval` controls how often the background goroutine scans and removes expired entries — set it lower than `ttl` to free memory sooner.

### `Close()`

Stops the background cleanup goroutine. Should be called when the cache is no longer needed, typically via `defer`.

### `Set(key K, value V)`

Stores a value under the given key. Overwrites any existing value for that key.

### `Get(key K) (V, bool)`

Returns the value associated with the key. Returns the zero value and `false` if the key does not exist or has expired.

### `Release(key K) (V, bool)`

Returns the value associated with the key and removes it from the cache in a single atomic operation. Returns the zero value and `false` if the key does not exist or has expired.

### `Delete(key K)`

Removes the key from the cache unconditionally.

### `Count() int`

Returns the number of non-expired items.

### `GetAll() map[K]V`

Returns all non-expired items as a map.

## Testing

```bash
go test -race ./...
```