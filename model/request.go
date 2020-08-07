package model

type PostTextMessageRequest struct {
	Target string `json:"target"`
	Text   string `json:"text"`
}

type ReactMessageRequest struct {
	Target    string `json:"target"`
	Emotion   string `json:"emotion"`
	MessageId string `json:"messageId"`
}
