package models

import (
	"encoding/json"

	pgvector "github.com/pgvector/pgvector-go"
)

type KnowledgeSourceType string

const (
	SourceFaq     KnowledgeSourceType = "faq"
	SourcePolicy  KnowledgeSourceType = "policy"
	SourcePost    KnowledgeSourceType = "post"
	SourceProduct KnowledgeSourceType = "product"
)

type CreateKnowledgeChunk struct {
	BusinessID string              `db:"business_id"`
	SourceType KnowledgeSourceType `db:"source_type"`
	SourceID   string              `db:"source_id"`
	ChunkIndex int                 `db:"chunk_index"`
	Content    string              `db:"content"`
	Embedding  pgvector.Vector     `db:"embedding"`
	Metadata   json.RawMessage     `db:"metadata"`
}

type EmbeddedChunk struct {
	ChunkIndex int       `json:"chunk_index"`
	Content    string    `json:"content"`
	Embedding  []float32 `json:"embedding"`
}

func ToChunkInserts(
	chunks []EmbeddedChunk,
	businessID, sourceID string,
	sourceType KnowledgeSourceType,
	metadata json.RawMessage,
) []CreateKnowledgeChunk {
	inserts := make([]CreateKnowledgeChunk, len(chunks))
	for i, c := range chunks {
		inserts[i] = CreateKnowledgeChunk{
			BusinessID: businessID,
			SourceType: sourceType,
			SourceID:   sourceID,
			ChunkIndex: c.ChunkIndex,
			Content:    c.Content,
			Embedding:  pgvector.NewVector(c.Embedding),
			Metadata:   metadata,
		}
	}
	return inserts
}
