// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"chatapp/infra"
	"chatapp/util/logger"
	"context"
	"encoding/json"
	"net/http"
	"os"
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
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

var hub *Hub
var onceInitHub sync.Once

// GetHub return singleton hub
func GetHub() *Hub {
	onceInitHub.Do(func() {
		psbcchan := "chatappbroadcastchan"
		psroomchan := "chatapproomchan"
		subroom := infra.GetRedis().Subscribe(context.Background(), "chatapproomchan")
		subbroadcast := infra.GetRedis().Subscribe(context.Background(), "chatapproomchan")
		hub = &Hub{
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
		go hub.run()
	})
	return hub
}

func (h *Hub) run() {
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
			for client := range h.clients {
				ok := client.rooms[message.RoomId]
				if ok {
					select {
					case client.send <- message.Message:
					default:
						delete(h.clients, client)
						go client.clean()
					}
				}
			}
		// Two pubsub channel for receiving message from other node
		case message := <-h.subroomchan:
			m := wsMessageForRoom{}
			if err := json.Unmarshal([]byte(message.Payload), &m); err != nil {
				logger.Error(err)
			}
			if m.AppName != os.Getenv("app_name") {
				for client := range h.clients {
					ok := client.rooms[m.RoomId]
					if ok {
						select {
						case client.send <- m.Message:
						default:
							delete(h.clients, client)
							go client.clean()
						}
					}
				}
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
