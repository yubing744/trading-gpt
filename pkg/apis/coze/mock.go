package coze

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockCozeClient is a mock type for the ICozeClient interface
type MockCozeClient struct {
	mock.Mock
}

// Chat mocks the Chat method of ICozeClient
func (m *MockCozeClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*ChatResponse), args.Error(1)
}
