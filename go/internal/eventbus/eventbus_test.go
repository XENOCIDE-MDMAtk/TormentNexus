package eventbus

import (
	"strings"
	"testing"
	"sync/atomic"
)

func TestEmitExact(t *testing.T) {
	bus := New(100)
	var called int32
	bus.Subscribe("tool:call", func(e SystemEvent) {
		atomic.AddInt32(&called, 1)
		if e.Source != "test" {
			t.Errorf("expected source=test, got %s", e.Source)
		}
	})
	bus.EmitEvent("tool:call", "test", nil)
	if atomic.LoadInt32(&called) != 1 {
		t.Errorf("expected 1 call, got %d", called)
	}
}

func TestEmitWildcard(t *testing.T) {
	bus := New(100)
	var called int32
	bus.Subscribe("file:*", func(e SystemEvent) {
		atomic.AddInt32(&called, 1)
	})
	bus.EmitEvent("file:change", "watcher", nil)
	bus.EmitEvent("file:write", "watcher", nil)
	bus.EmitEvent("tool:call", "other", nil) // should NOT match
	if atomic.LoadInt32(&called) != 2 {
		t.Errorf("expected 2 calls, got %d", called)
	}
}

func TestGlobalListener(t *testing.T) {
	bus := New(100)
	var count int32
	bus.OnGlobal(func(e SystemEvent) {
		atomic.AddInt32(&count, 1)
	})
	bus.EmitEvent("agent:start", "a", nil)
	bus.EmitEvent("tool:call", "b", nil)
	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestHistory(t *testing.T) {
	bus := New(5)
	for i := 0; i < 10; i++ {
		bus.EmitEvent("agent:heartbeat", "src", i)
	}
	h := bus.GetHistory(3)
	if len(h) != 3 {
		t.Fatalf("expected 3, got %d", len(h))
	}
	if h[0].Payload.(int) != 7 {
		t.Errorf("expected oldest=7, got %v", h[0].Payload)
	}
	if h[2].Payload.(int) != 9 {
		t.Errorf("expected newest=9, got %v", h[2].Payload)
	}
}

func TestGlobToRegex(t *testing.T) {
	re := globToRegex("file:*")
	if !re.MatchString("file:change") {
		t.Error("should match file:change")
	}
	if re.MatchString("tool:call") {
		t.Error("should NOT match tool:call")
	}
	_ = strings.Contains("x", "x") // use strings import
}
