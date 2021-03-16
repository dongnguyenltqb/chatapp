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

// 	serveWs handles websocket requests from the peer.
func (h *Hub) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{
		hub:    h,
		conn:   conn,
		id:     uuid.New().String(),
		send:   make(chan []byte, 256),
		rchan:  make(chan wsRoomActionMessage, 100),
		rooms:  make([]string, maxRoomSize),
		logger: logger.Get(),
	}
	client.rooms = []string{client.id}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.roomPump()
	go client.writePump()
	go client.readPump()
	client.sendIdentityMsg()
	client.broadcastMsg([]byte("xin chao : " + client.id))
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

// getHub return singleton hub
func getHub() *Hub {
	onceInitHub.Do(func() {
		psroomchan := "chat_app_room_chan"
		psbcchan := "chat_app_broadcast_chan"

		subroom := infra.GetRedis().Subscribe(context.Background(), psroomchan)
		subbroadcast := infra.GetRedis().Subscribe(context.Background(), psbcchan)

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

func (h *Hub) sendMsgToRoom(roomId string, message []byte) {
	h.room <- wsMessageForRoom{
		AppName: os.Getenv("app_name"),
		RoomId:  roomId,
		Message: message,
	}
}

func (h *Hub) broadcastMsg(msg []byte) {
	hub.broadcast <- msg
}

func (h *Hub) run() {
	appName := os.Getenv("app_name")
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
		// broadcast and push message to redis channel
		case message := <-h.broadcast:
			msg := wsBroadcastMessage{
				AppName: appName,
				Message: message,
			}
			b, _ := json.Marshal(msg)
			go infra.GetRedis().Publish(context.Background(), h.psbcchan, b)
			for client := range h.clients {
				select {
				case client.send <- msg.Message:
				default:
					delete(h.clients, client)
					go client.clean()
				}
			}

		// send message for client in this node then push to redis channel
		case message := <-h.room:
			b, _ := json.Marshal(message)
			if err := infra.GetRedis().Publish(context.Background(), h.psroomchan, b).Err(); err != nil {
				logger.Error(err)
			}
			for client := range h.clients {
				ok := client.exist(message.RoomId)
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
			if m.AppName != appName {
				for client := range h.clients {
					ok := client.exist(m.RoomId)
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
			msg := wsBroadcastMessage{}
			if err := json.Unmarshal([]byte(message.Payload), &msg); err != nil {
				logger.Error(err)
			}
			if msg.AppName != appName {
				for client := range h.clients {
					select {
					case client.send <- msg.Message:
					default:
						delete(h.clients, client)
						go client.clean()
					}
				}
			}
		}
	}
}
