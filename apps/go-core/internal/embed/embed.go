// Package embed is a small client for the AI service's embedding endpoints.
// It lives outside service/ and workers/ so both can depend on it without an
// import cycle.
package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bimal009/Zovly/internal/models"
	"github.com/pgvector/pgvector-go"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// EmbedChat returns the embedding vector for a single chat message. kind selects
// the embedding variant the AI service should use ("query", "passage", ...).
func (c *Client) EmbedChat(ctx context.Context, message, kind string) (pgvector.Vector, error) {
	body, err := json.Marshal(map[string]string{"message": message, "kind": kind})
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("marshal embed request: %w", err)
	}

	resp, err := c.post(ctx, "/api/v1/ml/chat/embed", body)
	if err != nil {
		return pgvector.Vector{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return pgvector.Vector{}, fmt.Errorf("embed service returned status %d", resp.StatusCode)
	}

	var result models.ChatEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return pgvector.Vector{}, fmt.Errorf("decode embed response: %w", err)
	}
	if len(result.Embeddings) == 0 {
		return pgvector.Vector{}, fmt.Errorf("embed service returned no embeddings")
	}
	return pgvector.NewVector(result.Embeddings[0].Embedding), nil
}

func (c *Client) EmbedFaq(ctx context.Context, question, answer string) ([]models.FaqChunksResponse, error) {
	body, err := json.Marshal(map[string]string{"question": question, "answer": answer})
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	resp, err := c.post(ctx, "/api/v1/ml/embed/faq", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed service returned status %d", resp.StatusCode)
	}

	var result []models.FaqChunksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	return result, nil
}

func (c *Client) post(ctx context.Context, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call embed service: %w", err)
	}
	return resp, nil
}
