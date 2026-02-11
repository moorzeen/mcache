package mcache

import (
	"sync"
	"testing"
	"time"
)

const (
	ttl             = 100 * time.Millisecond
	cleanupInterval = 50 * time.Millisecond
)

func newTestCache() *Cache[string, int] {
	return NewCache[string, int](ttl, cleanupInterval)
}

func TestSet_Get(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	c.Set("a", 1)

	val, ok := c.Get("a")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if val != 1 {
		t.Fatalf("expected 1, got %d", val)
	}
}

func TestGet_MissingKey(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	val, ok := c.Get("missing")
	if ok {
		t.Fatal("expected key to be missing")
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %d", val)
	}
}

func TestGet_Expired(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	c.Set("a", 42)
	time.Sleep(ttl + 10*time.Millisecond)

	_, ok := c.Get("a")
	if ok {
		t.Fatal("expected key to be expired")
	}
}

func TestSet_Overwrite(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	c.Set("a", 1)
	c.Set("a", 2)

	val, ok := c.Get("a")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if val != 2 {
		t.Fatalf("expected 2, got %d", val)
	}
}

func TestDelete(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	c.Set("a", 1)
	c.Delete("a")

	_, ok := c.Get("a")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestDelete_MissingKey(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	// should not panic
	c.Delete("missing")
}

func TestRelease(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	c.Set("a", 10)

	val, ok := c.Release("a")
	if !ok {
		t.Fatal("expected key to exist on release")
	}
	if val != 10 {
		t.Fatalf("expected 10, got %d", val)
	}

	_, ok = c.Get("a")
	if ok {
		t.Fatal("expected key to be removed after release")
	}
}

func TestRelease_Expired(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	c.Set("a", 1)
	time.Sleep(ttl + 10*time.Millisecond)

	_, ok := c.Release("a")
	if ok {
		t.Fatal("expected release to fail on expired key")
	}
}

func TestGetAll(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	all := c.GetAll()
	if len(all) != 3 {
		t.Fatalf("expected 3 items, got %d", len(all))
	}
}

func TestGetAll_ExcludesExpired(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	c.Set("a", 1)
	time.Sleep(ttl + 10*time.Millisecond)
	c.Set("b", 2)

	all := c.GetAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 item, got %d", len(all))
	}
	if _, ok := all["b"]; !ok {
		t.Fatal("expected key 'b' to be present")
	}
}

func TestCount(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)

	if n := c.Count(); n != 2 {
		t.Fatalf("expected count 2, got %d", n)
	}

	c.Delete("a")

	if n := c.Count(); n != 1 {
		t.Fatalf("expected count 1, got %d", n)
	}
}

func TestCount_ExcludesExpired(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	c.Set("a", 1)
	time.Sleep(ttl + 10*time.Millisecond)

	if n := c.Count(); n != 0 {
		t.Fatalf("expected count 0, got %d", n)
	}
}

func TestCleanup(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	c.Set("a", 1)
	time.Sleep(ttl + cleanupInterval + 10*time.Millisecond)

	c.mu.Lock()
	n := len(c.items)
	c.mu.Unlock()

	if n != 0 {
		t.Fatalf("expected cleanup to remove expired item, got %d items", n)
	}
}

func TestClose(t *testing.T) {
	c := newTestCache()
	// should not panic or block
	c.Close()
}

func TestConcurrency(t *testing.T) {
	c := newTestCache()
	defer c.Close()

	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune('a' + i%26))
			c.Set(key, i)
			c.Get(key)
			c.Count()
		}(i)
	}

	wg.Wait()
}
