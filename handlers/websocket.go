package handlers

import (
	"log"
	"time"

	ws "eman-backend/websocket"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10
)

type WebSocketHandler struct {
	hub *ws.Hub
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		hub: ws.GetHub(),
	}
}

// Upgrade checks if the request is a WebSocket upgrade request
func (h *WebSocketHandler) Upgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// Handle manages WebSocket connections
func (h *WebSocketHandler) Handle(c *websocket.Conn) {
	client := &ws.Client{
		Conn: c,
		Hub:  h.hub,
	}

	h.hub.Register(client)

	// Start ping ticker for keepalive
	ticker := time.NewTicker(pingPeriod)
	done := make(chan struct{})

	// Goroutine for sending pings
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("Ping error: %v", err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Set initial read deadline
	c.SetReadDeadline(time.Now().Add(pongWait))

	// Handle pong messages to reset read deadline
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Keep connection alive and handle incoming messages
	defer func() {
		close(done)
		h.hub.Unregister(client)
	}()

	for {
		messageType, msg, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Handle text messages (like heartbeat from client)
		if messageType == websocket.TextMessage {
			// Client sent a heartbeat or message
			if string(msg) == "ping" {
				c.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.WriteMessage(websocket.TextMessage, []byte("pong")); err != nil {
					break
				}
			}
		}
	}
}

// GetHub returns the WebSocket hub for broadcasting
func (h *WebSocketHandler) GetHub() *ws.Hub {
	return h.hub
}
