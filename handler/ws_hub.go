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
	"github.com/sirupsen/logrus"
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
	client.welcome()

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
	pubSubRoomChannel      string
	pubSubBroadcastChannel string
	subscribeRoomChan      <-chan *redis.Message
	subscribeBroadcastChan <-chan *redis.Message

	// Logger
	logger *logrus.Logger
}

var hub *Hub
var onceInitHub sync.Once

// getHub return singleton hub
func getHub() *Hub {
	onceInitHub.Do(func() {
		pubSubRoomChannel := "chat_app_room_chan"
		pubSubBroadcastChannel := "chat_app_broadcast_chan"

		redisSubscribeRoom := infra.GetRedis().Subscribe(context.Background(), pubSubRoomChannel)
		redisSubscribeBroadcast := infra.GetRedis().Subscribe(context.Background(), pubSubBroadcastChannel)

		hub = &Hub{
			broadcast:              make(chan []byte),
			room:                   make(chan wsMessageForRoom),
			register:               make(chan *Client),
			unregister:             make(chan *Client),
			clients:                make(map[*Client]bool),
			pubSubRoomChannel:      "chat_app_room_chan",
			pubSubBroadcastChannel: "chat_app_broadcast_chan",
			subscribeRoomChan:      redisSubscribeRoom.Channel(),
			subscribeBroadcastChan: redisSubscribeBroadcast.Channel(),
			logger:                 logger.Get(),
		}
		go hub.run()
	})
	return hub
}

func (h *Hub) sendMsgToRoom(roomId string, message []byte) {
	h.room <- wsMessageForRoom{
		NodeId:  os.Getenv("node_id"),
		RoomId:  roomId,
		Message: message,
	}
}

func (h *Hub) broadcastMsg(msg []byte) {
	hub.broadcast <- msg
}

func (h *Hub) run() {
	nodeId := os.Getenv("node_id")
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
				NodeId:  nodeId,
				Message: message,
			}
			b, _ := json.Marshal(msg)
			go infra.GetRedis().Publish(context.Background(), h.pubSubBroadcastChannel, b)
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
			if err := infra.GetRedis().Publish(context.Background(), h.pubSubRoomChannel, b).Err(); err != nil {
				h.logger.Error(err)
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
		case message := <-h.subscribeRoomChan:
			m := wsMessageForRoom{}
			if err := json.Unmarshal([]byte(message.Payload), &m); err != nil {
				h.logger.Error(err)
			}
			if m.NodeId != nodeId {
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
		case message := <-h.subscribeBroadcastChan:
			msg := wsBroadcastMessage{}
			if err := json.Unmarshal([]byte(message.Payload), &msg); err != nil {
				h.logger.Error(err)
			}
			if msg.NodeId != nodeId {
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
