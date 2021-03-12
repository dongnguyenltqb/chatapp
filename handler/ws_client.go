// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"chatapp/logger"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// Identity
	id string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Room channel event
	roomActionChan chan wsRoomActionMessage

	// Room
	rooms map[string]bool
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		go c.processMsg(message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendMsgToRoom(roomId string, message []byte) {
	c.hub.room <- wsMessageForRoom{
		RoomId:  roomId,
		Message: message,
	}
}

func (c *Client) broadcastMsg(msg []byte) {
	c.hub.broadcast <- msg
}

func (c *Client) sendIdentityMsg() {
	// Emit clientId to front end
	clientId := wsIdentityMessage{
		ClientId: c.id,
	}
	b, _ := json.Marshal(clientId)
	msg := wsMessage{
		Type: msgIdentity,
		Raw:  b,
	}
	b, _ = json.Marshal(msg)
	c.send <- b
}

func (c *Client) processRoomAction() {
	for {
		select {
		case msg, ok := <-c.roomActionChan:
			if !ok {
				// The hub closed the channel
				return
			}
			if msg.Join == true {
				for i := 0; i < len(msg.Ids); i++ {
					id := msg.Ids[i]
					c.rooms[id] = true
				}
			}
			if msg.Leave == true {
				for i := 0; i < len(msg.Ids); i++ {
					id := msg.Ids[i]
					delete(c.rooms, id)
				}
			}
		}
	}
}

// process message from readPump
func (c *Client) processMsg(message []byte) {
	logger := logger.Get()
	msg := wsMessage{}
	if err := json.Unmarshal(message, &msg); err != nil {
		logger.Error(err)
		return
	}

	// Handle room action
	if msg.Type == msgJoinRoom {
		r := wsRoomActionMessage{}
		if err := json.Unmarshal(msg.Raw, &r); err != nil {
			return
		}
		c.roomActionChan <- wsRoomActionMessage{
			Join: true,
			Ids:  r.Ids,
		}
	}
	if msg.Type == msgLeaveRoom {
		r := wsRoomActionMessage{}
		if err := json.Unmarshal(msg.Raw, &r); err != nil {
			return
		}
		c.roomActionChan <- wsRoomActionMessage{
			Leave: true,
			Ids:   r.Ids,
		}
	}

	// Handle chat message, broadcast to all connected client
	if msg.Type == msgChat {
		b, _ := json.Marshal(msg)
		for roomId := range c.rooms {
			go c.sendMsgToRoom(roomId, b)
		}
	}

	// Handle join room video call
	if msg.Type == msgJoinRoomVideoCall {
		j := wsJoinRoomVideoCallMessage{}
		if err := json.Unmarshal(msg.Raw, &j); err != nil {
			logger.Error(err)
			return
		}
		j.MemberId = c.id
		b, _ := json.Marshal(wsJoinRoomVideoCallMessage{
			RoomId:   j.RoomId,
			MemberId: c.id,
		})
		b, _ = json.Marshal(wsMessage{
			Type: msgJoinRoomVideoCall,
			Raw:  b,
		})
		c.sendMsgToRoom(j.RoomId, b)
	}
}
