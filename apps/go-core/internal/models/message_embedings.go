package models

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

type MessageEmbedding struct {
	ID             string          `db:"id" json:"id"`
	MessageID      string          `db:"message_id" json:"message_id"`
	BusinessID     string          `db:"business_id" json:"business_id"`
	ConversationID string          `db:"conversation_id" json:"conversation_id"`
	Content        string          `db:"content" json:"content"`
	Embedding      pgvector.Vector `db:"embedding" json:"-"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}

type CreateMessageEmbedding struct {
	MessageID      string          `db:"message_id"`
	BusinessID     string          `db:"business_id"`
	ConversationID string          `db:"conversation_id"`
	Content        string          `db:"content"`
	Embedding      pgvector.Vector `db:"embedding"`
}

type MessageEmbeddingResponse struct {
	chunk_index int
	content     string
	embedding   pgvector.Vector
}

type ChatEmbedResponse struct {
	Embeddings []EmbeddedChunk `json:"embeddings"`
}

type PastChatChunk struct {
	Content        string  `db:"content" json:"content"`
	ConversationID string  `db:"conversation_id" json:"conversation_id"`
	Score          float64 `db:"score" json:"score"`
}

type KnowledgeChunk struct {
	Content    string  `db:"content" json:"content"`
	SourceType string  `db:"source_type" json:"source_type"`
	SourceID   string  `db:"source_id" json:"source_id"`
	Score      float64 `db:"score" json:"score"`
}
