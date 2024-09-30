package models

const TopicUserJoined WebsocketTopic = "UserJoined"

type UserJoinedEvt struct {
	UserName string
	RoomID   string
}
