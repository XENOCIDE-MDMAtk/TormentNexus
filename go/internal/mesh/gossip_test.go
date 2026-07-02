package mesh

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestGossipProtocol(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g1 := NewGossipProtocol("node1", 10001, []string{"127.0.0.1:10002"})
	g2 := NewGossipProtocol("node2", 10002, []string{"127.0.0.1:10001"})

	if err := g1.Start(ctx); err != nil {
		t.Fatalf("Failed to start g1: %v", err)
	}
	defer g1.Stop()

	if err := g2.Start(ctx); err != nil {
		t.Fatalf("Failed to start g2: %v", err)
	}
	defer g2.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	var received bool
	g2.OnMessage(func(msg GossipMessage) {
		if msg.Payload["test"] == "hello" {
			received = true
			wg.Done()
		}
	})

	// Wait for sockets to bind
	time.Sleep(100 * time.Millisecond)

	g1.Broadcast(map[string]string{"test": "hello"})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for gossip message")
	case <-done:
		if !received {
			t.Fatal("Failed to receive expected gossip payload")
		}
	}
}
