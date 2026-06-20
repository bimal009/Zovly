package models

// ─── Facebook (Messenger) webhook ────────────────────────────────────────────
// object: "page"

type FacebookWebhookPayload struct {
	Object string          `json:"object" db:"object"`
	Entry  []FacebookEntry `json:"entry"  db:"entry"`
}

type FacebookEntry struct {
	ID        string                   `json:"id"                  db:"id"`
	Time      int64                    `json:"time"                db:"time"`
	Messaging []FacebookMessagingEvent `json:"messaging,omitempty" db:"messaging"`
	Changes   []FacebookChangeEvent    `json:"changes,omitempty"   db:"changes"`
}

type FacebookMessagingEvent struct {
	Sender    FacebookUser      `json:"sender"             db:"sender"`
	Recipient FacebookUser      `json:"recipient"          db:"recipient"`
	Timestamp int64             `json:"timestamp"          db:"timestamp"`
	Message   *FacebookMessage  `json:"message,omitempty"  db:"message"`
	Postback  *FacebookPostback `json:"postback,omitempty" db:"postback"`
	Delivery  *FacebookDelivery `json:"delivery,omitempty" db:"delivery"`
	Read      *FacebookRead     `json:"read,omitempty"     db:"read"`
}

type FacebookUser struct {
	ID string `json:"id" db:"id"`
}

// ── attachment types ──────────────────────────────────────────────────────────

type FacebookAttachmentType string

const (
	FacebookAttachmentTypeImage    FacebookAttachmentType = "image"
	FacebookAttachmentTypeVideo    FacebookAttachmentType = "video"
	FacebookAttachmentTypeAudio    FacebookAttachmentType = "audio"
	FacebookAttachmentTypeFile     FacebookAttachmentType = "file"
	FacebookAttachmentTypeTemplate FacebookAttachmentType = "template"
	FacebookAttachmentTypeFallback FacebookAttachmentType = "fallback"
	FacebookAttachmentTypeURL      FacebookAttachmentType = "url"
	FacebookAttachmentTypeLink     FacebookAttachmentType = "link"
)

type FacebookMessage struct {
	Mid         string               `json:"mid"                   db:"mid"`
	Text        string               `json:"text,omitempty"        db:"text"`
	Attachments []FacebookAttachment `json:"attachments,omitempty" db:"attachments"`
	ReplyTo     *FacebookReplyTo     `json:"reply_to,omitempty"    db:"reply_to"`
	IsEcho      bool                 `json:"is_echo,omitempty"     db:"is_echo"`
}

type FacebookAttachment struct {
	Type    FacebookAttachmentType    `json:"type"    db:"type"`
	Payload FacebookAttachmentPayload `json:"payload" db:"payload"`
}

type FacebookAttachmentPayload struct {
	URL string `json:"url,omitempty" db:"url"`
}

type FacebookReplyTo struct {
	Mid string `json:"mid" db:"mid"`
}

type FacebookPostback struct {
	Mid     string `json:"mid"     db:"mid"`
	Title   string `json:"title"   db:"title"`
	Payload string `json:"payload" db:"payload"`
}

type FacebookDelivery struct {
	Mids      []string `json:"mids"      db:"mids"`
	Watermark int64    `json:"watermark" db:"watermark"`
}

type FacebookRead struct {
	Watermark int64 `json:"watermark" db:"watermark"`
}

// ── change / feed events ──────────────────────────────────────────────────────

type FacebookChangeEvent struct {
	Field string              `json:"field" db:"field"`
	Value FacebookChangeValue `json:"value" db:"value"`
}

type FacebookChangeValue struct {
	// feed / comment
	From        *FacebookFrom `json:"from,omitempty"         db:"from"`
	Item        string        `json:"item,omitempty"         db:"item"`
	Verb        string        `json:"verb,omitempty"         db:"verb"`
	CommentID   string        `json:"comment_id,omitempty"   db:"comment_id"`
	PostID      string        `json:"post_id,omitempty"      db:"post_id"`
	Message     string        `json:"message,omitempty"      db:"message"`
	CreatedTime int64         `json:"created_time,omitempty" db:"created_time"`
	// leadgen
	LeadgenID string `json:"leadgen_id,omitempty" db:"leadgen_id"`
	FormID    string `json:"form_id,omitempty"    db:"form_id"`
	PageID    string `json:"page_id,omitempty"    db:"page_id"`
}

type FacebookFrom struct {
	ID   string `json:"id"             db:"id"`
	Name string `json:"name,omitempty" db:"name"`
}

type MessengerProfile struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	ProfilePic string `json:"profile_pic"`
	ID         string `json:"id"`
}
