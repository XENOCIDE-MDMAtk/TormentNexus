package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tormentnexushq/tormentnexus-go/internal/eventbus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dashboard integration
	},
}

type WSClient struct {
	conn *websocket.Conn
	send chan []byte
}

type WSBroker struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.Mutex
}

var globalWSBroker = &WSBroker{
	clients:    make(map[*WSClient]bool),
	broadcast:  make(chan []byte, 100),
	register:   make(chan *WSClient),
	unregister: make(chan *WSClient),
}

func init() {
	go globalWSBroker.run()
}

func (b *WSBroker) run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = true
			b.mu.Unlock()
		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client.send)
			}
			b.mu.Unlock()
		case message := <-b.broadcast:
			b.mu.Lock()
			for client := range b.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(b.clients, client)
				}
			}
			b.mu.Unlock()
		}
	}
}

// StartWSBroker subscribes the global WebSocket broker to EventBus system events.
func (s *Server) StartWSBroker() {
	s.eventBus.OnGlobal(func(ev eventbus.SystemEvent) {
		// Translate EventBus events into TrafficInspector expected JSON packets
		var translated []byte
		var err error

		switch string(ev.Type) {
		case "tool:call":
			payloadMap, ok := ev.Payload.(map[string]interface{})
			if !ok {
				break
			}
			startPacket := map[string]interface{}{
				"type":      "TOOL_CALL_START",
				"id":        payloadMap["callId"],
				"tool":      payloadMap["tool"],
				"args":      payloadMap["args"],
				"timestamp": ev.Timestamp,
			}
			translated, err = json.Marshal(startPacket)

		case "tool:call:response", "tool:call:end":
			payloadMap, ok := ev.Payload.(map[string]interface{})
			if !ok {
				break
			}
			endPacket := map[string]interface{}{
				"type":      "TOOL_CALL_END",
				"id":        payloadMap["callId"],
				"tool":      payloadMap["tool"],
				"success":   true,
				"duration":  payloadMap["durationMs"],
				"result":    payloadMap["result"],
				"timestamp": ev.Timestamp,
			}
			translated, err = json.Marshal(endPacket)
		}

		if err == nil && len(translated) > 0 {
			select {
			case globalWSBroker.broadcast <- translated:
			default:
				// Channel buffer full, event dropped
			}
		}
	})
}

// handleMCPTrafficWS upgrades the HTTP connection to WebSocket and streams tool events.
func (s *Server) handleMCPTrafficWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[WS] Upgrade error: %v\n", err)
		return
	}

	client := &WSClient{
		conn: conn,
		send: make(chan []byte, 256),
	}
	globalWSBroker.register <- client

	// Read pump (keeps connection alive, handles close signals)
	go func() {
		defer func() {
			globalWSBroker.unregister <- client
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()

	// Write pump
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer func() {
			ticker.Stop()
			conn.Close()
		}()
		for {
			select {
			case message, ok := <-client.send:
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					return
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()
}
