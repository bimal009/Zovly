package task

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const (
	TypeChatReply = "chat:reply"
)

type ChatReplyPayload struct {
	BusinessID     string `json:"business_id"`
	ConversationID string `json:"conversation_id"`
}

func DebounceTaskID(conversationID string) string {
	return "chat:reply:" + conversationID
}

func NewChatReplyTask(businessID, conversationID string) (*asynq.Task, error) {
	if businessID == "" || conversationID == "" {
		return nil, fmt.Errorf("chat reply task: business_id and conversation_id are required")
	}
	payload, err := json.Marshal(ChatReplyPayload{
		BusinessID:     businessID,
		ConversationID: conversationID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal chat reply payload: %w", err)
	}
	return asynq.NewTask(TypeChatReply, payload), nil
}

func ParseChatReplyPayload(t *asynq.Task) (ChatReplyPayload, error) {
	var p ChatReplyPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return p, fmt.Errorf("unmarshal chat reply payload: %w", err)
	}
	return p, nil
}
