package handler

import "encoding/json"

const (
	// Message Type
	msgJoinRoom  = "joinRoom"
	msgLeaveRoom = "leaveRoom"
)

type wsMessage struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"raw"`
}

type wsRoomIdsMessage struct {
	Ids []string `json:"ids"`
}

type wsRoomActionMessage struct {
	Leave bool
	Join  bool
	Ids   []string
}
