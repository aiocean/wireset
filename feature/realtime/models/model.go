package models

type WebsocketTopic string

// Stringer interface
func (t WebsocketTopic) String() string {
	return string(t)
}

// FromString
func FromString(s string) WebsocketTopic {
	return WebsocketTopic(s)
}

type WebsocketMessage[T any] struct {
	Topic   WebsocketTopic `json:"topic"`
	Payload T              `json:"payload"`
}

const TopicError WebsocketTopic = "error"

type ErrorPayload struct {
	Message string `json:"message"`
}
