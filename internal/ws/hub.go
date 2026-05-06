package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/coder/websocket"
)

// subscribeMsg is sent to the hub to add/remove a client's subscription.
type subscribeMsg struct {
	client  *Client
	shortID string
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	clients     map[string]map[*Client]bool
	register    chan *Client
	unregister  chan *Client
	subscribe   chan subscribeMsg
	unsubscribe chan subscribeMsg
	broadcast   chan *Message
	blacklist   []string
	mailHost    string
	shutdown    chan struct{} // signals Run() to shut down
	done        chan struct{} // Run() signals completion
	clientCount atomic.Int64
}

// NewHub creates a new Hub. blacklist is a list of keywords that should not
// appear in short IDs. mailHost is the domain used for email addresses.
func NewHub(blacklist []string, mailHost string) *Hub {
	return &Hub{
		clients:     make(map[string]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		subscribe:   make(chan subscribeMsg),
		unsubscribe: make(chan subscribeMsg),
		broadcast:   make(chan *Message, 256),
		blacklist:   blacklist,
		mailHost:    mailHost,
		shutdown:    make(chan struct{}),
		done:        make(chan struct{}),
	}
}

// Run starts the hub's main event loop. It should be called as a goroutine.
func (h *Hub) Run(ctx context.Context) {
	defer close(h.done)
	for {
		select {
		case <-ctx.Done():
			h.cleanupAllClients()
			return
		case <-h.shutdown:
			h.cleanupAllClients()
			return

		case client := <-h.register:
			client.registeredShortIDs = make(map[string]bool)
			for id := range client.shortIDs {
				set, ok := h.clients[id]
				if !ok {
					set = make(map[*Client]bool)
					h.clients[id] = set
				}
				set[client] = true
				client.registeredShortIDs[id] = true
			}
			h.clientCount.Add(1)

		case client := <-h.unregister:
			found := false
			for id := range client.registeredShortIDs {
				if set, ok := h.clients[id]; ok {
					delete(set, client)
					if len(set) == 0 {
						delete(h.clients, id)
					}
					found = true
				}
			}
			client.registeredShortIDs = nil
			if found {
				h.clientCount.Add(-1)
			}

		case sub := <-h.subscribe:
			sub.client.shortIDs[sub.shortID] = true
			set, ok := h.clients[sub.shortID]
			if !ok {
				set = make(map[*Client]bool)
				h.clients[sub.shortID] = set
			}
			set[sub.client] = true

		case sub := <-h.unsubscribe:
			delete(sub.client.shortIDs, sub.shortID)
			if set, ok := h.clients[sub.shortID]; ok {
				delete(set, sub.client)
				if len(set) == 0 {
					delete(h.clients, sub.shortID)
				}
			}

		case msg := <-h.broadcast:
			if set, ok := h.clients[msg.ShortID]; ok {
				data, _ := json.Marshal(outboundMsg{
					Type:    "mail",
					ShortID: msg.ShortID,
					Data:    msg.Data,
				})
				for client := range set {
					client.trySend(data)
				}
			}
		}
	}
}

// cleanupAllClients closes all client connections and resets the hub state.
// Must be called from within Run().
func (h *Hub) cleanupAllClients() {
	for _, set := range h.clients {
		for client := range set {
			close(client.send)
			client.conn.Close(websocket.StatusGoingAway, "server shutting down")
		}
	}
	h.clients = make(map[string]map[*Client]bool)
	h.clientCount.Store(0)
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
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, lang string) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionContextTakeover,
	})
	if err != nil {
		slog.Error("websocket accept failed", "err", err, "remote", r.RemoteAddr)
		return
	}

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		shortIDs: make(map[string]bool),
		lang:     lang,
	}

	slog.Info("websocket connected", "remote", r.RemoteAddr)
	h.register <- client

	go client.writePump(r.Context())
	client.readPump(r.Context())
	slog.Info("websocket disconnected", "remote", r.RemoteAddr)
}

// ClientCount returns the number of active clients (for health/debug).
// This is safe to call from any goroutine.
func (h *Hub) ClientCount() int {
	return int(h.clientCount.Load())
}

// Close gracefully shuts down the hub by signaling Run() to clean up all
// client connections and waiting for completion.
func (h *Hub) Close() {
	close(h.shutdown)
	<-h.done
}
