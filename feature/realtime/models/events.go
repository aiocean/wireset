package models

const TopicUserJoined WebsocketTopic = "UserJoined"

type UserJoinedEvt struct {
	UserName string
	RoomID   string
}

const WebsocketEndpoint = "/api/v1/ws"