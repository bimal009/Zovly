package models

// ─── platform constants ───────────────────────────────────────────────────────

type Platform string

const (
	PlatformFacebook  Platform = "facebook"
	PlatformInstagram Platform = "instagram"
	PlatformWhatsApp  Platform = "whatsapp"
)

// ═════════════════════════════════════════════════════════════════════════════
// WhatsApp Business Account
// object: "whatsapp_business_account"
// ═════════════════════════════════════════════════════════════════════════════

type WhatsAppWebhookPayload struct {
	Object string          `json:"object" db:"object"`
	Entry  []WhatsAppEntry `json:"entry"  db:"entry"`
}

type WhatsAppEntry struct {
	ID      string           `json:"id"      db:"id"`
	Changes []WhatsAppChange `json:"changes" db:"changes"`
}

type WhatsAppChange struct {
	Field string        `json:"field" db:"field"`
	Value WhatsAppValue `json:"value" db:"value"`
}

type WhatsAppValue struct {
	MessagingProduct string            `json:"messaging_product" db:"messaging_product"`
	Metadata         WhatsAppMetadata  `json:"metadata"          db:"metadata"`
	Contacts         []WhatsAppContact `json:"contacts"          db:"contacts"`
	Messages         []WhatsAppMessage `json:"messages"          db:"messages"`
}

type WhatsAppMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number" db:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"      db:"phone_number_id"`
}

type WhatsAppContact struct {
	Profile WhatsAppProfile `json:"profile" db:"profile"`
	WaID    string          `json:"wa_id"   db:"wa_id"`
}

type WhatsAppProfile struct {
	Name string `json:"name" db:"name"`
}

// ── WhatsApp message types ────────────────────────────────────────────────────

type WhatsAppMessageType string

const (
	WhatsAppMsgTypeText     WhatsAppMessageType = "text"
	WhatsAppMsgTypeImage    WhatsAppMessageType = "image"
	WhatsAppMsgTypeVideo    WhatsAppMessageType = "video"
	WhatsAppMsgTypeAudio    WhatsAppMessageType = "audio"
	WhatsAppMsgTypeDocument WhatsAppMessageType = "document"
)

type WhatsAppMessage struct {
	From      string              `json:"from"               db:"from"`
	ID        string              `json:"id"                 db:"id"`
	Timestamp string              `json:"timestamp"          db:"timestamp"`
	Type      WhatsAppMessageType `json:"type"               db:"type"`
	Text      *WhatsAppText       `json:"text,omitempty"     db:"text"`
	Image     *WhatsAppMediaItem  `json:"image,omitempty"    db:"image"`
	Video     *WhatsAppMediaItem  `json:"video,omitempty"    db:"video"`
	Audio     *WhatsAppAudio      `json:"audio,omitempty"    db:"audio"`
	Document  *WhatsAppDocument   `json:"document,omitempty" db:"document"`
}

type WhatsAppText struct {
	Body string `json:"body" db:"body"`
}

// Image and Video share the same shape (caption + sha256 optional).
type WhatsAppMediaItem struct {
	ID       string `json:"id"                db:"id"`
	MimeType string `json:"mime_type"         db:"mime_type"`
	Caption  string `json:"caption,omitempty" db:"caption"`
	Sha256   string `json:"sha256,omitempty"  db:"sha256"`
}

// Audio has no caption.
type WhatsAppAudio struct {
	ID       string `json:"id"        db:"id"`
	MimeType string `json:"mime_type" db:"mime_type"`
}

type WhatsAppDocument struct {
	ID       string `json:"id"       db:"id"`
	Filename string `json:"filename" db:"filename"`
	MimeType string `json:"mime_type" db:"mime_type"`
}
