package coze

import "context"

// Message represents an item in the chat
type Message struct {
	Role        string `json:"role"`
	Type        string `json:"type,omitempty"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

// ChatRequest represents the payload for sending a chat request
type ChatRequest struct {
	BotID           string            `json:"bot_id"`
	ConversationID  string            `json:"conversation_id"`
	User            string            `json:"user"`
	Query           string            `json:"query,omitempty"`
	ChatHistory     []*Message        `json:"chat_history,omitempty"`
	Stream          bool              `json:"stream"`
	CustomVariables map[string]string `json:"custom_variables,omitempty"`
}

// ChatResponse represents the response from sending a message
type ChatResponse struct {
	ConversationID string    `json:"conversation_id"`
	Code           int       `json:"code"`
	Msg            string    `json:"msg"`
	Messages       []Message `json:"messages,omitempty"`
}

// CozeClient is an interface for Coze API client
type ICozeClient interface {
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}
