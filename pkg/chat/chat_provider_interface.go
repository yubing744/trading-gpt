package chat

type ChatProvider interface {
	GetName() string
	Listen() error
}
