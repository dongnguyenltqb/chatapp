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

type wsWelcomeMessage struct {
	ClientId string `json:"clientId"`
}

type wsMessageForRoom struct {
	NodeId  string `json:"nodeId"`
	RoomId  string `json:"roomId"`
	Message []byte `json:"message"`
}

type wsBroadcastMessage struct {
	NodeId  string `json:"nodeId"`
	Message []byte `json:"message"`
}

type wsIdentityMessage struct {
	ClientId string `json:"clientId"`
}

type wsRoomIdsMessage struct {
	Ids []string `json:"ids"`
}

type wsRoomActionMessage struct {
	Leave    bool     `json:"leave"`
	Join     bool     `json:"join"`
	Ids      []string `json:"ids"`
	MemberId string   `json:"memberId"`
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

type wsIceCandidateMessage struct {
	TargetID  string          `json:"targetId"`
	Candidate json.RawMessage `json:"candidate"`
}

type wsJoinRoomVideoCallMessage struct {
	RoomId   string `json:"roomId"`
	MemberId string `json:"memberId"`
}
