package ws

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
	ID             int64    `json:"id"`
	From           string   `json:"from"`
	To             string   `json:"to"`
	Subject        string   `json:"subject"`
	HTML           string   `json:"html"`
	IsRead         bool     `json:"is_read"`
	ExtractedCodes []string `json:"extracted_codes"`
	ExtractedLinks []string `json:"extracted_links"`
	CreatedAt      string   `json:"created_at"`
}

// Message is the internal broadcast message passed through the hub.
type Message struct {
	ShortID string
	Data    interface{}
}
