package model

import (
	"errors"
	"sync"
)

const defaultMediaBaseUrl = "https://api.asm.skype.com"

var ErrorEmptySkypeToken = errors.New("empty token")

type Endpoint struct {
	Id             string         `json:"id"`
	Type           string         `json:"type"`
	IsActive       bool           `json:"isActive"`
	ProductContext string         `json:"productContext"`
	Subscriptions  []Subscription `json:"subscriptions"`
}

type Subscription struct {
	Id                  int      `json:"id"`
	Type                string   `json:"type"`
	ChannelType         string   `json:"channelType"`
	ConversationType    int      `json:"conversationType"`
	EventChannel        string   `json:"eventChannel"`
	Template            string   `json:"template"`
	InterestedResources []string `json:"interestedResources"`
}

type Config struct {
	mediaBaseUrl      string       `json:"mediaBaseUrl"`
	messageBaseUrl    string       `json:"messageBaseUrl"`
	registrationToken string       `json:"registrationToken"`
	skypeToken        string       `json:"skypeToken"`
	endpoint          string       `json:"endpoint"`
	lock              sync.RWMutex `json:"_"`
}
