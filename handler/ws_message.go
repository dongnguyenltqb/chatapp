package handler

import "encoding/json"

const (
	// Message Type
	msgJoinRoom          = "joinRoom"
	msgLeaveRoom         = "leaveRoom"
	msgIdentity          = "identity"
	msgChat              = "chat"
	msgOffer             = "offer"
	msgAnswer            = "answer"
	msgIceCandidate      = "icecandidate"
	msgJoinRoomVideoCall = "joinRoomVideoCall"
)

type wsMessage struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"raw"`
}

type wsMessageForRoom struct {
	RoomId  string
	Message []byte
}

type wsIdentityMessage struct {
	ClientId string `json:"clientId"`
}

type wsRoomIdsMessage struct {
	Ids []string `json:"ids"`
}

type wsRoomActionMessage struct {
	Leave    bool
	Join     bool
	Ids      []string
	MemberId string
}

type wsChatMessage struct {
	Text string `json:"text"`
}

type wsOfferMessage struct {
	TargetID string          `json:"targetId"`
	Offer    json.RawMessage `json:"offer"`
}

type wsAnswerMessage struct {
	TargetID string          `json:"targetId"`
	Answer   json.RawMessage `json:"answer"`
}

type wsJoinRoomVideoCallMessage struct {
	RoomId   string `json:"roomId"`
	MemberId string `json:"memberId"`
}
