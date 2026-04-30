package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/coder/websocket"
)

// Pools for generating readable short IDs.
var firstNamePool = []string{
	"alex", "mike", "tom", "jack", "leo", "sam", "eric", "lucas", "liam", "noah",
	"emma", "olivia", "sophia", "mia", "ava", "lily", "grace", "ella", "zoe", "nina",
}

var tagPool = []string{
	"mail", "inbox", "user", "note", "cloud", "river", "forest", "stone", "ocean", "field",
	"sun", "moon", "star", "leaf", "bird", "fox", "wolf", "lake", "hill", "wind",
}

var shortIDRegex = regexp.MustCompile(`^[a-z0-9._\-+]{1,64}$`)

// ---------------------------------------------------------------------------
// Message types for the JSON protocol
// ---------------------------------------------------------------------------

// inboundMsg represents a message received from a client.
type inboundMsg struct {
	Type    string `json:"type"`
	ShortID string `json:"short_id,omitempty"`
}

// outboundMsg is the envelope for messages sent to clients.
type outboundMsg struct {
	Type    string      `json:"type"`
	ShortID string      `json:"short_id,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// MailData is the JSON representation of a mail sent to clients.
type MailData struct {
	ID        int64  `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	HTML      string `json:"html"`
	CreatedAt string `json:"created_at"`
}

// Message is the internal broadcast message passed through the hub.
type Message struct {
	ShortID string
	Data    interface{}
}

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// Client represents a single WebSocket connection.
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	shortID string
}

// readPump reads messages from the WebSocket connection and dispatches them
// to the hub. It runs until the connection is closed.
func (c *Client) readPump(ctx context.Context) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close(websocket.StatusNormalClosure, "bye")
	}()

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return
		}

		var msg inboundMsg
		if err := json.Unmarshal(data, &msg); err != nil {
			c.sendError("invalid message format")
			continue
		}

		switch msg.Type {
		case "request_shortid":
			c.hub.unregister <- c
			c.shortID = c.hub.generateShortID()
			c.hub.register <- c
			c.sendShortID(c.shortID)

		case "set_shortid":
			normalized := normalizeShortID(msg.ShortID)
			if normalized == "" {
				c.sendError("invalid short id")
				continue
			}
			if c.hub.isBlacklisted(normalized) {
				c.sendError("short id in blacklist")
				continue
			}
			c.hub.unregister <- c
			c.shortID = normalized
			c.hub.register <- c
			c.sendShortID(c.shortID)

		default:
			c.sendError("unknown message type")
		}
	}
}

// writePump pumps messages from the send channel to the WebSocket connection.
func (c *Client) writePump(ctx context.Context) {
	defer c.conn.Close(websocket.StatusNormalClosure, "bye")

	for {
		select {
		case data, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.Write(ctx, websocket.MessageText, data); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// sendShortID sends a shortid message to the client.
func (c *Client) sendShortID(id string) {
	msg := outboundMsg{
		Type:    "shortid",
		ShortID: id,
	}
	data, _ := json.Marshal(msg)
	c.trySend(data)
}

// sendError sends an error message to the client.
func (c *Client) sendError(text string) {
	msg := outboundMsg{
		Type:    "error",
		Message: text,
	}
	data, _ := json.Marshal(msg)
	c.trySend(data)
}

// trySend enqueues data into the send channel, dropping it silently if the
// channel is full (client too slow).
func (c *Client) trySend(data []byte) {
	select {
	case c.send <- data:
	default:
		// drop if client is not reading fast enough
	}
}

// ---------------------------------------------------------------------------
// Hub
// ---------------------------------------------------------------------------

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	clients    map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	blacklist  []string
	mailHost   string
}

// NewHub creates a new Hub. blacklist is a list of keywords that should not
// appear in short IDs. mailHost is the domain used for email addresses.
func NewHub(blacklist []string, mailHost string) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
		blacklist:  blacklist,
		mailHost:   mailHost,
	}
}

// Run starts the hub's main event loop. It should be called as a goroutine.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case client := <-h.register:
			set, ok := h.clients[client.shortID]
			if !ok {
				set = make(map[*Client]bool)
				h.clients[client.shortID] = set
			}
			set[client] = true

		case client := <-h.unregister:
			if set, ok := h.clients[client.shortID]; ok {
				delete(set, client)
				if len(set) == 0 {
					delete(h.clients, client.shortID)
				}
			}

		case msg := <-h.broadcast:
			if set, ok := h.clients[msg.ShortID]; ok {
				data, _ := json.Marshal(outboundMsg{
					Type: "mail",
					Data: msg.Data,
				})
				for client := range set {
					client.trySend(data)
				}
			}
		}
	}
}

// SendTo broadcasts a mail message to all clients watching the given shortID.
func (h *Hub) SendTo(shortID string, data any) {
	h.broadcast <- &Message{
		ShortID: shortID,
		Data:    data,
	}
}

// HandleWS upgrades an HTTP request to a WebSocket connection and starts
// the client read/write pumps.
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionContextTakeover,
	})
	if err != nil {
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	ctx := conn.CloseRead(context.Background())

	h.register <- client

	go client.writePump(ctx)
	client.readPump(r.Context())
}

// generateShortID produces a random human-readable short ID of the form
// firstName + tag + 3-digit number (e.g. "alexsun789"). It validates against
// the blacklist and retries up to 20 times before falling back to a
// timestamp-based ID.
func (h *Hub) generateShortID() string {
	for i := 0; i < 20; i++ {
		name := firstNamePool[rand.Intn(len(firstNamePool))]
		tag := tagPool[rand.Intn(len(tagPool))]
		suffix := rand.Intn(900) + 100
		candidate := fmt.Sprintf("%s%s%d", name, tag, suffix)

		if !shortIDRegex.MatchString(candidate) {
			continue
		}
		if h.isBlacklisted(candidate) {
			continue
		}
		if _, taken := h.clients[candidate]; taken {
			continue
		}
		return candidate
	}

	// Extremely rare fallback path.
	return fmt.Sprintf("mailuser%d", time.Now().UnixMilli())
}

// isBlacklisted checks whether id contains any blacklisted keyword as a
// substring (case-insensitive).
func (h *Hub) isBlacklisted(id string) bool {
	lower := strings.ToLower(id)
	for _, kw := range h.blacklist {
		if kw == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// normalizeShortID trims and lowercases the input, returning it only if it
// passes the regex validation. Returns empty string on failure.
func normalizeShortID(id string) string {
	normalized := strings.TrimSpace(strings.ToLower(id))
	if !shortIDRegex.MatchString(normalized) {
		return ""
	}
	return normalized
}

// ClientCount returns the number of active clients (for health/debug).
// This is safe to call from any goroutine but the value is approximate.
func (h *Hub) ClientCount() int {
	count := 0
	for _, set := range h.clients {
		count += len(set)
	}
	return count
}

// Close gracefully shuts down the hub by closing all client connections.
func (h *Hub) Close() {
	for _, set := range h.clients {
		for client := range set {
			close(client.send)
			client.conn.Close(websocket.StatusGoingAway, "server shutting down")
		}
	}
}
