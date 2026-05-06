package ws

import (
	"context"
	"encoding/json"
	"log/slog"

	"forsaken-mail/internal/i18n"

	"github.com/coder/websocket"
)

// Client represents a single WebSocket connection.
type Client struct {
	hub                *Hub
	conn               *websocket.Conn
	send               chan []byte
	shortIDs           map[string]bool
	registeredShortIDs map[string]bool // shortIDs actually registered in hub
	lang               string
}

// readPump reads messages from the WebSocket connection and dispatches them
// to the hub. It runs until the connection is closed.
func (c *Client) readPump(ctx context.Context) {
	defer func() {
		select {
		case c.hub.unregister <- c:
		case <-c.hub.done:
		}
		c.conn.Close(websocket.StatusNormalClosure, "bye")
	}()

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			slog.Debug("websocket read error", "err", err)
			return
		}

		slog.Debug("websocket received", "data", string(data))

		var msg inboundMsg
		if err := json.Unmarshal(data, &msg); err != nil {
			c.sendError(i18n.T(c.lang, "invalid_message_format"))
			continue
		}

		switch msg.Type {
		case "request_shortid":
			c.hub.unregister <- c
			c.shortIDs = map[string]bool{c.hub.generateShortID(): true}
			c.hub.register <- c
			for id := range c.shortIDs {
				c.sendShortID(id)
				slog.Info("websocket assigned shortid", "short_id", id)
			}

		case "set_shortid":
			normalized := normalizeShortID(msg.ShortID)
			if normalized == "" {
				c.sendError(i18n.T(c.lang, "invalid_short_id"))
				continue
			}
			if c.hub.isBlacklisted(normalized) {
				c.sendError(i18n.T(c.lang, "shortid_in_blacklist"))
				continue
			}
			c.hub.unregister <- c
			c.shortIDs = map[string]bool{normalized: true}
			c.hub.register <- c
			c.sendShortID(normalized)
			slog.Info("websocket set shortid", "short_id", normalized)

		case "subscribe":
			normalized := normalizeShortID(msg.ShortID)
			if normalized == "" {
				c.sendError(i18n.T(c.lang, "invalid_short_id"))
				continue
			}
			if c.hub.isBlacklisted(normalized) {
				c.sendError(i18n.T(c.lang, "shortid_in_blacklist"))
				continue
			}
			c.hub.subscribe <- subscribeMsg{client: c, shortID: normalized}
			c.sendShortID(normalized)
			slog.Info("websocket subscribe", "short_id", normalized)

		case "unsubscribe":
			normalized := normalizeShortID(msg.ShortID)
			if normalized == "" {
				continue
			}
			c.hub.unsubscribe <- subscribeMsg{client: c, shortID: normalized}

		default:
			c.sendError(i18n.T(c.lang, "unknown_message_type"))
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
