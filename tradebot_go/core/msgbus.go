package base

var MessageBus = &MsgBus{
	listeners: make(map[string][]func(msg any)),
}

type MsgBus struct {
	listeners map[string][]func(msg any)
}

func (mb *MsgBus) On(event string, callback func(msg any)) {
	mb.listeners[event] = append(mb.listeners[event], callback)
}

func (mb *MsgBus) Emit(event string, msg any) {
	for _, listener := range mb.listeners[event] {
		listener(msg)
	}
}
