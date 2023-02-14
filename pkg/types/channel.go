package types

type MessageCallback func(msg *Message)

type Channel interface {
	OnMessage(cb MessageCallback)
	Reply(msg *Message) error
}
