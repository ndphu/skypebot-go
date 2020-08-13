package model

import "path"

type CreateObjectRequest struct {
	Type        string              `json:"type"`
	Filename    string              `json:"filename"`
	Permissions map[string][]string `json:"permissions"`
}

type PostMessageRequest struct {
	MessageId     string `json:"clientmessageid"`
	ComposeTime   string
	Content       string   `json:"content"`
	MessageType   string   `json:"messagetype"`
	ContentType   string   `json:"contenttype"`
	DisplayName   string   `json:"imdisplayname"`
	AsmReferences []string `json:"amsreferences"`
}

type CreateObjectResponse struct {
	Id string `json:"id"`
}

type SkypeMessage struct {
	Id             string `json:"id"`
	Time           string `json:"composeTime"`
	Content        string `json:"content"`
	ConversationId string `json:"conversationId"`
	MessageType    string `json:"messageType"`
	Type           string `json:"type"`
	Version        string `json:"version"`
	From           string `json:"from"`
	SkypeEditedId  string `json:"skypeeditedid"`
}

type MetaData struct {
	SyncState                    string `json:"syncState"`
	BackwardLink                 string `json:"backwardLink"`
	LastCompleteSegmentStartTime int    `json:"lastCompleteSegmentStartTime"`
	LastCompleteSegmentEndTime   int    `json:"lastCompleteSegmentEndTime"`
}

type GetMessagesResponse struct {
	Messages []SkypeMessage `json:"messages"`
	MetaData MetaData       `json:"_metadata"`
}

type PollingResponse struct {
	ErrorCode int            `json:"errorCode"`
	Events    []MessageEvent `json:"eventMessages"`
}

type Resource struct {
	Id          string       `json:"id"`
	LastMessage SkypeMessage `json:"lastMessage"`
}

type MessageEvent struct {
	Id           int             `json:"id"`
	ResourceLink string          `json:"resourceLink"`
	ResourceType string          `json:"resourceType"`
	Time         string          `json:"time"`
	Type         string          `json:"type"`
	Resource     MessageResource `json:"resource"`
}

func (e *MessageEvent) GetThreadId() string {
	return path.Base(e.Resource.ConversationLink)
}

func (e *MessageEvent) GetFrom() string {
	return path.Base(e.Resource.From)
}

type MessageResource struct {
	Type             string `json:"type"`
	From             string `json:"from"`
	ClientMessageId  string `json:"clientmessageid"`
	Content          string `json:"content"`
	ContentType      string `json:"contenttype"`
	ThreadTopic      string `json:"thread_topic"`
	ConversationLink string `json:"conversationLink"`
	Id               string `json:"id"`
	MessageType      string `json:"messagetype"`
}

type SubscriptionRequest struct {
	ChannelType         string   `json:"channelType"`
	ConversationType    int      `json:"conversationType"`
	InterestedResources []string `json:"interestedResources"`
}

type SkypeError struct {
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}
