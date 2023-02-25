package types

type IChannel interface {
	INotifyChannel

	OnMessage(cb MessageCallback)
}
