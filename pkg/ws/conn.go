package ws

import (
	"app/store"
	"app/store/db"
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type message struct {
	Type int
	Data []byte
}

type Conn struct {
	sync.RWMutex
	conn    *websocket.Conn
	store   *store.Store
	session *db.AuthSessionDto
	done    chan struct{}
	queue   chan message
}

func NewConn(conn *websocket.Conn, store *store.Store, session *db.AuthSessionDto) {
	s := &Conn{conn: conn, store: store, session: session, done: make(chan struct{}, 1), queue: make(chan message, 100)}
	s.Run()
}

func (s *Conn) Run() {
	server.Register(s)
	ctx, cancelFunc := context.WithCancel(context.Background())

	go s.Pong(ctx)
	go s.Read(ctx)
	go s.Send(ctx)

	<-s.done

	cancelFunc()
}

func (s *Conn) Disconnected() {
	s.Lock()
	defer s.Unlock()

	close(s.done)
	server.Unregister(s)
}

func (s *Conn) Broadcast(data []byte) {
	server.Broadcast(data)
}

func (s *Conn) EmitTo(users []string, data []byte) {
	server.EmitTo(users, data)
}

func (s *Conn) Emit(data []byte) {
	s.queue <- message{websocket.TextMessage, data}
}

func (s *Conn) Pong(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.queue <- message{websocket.PongMessage, []byte{}}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Conn) Send(ctx context.Context) {
	for {
		select {
		case msg := <-s.queue:
			s.RLock()
			err := s.conn.WriteMessage(msg.Type, msg.Data)
			s.RUnlock()

			if err != nil {
				s.Disconnected()
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Conn) Read(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.RLock()
			typ, data, err := s.conn.ReadMessage()
			s.RUnlock()

			if err != nil || typ == websocket.CloseMessage {
				s.Disconnected()
				return
			}

			if typ == websocket.PingMessage {
				s.queue <- message{websocket.PongMessage, []byte{}}
				continue
			}

			s.OnMessage(data)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Conn) OnMessage(data []byte) {
	s.Emit(data)
}
