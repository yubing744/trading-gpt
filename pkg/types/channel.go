package types

type Channel interface {
	OnMessage(cb MessageCallback)
	Reply(msg *Message) error
}
