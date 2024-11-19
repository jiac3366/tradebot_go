package main

import (
	"fmt"
	"net"
	"sync"
)

// Reactor 模式的核心组件
type ChatServer struct {
	listener    net.Listener
	dispatcher  *EventDispatcher
	handlers    map[net.Conn]*ClientHandler
	handlerLock sync.RWMutex
}

// 事件分发器
type EventDispatcher struct {
	events chan Event
}

// 事件类型
type Event struct {
	conn      net.Conn
	data      []byte
	eventType string // "connect", "message", "disconnect"
}

// 客户端处理器
type ClientHandler struct {
	conn     net.Conn
	server   *ChatServer
	incoming chan []byte
}

// 创建新的聊天服务器
func NewChatServer(address string) (*ChatServer, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	server := &ChatServer{
		listener: listener,
		dispatcher: &EventDispatcher{
			events: make(chan Event, 100),
		},
		handlers: make(map[net.Conn]*ClientHandler),
	}

	return server, nil
}

// 启动服务器
func (s *ChatServer) Start() {
	// 1. 启动事件分发循环
	go s.dispatcher.dispatch(s)

	// 2. 接受新连接
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}

		// 3. 为新连接创建处理器
		handler := &ClientHandler{
			conn:     conn,
			server:   s,
			incoming: make(chan []byte, 100),
		}

		s.handlerLock.Lock()
		s.handlers[conn] = handler
		s.handlerLock.Unlock()

		// 4. 启动处理器
		go handler.handle()
	}
}

// 事件分发循环
func (d *EventDispatcher) dispatch(server *ChatServer) {
	for event := range d.events {
		switch event.eventType {
		case "message":
			// 广播消息给所有客户端
			server.broadcast(event.data, event.conn)
		case "disconnect":
			server.removeClient(event.conn)
		}
	}
}

// 广播消息
func (s *ChatServer) broadcast(message []byte, sender net.Conn) {
	s.handlerLock.RLock()
	defer s.handlerLock.RUnlock()

	for conn, handler := range s.handlers {
		if conn != sender { // 不发送给消息发送者
			handler.incoming <- message
		}
	}
}

// 移除客户端
func (s *ChatServer) removeClient(conn net.Conn) {
	s.handlerLock.Lock()
	defer s.handlerLock.Unlock()

	if handler, ok := s.handlers[conn]; ok {
		close(handler.incoming)
		delete(s.handlers, conn)
	}
}

// 处理客户端连接
func (h *ClientHandler) handle() {
	defer h.conn.Close()

	// 读取循环
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := h.conn.Read(buffer)
			if err != nil {
				h.server.dispatcher.events <- Event{
					conn:      h.conn,
					eventType: "disconnect",
				}
				return
			}

			// 将消息发送给分发器
			h.server.dispatcher.events <- Event{
				conn:      h.conn,
				data:      buffer[:n],
				eventType: "message",
			}
		}
	}()

	// 写入循环
	for message := range h.incoming {
		_, err := h.conn.Write(message)
		if err != nil {
			return
		}
	}
}

func main() {
	server, err := NewChatServer(":8080")
	if err != nil {
		panic(err)
	}

	fmt.Println("Chat server started on :8080")
	server.Start()
}
