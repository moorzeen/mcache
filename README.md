# mcache

A simple, lightweight in-memory cache library for Go with automatic TTL-based expiration.

## Features

- Thread-safe operations
- Automatic cleanup of expired items in the background
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
    // Create a cache with a 5-minute TTL
    c := mcache.NewCache(5 * time.Minute)

    // Store a value
    c.Set("key", "value")

    // Retrieve a value
    if val, ok := c.Get("key"); ok {
        fmt.Println(val) // value
    }

    // Retrieve and delete a value in one step
    if val, ok := c.Release("key"); ok {
        fmt.Println(val) // value
    }

    // Get all non-expired items
    items := c.GetAll()
    fmt.Println(items)
}
```

## API

### `NewCache(ttl time.Duration) *Cache`

Creates a new cache instance. The `ttl` parameter sets the lifetime for all stored items. A background goroutine is started to periodically clean up expired entries.

### `Set(key string, value interface{})`

Stores a value under the given key. Overwrites any existing value for that key.

### `Get(key string) (interface{}, bool)`

Returns the value associated with the key. Returns `nil, false` if the key does not exist or has expired.

### `Release(key string) (interface{}, bool)`

Returns the value associated with the key and removes it from the cache in a single atomic operation. Returns `nil, false` if the key does not exist or has expired.

### `GetAll() map[string]interface{}`

Returns all non-expired items as a map.
