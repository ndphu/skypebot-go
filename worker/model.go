package worker

import "github.com/ndphu/skypebot-go/skype/model"

type PollingResponse struct {
	ErrorCode     int            `json:"errorCode"`
	EventMessages []EventMessage `json:"eventMessages"`
}

type Resource struct {
	Id          string                `json:"id"`
	LastMessage model.ExistingMessage `json:"lastMessage"`
}

type EventMessage struct {
	Id           int                `json:"id"`
	ResourceLink string             `json:"resourceLink"`
	ResourceType string             `json:"resourceType"`
	Time         string             `json:"time"`
	Type         string             `json:"type"`
	Resource     NewMessageResource `json:"resource"`
}

type NewMessageResource struct {
	Type             string `json:"type"`
	From             string `json:"from"`
	ClientMessageId  string `json:"clientmessageid"`
	Content          string `json:"content"`
	ContentType      string `json:"contenttype"`
	ThreadTopic      string `json:"thread_topic"`
	ConversationLink string `json:"conversationLink"`
	Id               string `json:"id"`
}

type SubscriptionRequest struct {
	ChannelType         string   `json:"channelType"`
	ConversationType    int      `json:"conversationType"`
	InterestedResources []string `json:"interestedResources"`
}
