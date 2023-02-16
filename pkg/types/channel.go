package types

type Channel interface {
	GetID() string
	OnMessage(cb MessageCallback)
	Reply(msg *Message) error
}
