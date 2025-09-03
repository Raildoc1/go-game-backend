package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub stores active websocket connections by user ID.
type Hub struct {
	mu    sync.Mutex
	conns map[int64]*websocket.Conn
}

// NewHub creates a new hub.
func NewHub() *Hub { return &Hub{conns: make(map[int64]*websocket.Conn)} }

// Add registers connection for user.
func (h *Hub) Add(userID int64, c *websocket.Conn) {
	h.mu.Lock()
	h.conns[userID] = c
	h.mu.Unlock()
}

// Remove closes and removes connection for user.
func (h *Hub) Remove(userID int64) {
	h.mu.Lock()
	if c, ok := h.conns[userID]; ok {
		_ = c.Close()
		delete(h.conns, userID)
	}
	h.mu.Unlock()
}

// Notify sends message to user's connection if present.
func (h *Hub) Notify(userID int64, msg string) {
	h.mu.Lock()
	c, ok := h.conns[userID]
	h.mu.Unlock()
	if ok {
		_ = c.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}
