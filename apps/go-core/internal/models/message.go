package models

import "time"

type MessageDirection string

const (
	MessageDirectionIn  MessageDirection = "in"
	MessageDirectionOut MessageDirection = "out"
)

type MessageSender string

const (
	MessageSenderAI    MessageSender = "ai"
	MessageSenderHuman MessageSender = "human"
)

type MessageMediaType string

const (
	MediaTypeImage    MessageMediaType = "image"
	MediaTypeVideo    MessageMediaType = "video"
	MediaTypeAudio    MessageMediaType = "audio"
	MediaTypeDocument MessageMediaType = "document"
	MediaTypeLink     MessageMediaType = "link"
)

type MessageStatus string

const (
	MessageStatusPending MessageStatus = "pending"
	MessageStatusSent    MessageStatus = "sent"
	MessageStatusFailed  MessageStatus = "failed"
	MessageStatusSkipped MessageStatus = "skipped"
)

type Message struct {
	ID             string `db:"id" json:"id"`
	ConversationID string `db:"conversation_id" json:"conversation_id"`
	BusinessID     string `db:"business_id" json:"business_id"`

	Direction MessageDirection `db:"direction" json:"direction"`
	SentBy    *MessageSender   `db:"sent_by" json:"sent_by"` // null for inbound

	Content *string `db:"content" json:"content"`

	MediaUrl  *string           `db:"media_url" json:"media_url"`
	MediaType *MessageMediaType `db:"media_type" json:"media_type"`

	IsVectorized bool `db:"is_vectorized" json:"is_vectorized"`

	Status            *MessageStatus `db:"status" json:"status"` // null for inbound
	ErrorMessage      *string        `db:"error_message" json:"error_message"`
	SentToPlatformAt  *time.Time     `db:"sent_to_platform_at" json:"sent_to_platform_at"`
	PlatformMessageID *string        `db:"platform_message_id" json:"platform_message_id"`

	SentAt time.Time `db:"sent_at" json:"sent_at"`
}

type CreateMessage struct {
	ConversationID string            `db:"conversation_id"`
	BusinessID     string            `db:"business_id"`
	Direction      MessageDirection  `db:"direction"`
	SentBy         *MessageSender    `db:"sent_by"`
	Content        *string           `db:"content"`
	MediaUrl       *string           `db:"media_url"`
	MediaType      *MessageMediaType `db:"media_type"`
	Status         *MessageStatus    `db:"status"`
}
