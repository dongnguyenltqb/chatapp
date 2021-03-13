// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"chatapp/infra"
	"chatapp/logger"
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	logger := logger.Get()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err)
		return
	}
	client := &Client{
		hub:   hub,
		conn:  conn,
		id:    uuid.New().String(),
		send:  make(chan []byte, 256),
		rchan: make(chan wsRoomActionMessage, 100),
		rooms: make(map[string]bool),
	}
	client.rooms[client.id] = true
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.roomPump()
	go client.writePump()
	go client.readPump()
	client.sendIdentityMsg()
}

// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	sync.Mutex
	// Registered clients.
	clients map[*Client]bool

	// Outbound messages
	broadcast chan []byte

	// Outbound message for a specific room
	room chan wsMessageForRoom

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Pubsub channel
	psbcchan         string
	psroomchan       string
	subroomchan      <-chan *redis.Message
	subbroadcastchan <-chan *redis.Message
}

// NewHub return a new hub for a specific enpoint
func NewHub() *Hub {
	psbcchan := "chatappbroadcastchan"
	psroomchan := "chatapproomchan"
	subroom := infra.GetRedis().Subscribe(context.Background(), "chatapproomchan")
	subbroadcast := infra.GetRedis().Subscribe(context.Background(), "chatapproomchan")
	return &Hub{
		broadcast:        make(chan []byte),
		room:             make(chan wsMessageForRoom),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		clients:          make(map[*Client]bool),
		psbcchan:         psbcchan,
		psroomchan:       psroomchan,
		subroomchan:      subroom.Channel(),
		subbroadcastchan: subbroadcast.Channel(),
	}
}

// Run the hub
func (h *Hub) Run() {
	logger := logger.Get()
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				go client.clean()
			}
		case message := <-h.broadcast:
			infra.GetRedis().Publish(context.Background(), h.psbcchan, message)
		case message := <-h.room:
			b, _ := json.Marshal(message)
			if err := infra.GetRedis().Publish(context.Background(), h.psroomchan, b).Err(); err != nil {
				logger.Error(err)
			}
		// Two pubsub channel for receiving message from other node
		case message := <-h.subroomchan:
			m := wsMessageForRoom{}
			if err := json.Unmarshal([]byte(message.Payload), &m); err != nil {
				logger.Error(err)
			}
			for client := range h.clients {
				h.Lock()
				ok := client.rooms[m.RoomId]
				if ok {
					select {
					case client.send <- m.Message:
					default:
						delete(h.clients, client)
						go client.clean()
					}
				}
				h.Unlock()
			}
		case message := <-h.subbroadcastchan:
			for client := range h.clients {
				select {
				case client.send <- []byte(message.Payload):
				default:
					delete(h.clients, client)
					go client.clean()
				}
			}
		}
	}
}
