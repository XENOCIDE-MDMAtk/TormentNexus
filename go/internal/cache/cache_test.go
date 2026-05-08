package cache

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	c := New(CacheOptions{MaxSize: 10, DefaultTTL: 5000})
	c.Set("foo", "bar")
	val, ok := c.Get("foo")
	if !ok || val.(string) != "bar" {
		t.Fatalf("expected bar, got %v, ok=%v", val, ok)
	}
}

func TestTTLExpiry(t *testing.T) {
	c := New(CacheOptions{MaxSize: 10, DefaultTTL: 50}) // 50ms TTL
	c.Set("short", "lived")
	time.Sleep(80 * time.Millisecond)
	_, ok := c.Get("short")
	if ok {
		t.Fatal("expected expired entry to be gone")
	}
}

func TestLRUEviction(t *testing.T) {
	c := New(CacheOptions{MaxSize: 3, DefaultTTL: 60000})
	c.Set("a", 1)
	time.Sleep(2 * time.Millisecond) // ensure different timestamps
	c.Set("b", 2)
	time.Sleep(2 * time.Millisecond)
	c.Set("c", 3)
	time.Sleep(2 * time.Millisecond)
	// Access "a" to make it recently used
	c.Get("a")
	time.Sleep(2 * time.Millisecond)
	// Adding "d" should evict "b" (oldest accessed)
	c.Set("d", 4)

	if c.Has("b") {
		t.Error("expected b to be evicted")
	}
	if !c.Has("a") {
		t.Error("expected a to remain (recently accessed)")
	}
}

func TestDelete(t *testing.T) {
	c := New(CacheOptions{})
	c.Set("x", 1)
	if !c.Delete("x") {
		t.Fatal("expected delete to return true")
	}
	if c.Delete("x") {
		t.Fatal("expected second delete to return false")
	}
}

func TestEventCallback(t *testing.T) {
	var events []string
	c := New(CacheOptions{})
	c.OnEvent(func(e CacheEvent) {
		events = append(events, e.Type)
	})
	c.Set("k", "v")
	c.Get("k")
	c.Get("missing")
	c.Delete("k")

	expected := []string{"set", "hit", "miss", "delete"}
	if len(events) != len(expected) {
		t.Fatalf("expected %d events, got %d: %v", len(expected), len(events), events)
	}
	for i, e := range expected {
		if events[i] != e {
			t.Errorf("event[%d]: expected %s, got %s", i, e, events[i])
		}
	}
}

func TestSingleton(t *testing.T) {
	a := GetInstance("test-ns")
	b := GetInstance("test-ns")
	if a != b {
		t.Fatal("expected same instance")
	}
}

func TestCached(t *testing.T) {
	c := New(CacheOptions{DefaultTTL: 5000})
	var calls int32
	fn := func() (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		return "computed", nil
	}

	val, err := Cached(c, "key", fn, 0)
	if err != nil || val.(string) != "computed" {
		t.Fatalf("unexpected: %v %v", val, err)
	}
	// Second call should use cache
	val, err = Cached(c, "key", fn, 0)
	if err != nil || val.(string) != "computed" {
		t.Fatalf("unexpected: %v %v", val, err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestCleanup(t *testing.T) {
	c := New(CacheOptions{
		MaxSize:         100,
		DefaultTTL:      30, // 30ms
		CleanupInterval: 20, // 20ms
	})
	c.Set("ephemeral", "gone")
	time.Sleep(80 * time.Millisecond)
	size, _ := c.Stats()
	if size != 0 {
		t.Fatalf("expected cleanup to remove expired entry, size=%d", size)
	}
	c.Dispose()
}
