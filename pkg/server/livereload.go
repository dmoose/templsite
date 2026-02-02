package server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// LiveReload manages WebSocket connections for live reload functionality
type LiveReload struct {
	connections map[*websocket.Conn]bool
	mu          sync.RWMutex
	upgrader    websocket.Upgrader
	broadcast   chan string
}

// NewLiveReload creates a new live reload manager
func NewLiveReload() *LiveReload {
	return &LiveReload{
		connections: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins in development
				return true
			},
		},
		broadcast: make(chan string, 10),
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (lr *LiveReload) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := lr.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("websocket upgrade failed", "error", err)
		return
	}

	lr.addConnection(conn)
	slog.Debug("live reload client connected", "remote", r.RemoteAddr)

	// Keep connection alive with ping/pong
	go lr.keepAlive(conn)

	// Read loop (mostly for detecting disconnects)
	lr.readLoop(conn)
}

// addConnection adds a WebSocket connection to the pool
func (lr *LiveReload) addConnection(conn *websocket.Conn) {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	lr.connections[conn] = true
}

// removeConnection removes a WebSocket connection from the pool
func (lr *LiveReload) removeConnection(conn *websocket.Conn) {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	delete(lr.connections, conn)
	conn.Close()
}

// readLoop reads messages from the WebSocket (mainly for disconnect detection)
func (lr *LiveReload) readLoop(conn *websocket.Conn) {
	defer lr.removeConnection(conn)

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Debug("websocket read error", "error", err)
			}
			break
		}
	}
}

// keepAlive sends periodic ping messages to keep the connection alive
func (lr *LiveReload) keepAlive(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}
}

// Start begins the broadcast loop
func (lr *LiveReload) Start(ctx context.Context) {
	go lr.broadcastLoop(ctx)
}

// broadcastLoop listens for reload signals and broadcasts to all clients
func (lr *LiveReload) broadcastLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-lr.broadcast:
			lr.notifyAll(msg)
		}
	}
}

// NotifyReload signals all connected clients to reload
func (lr *LiveReload) NotifyReload() {
	select {
	case lr.broadcast <- "reload":
		slog.Debug("reload notification queued")
	default:
		// Channel full, skip
	}
}

// notifyAll sends a message to all connected clients
func (lr *LiveReload) notifyAll(message string) {
	lr.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(lr.connections))
	for conn := range lr.connections {
		connections = append(connections, conn)
	}
	lr.mu.RUnlock()

	slog.Debug("broadcasting reload to clients", "count", len(connections))

	for _, conn := range connections {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			slog.Debug("failed to send reload message", "error", err)
			lr.removeConnection(conn)
		}
	}
}

// ConnectionCount returns the number of active connections
func (lr *LiveReload) ConnectionCount() int {
	lr.mu.RLock()
	defer lr.mu.RUnlock()
	return len(lr.connections)
}

// Close closes all connections and stops the broadcast loop
func (lr *LiveReload) Close() {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	for conn := range lr.connections {
		conn.Close()
	}
	lr.connections = make(map[*websocket.Conn]bool)
	close(lr.broadcast)
}

// LiveReloadScript returns the JavaScript code for live reload
func LiveReloadScript() string {
	return `<script>
(function() {
  'use strict';

  var protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  var wsUrl = protocol + '//' + window.location.host + '/_live-reload';
  var ws;
  var reconnectDelay = 1000;
  var maxReconnectDelay = 30000;

  function connect() {
    ws = new WebSocket(wsUrl);

    ws.onopen = function() {
      console.log('[templsite] Live reload connected');
      reconnectDelay = 1000;
    };

    ws.onmessage = function(event) {
      if (event.data === 'reload') {
        console.log('[templsite] Reloading page...');
        window.location.reload();
      }
    };

    ws.onclose = function() {
      console.log('[templsite] Live reload disconnected, reconnecting...');
      setTimeout(function() {
        reconnectDelay = Math.min(reconnectDelay * 2, maxReconnectDelay);
        connect();
      }, reconnectDelay);
    };

    ws.onerror = function(error) {
      console.log('[templsite] Live reload error:', error);
      ws.close();
    };
  }

  connect();
})();
</script>`
}
