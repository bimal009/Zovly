package models

type ChatContext struct {
	Knowledge        []KnowledgeChunk `json:"knowledge"`
	PastChats        []PastChatChunk  `json:"past_chats"`
	Business         *Business        `json:"business"`
	PastConversation []Message        `json:"past_conversation"`
	Customer         *Conversation    `json:"customer"`
	ImageMessage     string           `json:"image_message"`
	AudioMessage     string           `json:"audio_message"`
}

type ChatReplyPayload struct {
	BusinessID string      `json:"business_id"`
	Message    string      `json:"message"`
	Context    ChatContext `json:"context"`
}

type ChatReplyResponse struct {
	Reply string `json:"reply"`
}
