package ws

import (
	"app/store"
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nhnghia272/gopkg"
)

var (
	server   *Ws
	upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024, CheckOrigin: func(r *http.Request) bool { return true }}
)

type payload struct {
	Data      []byte   `json:"data"`
	Users     []string `json:"users"`
	Broadcast bool     `json:"broadcast"`
}

type Ws struct {
	sync.RWMutex
	store *store.Store
	users gopkg.CacheShard[[]*Conn]
}

func New(store *store.Store) *Ws {
	server = &Ws{store: store, users: gopkg.NewCacheShard[[]*Conn](64)}

	gopkg.Async().Go(func() {
		for c := range store.Rdb.Subscribe(context.Background(), "wss").Channel() {
			var p payload
			json.Unmarshal([]byte(c.Payload), &p)
			if p.Broadcast {
				p.Users = server.users.Keys()
			}
			gopkg.LoopFunc(p.Users, func(user string) {
				conns, _ := server.users.Get(user)
				gopkg.LoopFunc(conns, func(conn *Conn) { conn.Emit(p.Data) })
			})
		}
	})

	return server
}

func (s *Ws) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return upgrader.Upgrade(w, r, nil)
}

func (s *Ws) Broadcast(data []byte) {
	data, _ = json.Marshal(payload{Data: data, Broadcast: true})
	s.store.Rdb.Publish(context.Background(), "wss", data)
}

func (s *Ws) EmitTo(users []string, data []byte) {
	data, _ = json.Marshal(payload{Data: data, Users: users})
	s.store.Rdb.Publish(context.Background(), "wss", data)
}

func (s *Ws) Register(conn *Conn) {
	s.Lock()
	defer s.Unlock()

	conns, _ := s.users.Get(conn.session.UserId)
	s.users.Set(conn.session.UserId, append(conns, conn), -1)
}

func (s *Ws) Unregister(conn *Conn) {
	s.Lock()
	defer s.Unlock()

	conns, _ := s.users.Get(conn.session.UserId)
	s.users.Set(conn.session.UserId, slices.DeleteFunc(conns, func(e *Conn) bool { return e.conn == conn.conn }), -1)
}
