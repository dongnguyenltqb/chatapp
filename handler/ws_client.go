// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 1024

	// Max room size
	maxRoomSize = 100
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
	rchan chan wsRoomActionMessage

	// Room
	rooms []string

	// Logger
	logger *logrus.Logger
}

// clean delete everything of client after a hour
func (c *Client) clean() {
	<-time.After(time.Minute)
	close(c.send)
	close(c.rchan)
	c.rooms = nil
}

// roomPump pumps action for channel and process one by one
func (c *Client) roomPump() {
	for {
		select {
		case msg, ok := <-c.rchan:
			if !ok {
				// The hub closed the channel
				return
			}
			if msg.Join {
				for i := 0; i < len(msg.Ids); i++ {
					id := msg.Ids[i]
					c.join(id)
				}
			}
			if msg.Leave {
				for i := 0; i < len(msg.Ids); i++ {
					id := msg.Ids[i]
					c.leave(id)
				}
			}
		}
	}
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
				c.logger.Error(err)
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
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
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

func (c *Client) welcome() {
	welMsg := wsWelcomeMessage{
		ClientId: c.id,
	}
	b, _ := json.Marshal(welMsg)
	msg := wsMessage{
		Type: "welcome",
		Raw:  b,
	}
	b, _ = json.Marshal(msg)
	c.broadcastMsg(b)
}

func (c *Client) join(roomId string) {
	n := len(c.rooms)
	joined := false
	for i := 0; i < n; i++ {
		if c.rooms[i] == roomId {
			joined = true
		}
	}
	if !joined {
		c.rooms = append(c.rooms, roomId)
	}
}

func (c *Client) leave(roomId string) {
	n := len(c.rooms)
	r := make([]string, 0)
	for i := 0; i < n; i++ {
		if c.rooms[i] != roomId {
			r = append(r, c.rooms[i])
		}
	}
	c.rooms = r
}

func (c *Client) exist(roomId string) bool {
	n := len(c.rooms)
	for i := 0; i < n; i++ {
		if c.rooms[i] == roomId {
			return true
		}
	}
	return false
}

func (c *Client) sendMsg(msg []byte) {
	c.send <- msg
}

func (c *Client) sendMsgToRoom(roomId string, message []byte) {
	c.hub.sendMsgToRoom(roomId, message)
}

func (c *Client) broadcastMsg(msg []byte) {
	c.hub.broadcastMsg(msg)
}

func (c *Client) sendIdentityMsg() {
	clientId := wsIdentityMessage{
		ClientId: c.id,
	}
	b, _ := json.Marshal(clientId)
	msg := wsMessage{
		Type: msgIdentity,
		Raw:  b,
	}
	b, _ = json.Marshal(msg)
	go c.sendMsg(b)
}

func (c *Client) processRoomActionMsg(message wsMessage) {
	rMsg := wsRoomActionMessage{}
	if err := json.Unmarshal(message.Raw, &rMsg); err != nil {
		return
	}
	if message.Type == msgJoinRoom {
		rMsg.Join = true
	}
	if message.Type == msgLeaveRoom {
		rMsg.Leave = true
	}

	// process leave/join
	c.rchan <- rMsg
	// prepare message to send to each group
	rMsg.MemberId = c.id
	msgType := msgJoinRoom
	if rMsg.Leave {
		msgType = msgLeaveRoom
	}
	// Loop and send mesage for each room
	for _, roomId := range rMsg.Ids {
		rMsg.Ids = []string{roomId}
		b, _ := json.Marshal(rMsg)
		m := wsMessage{
			Type: msgType,
			Raw:  b,
		}
		r, _ := json.Marshal(&m)
		c.sendMsgToRoom(roomId, r)
		// if message type is leave, emit to this guy
		if msgType == msgLeaveRoom {
			c.sendMsgToRoom(c.id, r)
		}
	}
}

func (c *Client) processChatMsg(message []byte) {
	for _, roomId := range c.rooms {
		if roomId != c.id {
			go c.sendMsgToRoom(roomId, message)
		}
	}
}

func (c *Client) processOfferMsg(msg wsMessage) {
	offer := wsOfferMessage{}
	if err := json.Unmarshal(msg.Raw, &offer); err != nil {
		c.logger.Error(err)
		return
	}
	targetId := offer.TargetID
	offer.TargetID = c.id
	b, _ := json.Marshal(offer)
	m := wsMessage{
		Type: "offer",
		Raw:  b,
	}
	b, _ = json.Marshal(m)
	c.sendMsgToRoom(targetId, b)
}

func (c *Client) processAnswerMsg(msg wsMessage) {
	answer := wsAnswerMessage{}
	if err := json.Unmarshal(msg.Raw, &answer); err != nil {
		c.logger.Error(err)
		return
	}
	targetId := answer.TargetID
	answer.TargetID = c.id
	b, _ := json.Marshal(answer)
	m := wsMessage{
		Type: "answer",
		Raw:  b,
	}
	b, _ = json.Marshal(m)
	c.sendMsgToRoom(targetId, b)
}

func (c *Client) processIceCandidateMsg(msg wsMessage) {
	candidate := wsIceCandidateMessage{}
	if err := json.Unmarshal(msg.Raw, &candidate); err != nil {
		c.logger.Error(err)
		return
	}
	targetId := candidate.TargetID
	candidate.TargetID = c.id
	b, _ := json.Marshal(candidate)
	m := wsMessage{
		Type: "icecandidate",
		Raw:  b,
	}
	b, _ = json.Marshal(m)
	c.sendMsgToRoom(targetId, b)
}

func (c *Client) processJoinVideoCallMsg(msg wsMessage) {
	j := wsJoinRoomVideoCallMessage{}
	if err := json.Unmarshal(msg.Raw, &j); err != nil {
		c.logger.Error(err)
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

// process message from readPump
func (c *Client) processMsg(message []byte) {
	msg := wsMessage{}
	if err := json.Unmarshal(message, &msg); err != nil {
		c.logger.Error(err)
		return
	}
	// Handle chat message, broadcast room
	if msg.Type == msgChat {
		c.processChatMsg(message)
	}
	// Handle room action
	if msg.Type == msgJoinRoom || msg.Type == msgLeaveRoom {
		c.processRoomActionMsg(msg)
	}
	// Handle join room video call
	if msg.Type == msgJoinRoomVideoCall {
		c.processJoinVideoCallMsg(msg)
	}
	// Handle RTC message
	if msg.Type == msgOffer {
		c.processOfferMsg(msg)
	}
	if msg.Type == msgAnswer {
		c.processAnswerMsg(msg)
	}
	if msg.Type == msgIceCandidate {
		c.processIceCandidateMsg(msg)
	}
}
