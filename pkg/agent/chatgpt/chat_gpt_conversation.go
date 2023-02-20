package chatgpt

import "github.com/google/uuid"

type ChatGPTConversation struct {
	id              string
	parentMessageID string
}

func NewChatGPTConversation() *ChatGPTConversation {
	return &ChatGPTConversation{
		id:              "",
		parentMessageID: uuid.NewString(),
	}
}

func (conv *ChatGPTConversation) IsNew() bool {
	return conv.id == ""
}

func (conv *ChatGPTConversation) GetID() string {
	return conv.id
}

func (conv *ChatGPTConversation) GetParentMessageID() string {
	return conv.parentMessageID
}

func (conv *ChatGPTConversation) GetIDRef() *string {
	if conv.IsNew() {
		return nil
	} else {
		return &conv.id
	}
}

func (conv *ChatGPTConversation) GetParentMessageIDRef() *string {
	return &conv.parentMessageID
}

func (conv *ChatGPTConversation) Update(id string, parentId string) {
	conv.id = id
	conv.parentMessageID = parentId
}
