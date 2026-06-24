package models

// ─── Instagram OAuth ──────────────────────────────────────────────────────────

type IgLongLivedResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // seconds, ~60 days
}

// ─── Instagram webhook ────────────────────────────────────────────────────────
// object: "instagram"

type InstagramWebhookPayload struct {
	Object string           `json:"object" db:"object"`
	Entry  []InstagramEntry `json:"entry"  db:"entry"`
}

type InstagramEntry struct {
	ID        string                    `json:"id,omitempty"        db:"id"`
	Time      int64                     `json:"time,omitempty"      db:"time"`
	Messaging []InstagramMessagingEvent `json:"messaging,omitempty" db:"messaging"`
	Changes   []InstagramChangeEvent    `json:"changes,omitempty"   db:"changes"`
}

// ─── Instagram DM ─────────────────────────────────────────────────────────────

type InstagramMessagingEvent struct {
	Sender    InstagramUser     `json:"sender"             db:"sender"`
	Recipient InstagramUser     `json:"recipient"          db:"recipient"`
	Timestamp int64             `json:"timestamp"          db:"timestamp"`
	Message   *InstagramMessage `json:"message,omitempty"  db:"message"`
}
type InstagramUser struct {
	ID             string `json:"id"                  db:"id"`
	UserID         string `json:"user_id"             db:"user_id"` // IG professional account ID — matches webhook entry.id
	Username       string `json:"username"            db:"username"`
	ProfilePic     string `json:"profile_picture_url" db:"profile_pic"`
	Name           string `json:"name"                db:"name"`
	FollowersCount int    `json:"followers_count"     db:"followers_count"`
	FollowsCount   int    `json:"follows_count"       db:"follows_count"`
	MediaCount     int    `json:"media_count"         db:"media_count"`
}

type InstagramAttachmentType string

const (
	InstagramAttachmentTypeImage InstagramAttachmentType = "image"
	InstagramAttachmentTypeVideo InstagramAttachmentType = "video"
	InstagramAttachmentTypeAudio InstagramAttachmentType = "audio"
	InstagramAttachmentTypeFile  InstagramAttachmentType = "file"
	InstagramAttachmentTypeURL   InstagramAttachmentType = "url"
	InstagramAttachmentTypeLink  InstagramAttachmentType = "link"
	InstagramAttachmentTypeReel  InstagramAttachmentType = "ig_reel"
	InstagramAttachmentTypePost  InstagramAttachmentType = "ig_post"
	InstagramAttachmentTypeStory InstagramAttachmentType = "story_mention"
)

type InstagramMessage struct {
	Mid         string                `json:"mid"                   db:"mid"`
	Text        string                `json:"text,omitempty"        db:"text"`
	Attachments []InstagramAttachment `json:"attachments,omitempty" db:"attachments"`
	IsEcho      bool                  `json:"is_echo,omitempty"     db:"is_echo"`
}

type InstagramAttachment struct {
	Type    InstagramAttachmentType    `json:"type"    db:"type"`
	Payload InstagramAttachmentPayload `json:"payload" db:"payload"`
}

type InstagramAttachmentPayload struct {
	URL string `json:"url,omitempty" db:"url"`
	// Title carries the caption Instagram includes for shared reels/posts.
	Title string `json:"title,omitempty" db:"title"`
}

// ─── Instagram changes (comments / mentions / story) ─────────────────────────

type InstagramChangeEvent struct {
	Field string               `json:"field" db:"field"`
	Value InstagramChangeValue `json:"value" db:"value"`
}

// Discriminate on InstagramChangeEvent.Field:
//
//	"comments"       → ID, Text, From, Media
//	"mentions"       → MediaID, CommentID, Username
//	"story_insights" → MediaID, MentionID
type InstagramChangeValue struct {
	// comments
	ID    string                 `json:"id,omitempty"    db:"id"`
	Text  string                 `json:"text,omitempty"  db:"text"`
	From  *InstagramCommentFrom  `json:"from,omitempty"  db:"from"`
	Media *InstagramCommentMedia `json:"media,omitempty" db:"media"`

	// mentions
	MediaID   string `json:"media_id,omitempty"   db:"media_id"`
	CommentID string `json:"comment_id,omitempty" db:"comment_id"`
	Username  string `json:"username,omitempty"   db:"username"`

	// story_insights
	MentionID string `json:"mention_id,omitempty" db:"mention_id"`
}

type InstagramCommentFrom struct {
	ID       string `json:"id"                 db:"id"`
	Username string `json:"username,omitempty" db:"username"`
}

type InstagramCommentMedia struct {
	ID string `json:"id" db:"id"`
}
