package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type GossipMessage struct {
	NodeID    string            `json:"nodeId"`
	Timestamp int64             `json:"timestamp"`
	Payload   map[string]string `json:"payload"`
}

type GossipProtocol struct {
	nodeID    string
	port      int
	peers     []string
	mu        sync.RWMutex
	conn      *net.UDPConn
	knownMsgs map[string]int64
	onReceive func(GossipMessage)
}

func NewGossipProtocol(nodeID string, port int, initialPeers []string) *GossipProtocol {
	return &GossipProtocol{
		nodeID:    nodeID,
		port:      port,
		peers:     initialPeers,
		knownMsgs: make(map[string]int64),
	}
}

func (g *GossipProtocol) Start(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	g.conn = conn

	go g.listenLoop(ctx)
	go g.antiEntropyLoop(ctx)

	return nil
}

func (g *GossipProtocol) Stop() {
	if g.conn != nil {
		g.conn.Close()
	}
}

func (g *GossipProtocol) Broadcast(payload map[string]string) {
	msg := GossipMessage{
		NodeID:    g.nodeID,
		Timestamp: time.Now().UnixNano(),
		Payload:   payload,
	}

	msgID := fmt.Sprintf("%s-%d", msg.NodeID, msg.Timestamp)
	g.mu.Lock()
	g.knownMsgs[msgID] = msg.Timestamp
	g.mu.Unlock()

	data, _ := json.Marshal(msg)

	g.mu.RLock()
	peers := make([]string, len(g.peers))
	copy(peers, g.peers)
	g.mu.RUnlock()

	// Select a random subset to gossip to
	for _, peer := range peers {
		// in a real impl, we'd limit fanout. for now, broadcast to all config peers.
		g.sendUDP(peer, data)
	}
}

func (g *GossipProtocol) OnMessage(fn func(GossipMessage)) {
	g.onReceive = fn
}

func (g *GossipProtocol) listenLoop(ctx context.Context) {
	buf := make([]byte, 8192)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		g.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, _, err := g.conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		var msg GossipMessage
		if err := json.Unmarshal(buf[:n], &msg); err != nil {
			continue
		}

		msgID := fmt.Sprintf("%s-%d", msg.NodeID, msg.Timestamp)

		g.mu.Lock()
		_, known := g.knownMsgs[msgID]
		if !known {
			g.knownMsgs[msgID] = msg.Timestamp
		}
		g.mu.Unlock()

		if !known {
			if g.onReceive != nil {
				g.onReceive(msg)
			}

			// Re-gossip
			data, _ := json.Marshal(msg)
			g.mu.RLock()
			for _, peer := range g.peers {
				g.sendUDP(peer, data)
			}
			g.mu.RUnlock()
		}
	}
}

func (g *GossipProtocol) antiEntropyLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Cleanup old messages
			now := time.Now().UnixNano()
			g.mu.Lock()
			for id, ts := range g.knownMsgs {
				if now-ts > int64(10*time.Minute) {
					delete(g.knownMsgs, id)
				}
			}
			g.mu.Unlock()
		}
	}
}

func (g *GossipProtocol) sendUDP(addr string, data []byte) {
	uaddr, err := net.ResolveUDPAddr("udp", addr)
	if err == nil && g.conn != nil {
		g.conn.WriteToUDP(data, uaddr)
	}
}
