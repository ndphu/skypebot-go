package model

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

type ExistingMessage struct {
	Id             string `json:"id"`
	Time           string `json:"composeTime"`
	Content        string `json:"content"`
	ConversationId string `json:"conversationId"`
	MessageType    string `json:"messageType"`
	Type           string `json:"type"`
	Version        string `json:"version"`
	From           string `json:"from"`
}

type MetaData struct {
	SyncState string `json:"syncState"`
	BackwardLink string `json:"backwardLink"`
	LastCompleteSegmentStartTime int `json:"lastCompleteSegmentStartTime"`
	LastCompleteSegmentEndTime int `json:"lastCompleteSegmentEndTime"`
}


type GetMessagesResponse struct {
	Messages []ExistingMessage `json:"messages"`
	MetaData MetaData `json:"_metadata"`
}
