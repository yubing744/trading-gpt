package coze

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	baseURL := "https://api.coze.com"
	apiKey := "test-api-key"
	timeout := 10 * time.Second

	client := NewClient(baseURL, apiKey, WithTimeout(timeout)).(*Client)

	assert.Equal(t, baseURL, client.BaseURL)
	assert.Equal(t, apiKey, client.APIKey)
	assert.Equal(t, timeout, client.HTTPClient.Timeout)
}

// Mock server response for a successful chat request
func mockChatHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	response := ChatResponse{
		ConversationID: "123",
		Code:           200,
		Msg:            "Success",
		Messages: []Message{
			{
				Role:        "assistant",
				Content:     "Hello! How can I assist you today?",
				ContentType: "text",
			},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func TestChatSuccess(t *testing.T) {
	// Set up the mock server
	server := httptest.NewServer(http.HandlerFunc(mockChatHandler))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key", WithTimeout(time.Second*10)).(*Client)

	req := &ChatRequest{
		BotID:          "bot-id",
		ConversationID: "123",
		User:           "test-user",
		Query:          "Hello!",
	}

	ctx := context.Background()
	resp, err := client.Chat(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "123", resp.ConversationID)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "Success", resp.Msg)
	assert.Len(t, resp.Messages, 1)
	assert.Equal(t, "assistant", resp.Messages[0].Role)
	assert.Equal(t, "Hello! How can I assist you today?", resp.Messages[0].Content)
	assert.Equal(t, "text", resp.Messages[0].ContentType)
}

// Mock server response for a failed chat request
func mockChatHandlerFail(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
}

func TestChatFailure(t *testing.T) {
	// Set up the mock server
	server := httptest.NewServer(http.HandlerFunc(mockChatHandlerFail))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key", WithTimeout(time.Second*10)).(*Client)

	req := &ChatRequest{
		// Simulated request data
	}

	ctx := context.Background()
	resp, err := client.Chat(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestChatTimeout(t *testing.T) {
	// Set up the mock server with a handler that delays the response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Delay longer than client timeout
		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	client := NewClient(server.URL, "test-api-key", WithTimeout(time.Second*1)).(*Client)

	req := &ChatRequest{
		// Simulated request data
	}

	ctx := context.Background()
	resp, err := client.Chat(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Mock server response for a successful workflow request
func mockWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	response := WorkflowResponse{
		Data:     `{"output":"Test workflow output"}`,
		DebugURL: "https://www.coze.com/work_flow?execute_id=123&space_id=456&workflow_id=789",
	}
	json.NewEncoder(w).Encode(response)
}

func TestRunWorkflowSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mockWorkflowHandler))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key", WithTimeout(time.Second*10)).(*Client)

	req := &WorkflowRequest{
		WorkflowID: "test-workflow-id",
		BotID:      "test-bot-id",
		Parameters: map[string]string{
			"input": "test input",
		},
	}

	ctx := context.Background()
	resp, err := client.RunWorkflow(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Data, "Test workflow output")
	assert.Contains(t, resp.DebugURL, "work_flow")
}

func TestRunWorkflowTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key", WithTimeout(time.Second*1)).(*Client)

	req := &WorkflowRequest{
		WorkflowID: "test-workflow-id",
	}

	ctx := context.Background()
	resp, err := client.RunWorkflow(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "workflow execution timed out")
}

func TestRunWorkflowFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mockChatHandlerFail))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key", WithTimeout(time.Second*10)).(*Client)

	req := &WorkflowRequest{
		WorkflowID: "test-workflow-id",
	}

	ctx := context.Background()
	resp, err := client.RunWorkflow(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}
