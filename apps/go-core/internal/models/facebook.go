package models

// top-level payload from Meta
type MetaWebhookPayload struct {
	Object string         `json:"object"` // "page" or "instagram"
	Entry  []WebhookEntry `json:"entry"`
}

type WebhookEntry struct {
	ID        string           `json:"id"` // Page ID or IG user ID
	Time      int64            `json:"time"`
	Messaging []MessagingEvent `json:"messaging"`
	Changes   []ChangeEvent    `json:"changes,omitempty"` // comments, feed updates
}

type MessagingEvent struct {
	Sender    WebhookUser     `json:"sender"`
	Recipient WebhookUser     `json:"recipient"`
	Timestamp int64           `json:"timestamp"`
	Message   *InboundMessage `json:"message,omitempty"`
	Postback  *Postback       `json:"postback,omitempty"`
	Delivery  *Delivery       `json:"delivery,omitempty"`
	Read      *Read           `json:"read,omitempty"`
}

type WebhookUser struct {
	ID string `json:"id"`
}

type InboundMessage struct {
	Mid         string       `json:"mid"`
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	ReplyTo     *ReplyTo     `json:"reply_to,omitempty"`
	IsEcho      bool         `json:"is_echo,omitempty"` // true = sent BY the page, not to it
}

type Attachment struct {
	Type    string            `json:"type"` // image, video, audio, file, template
	Payload AttachmentPayload `json:"payload"`
}

type AttachmentPayload struct {
	URL string `json:"url,omitempty"`
}

type ReplyTo struct {
	Mid string `json:"mid"`
}

type Postback struct {
	Mid     string `json:"mid"`
	Title   string `json:"title"`
	Payload string `json:"payload"`
}

type Delivery struct {
	Mids      []string `json:"mids"`
	Watermark int64    `json:"watermark"`
}

type Read struct {
	Watermark int64 `json:"watermark"`
}

// ── changes (comments, feed) ─────────────────────────────────

type ChangeEvent struct {
	Field string      `json:"field"` // "feed", "comments", "mentions"
	Value ChangeValue `json:"value"`
}

type ChangeValue struct {
	From        *WebhookFrom `json:"from,omitempty"`
	Item        string       `json:"item,omitempty"` // "comment", "post", "status"
	Verb        string       `json:"verb,omitempty"` // "add", "edit", "remove"
	CommentID   string       `json:"comment_id,omitempty"`
	PostID      string       `json:"post_id,omitempty"`
	Message     string       `json:"message,omitempty"` // comment text
	CreatedTime int64        `json:"created_time,omitempty"`
}

type WebhookFrom struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}
