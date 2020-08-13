package model

type Conversation struct {
	Id         string `json:"id"`
	Type       string `json:"type"`
	Version    int64    `json:"version"`
	TargetLink string `json:"targetLink"`
	Topic      string `json:"topic"`
}

type SkypeConversationResponse struct {
	Conversations []ConversationObject `json:"conversations"`
}

type ThreadProperties struct {
	Topic string `json:"topic"`
}

type ConversationObject struct {
	Id               string           `json:"id"`
	Type             string           `json:"type"`
	Version          int64             `json:"version"`
	TargetLink       string           `json:"targetLink"`
	ThreadProperties ThreadProperties `json:"threadProperties"`
}
