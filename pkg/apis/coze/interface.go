package coze

import (
	"context"
	"fmt"
)

// Common error codes
const (
	ErrCodeUnauthorized   = 401
	ErrCodeForbidden      = 403
	ErrCodeRateLimit      = 429
	ErrCodeInternalError  = 500
	ErrCodeGatewayTimeout = 504
)

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: status=%d, message=%s", e.StatusCode, e.Message)
}

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

// WorkflowRequest represents the payload for running a workflow
type WorkflowRequest struct {
	WorkflowID string            `json:"workflow_id"`
	BotID      string            `json:"bot_id,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// WorkflowResponse represents the response from running a workflow
type WorkflowResponse struct {
	Cost     string `json:"cost"`
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Data     string `json:"data"`
	DebugURL string `json:"debug_url"`
	Token    int    `json:"token"`
}

// CozeError represents an error response from the Coze API
type CozeError struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func (e *CozeError) Error() string {
	return fmt.Sprintf("coze api error: code=%d, message=%s", e.Code, e.Message)
}

// CozeClient is an interface for Coze API client
type ICozeClient interface {
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	RunWorkflow(ctx context.Context, req *WorkflowRequest) (*WorkflowResponse, error)
}
