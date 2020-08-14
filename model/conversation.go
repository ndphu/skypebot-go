package model

const ConversationTypeGroup = "GROUP"
const ConversationTypeDirect = "DIRECT"

type Conversation struct {
	Id         string `json:"id"`
	Type       string `json:"type"`
	Version    int64  `json:"version"`
	TargetLink string `json:"targetLink"`
	Topic      string `json:"topic"`
}

type SkypeConversationResponse struct {
	Conversations []SkypeConversation `json:"conversations"`
}

type ThreadProperties struct {
	Topic       string `json:"topic"`
	LastLeaveAt string `json:"lastleaveat"`
}

type SkypeConversation struct {
	Id               string                      `json:"id"`
	Type             string                      `json:"type"`
	Version          int64                       `json:"version"`
	TargetLink       string                      `json:"targetLink"`
	ThreadProperties ThreadProperties            `json:"threadProperties"`
	Properties       SkypeConversationProperties `json:"properties"`
}

type SkypeConversationProperties struct {
}
