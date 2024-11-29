package base

import "testing"

func TestMsgBus(t *testing.T) {
	MessageBus.On("test", func(msg any) {
		t.Log(msg)
	})
	MessageBus.Emit("test", "hello")
}
