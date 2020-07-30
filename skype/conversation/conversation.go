package conversation

import (
	"encoding/json"
	"github.com/ndphu/skypebot-go/config"
	"github.com/ndphu/skypebot-go/utils"
	"log"
	"net/http"
	"strconv"
)

type Conversation struct {
	Id         string `json:"id"`
	Type       string `json:"type"`
	Version    int    `json:"version"`
	TargetLink string `json:"targetLink"`
	Topic      string `json:"topic"`
}

type ConversationResponse struct {
	Conversations []ConversationObject `json:"conversations"`
}

type ConversationObject struct {
	Id               string           `json:"id"`
	Type             string           `json:"type"`
	Version          int              `json:"version"`
	TargetLink       string           `json:"targetLink"`
	ThreadProperties ThreadProperties `json:"threadProperties"`
}

type ThreadProperties struct {
	Topic string `json:"topic"`
}

func GetConversations(limit int) ([]Conversation, error) {
	conversations := make([]Conversation, 0)
	req, _ := http.NewRequest("GET", config.Get().MessageBaseUrl()+"/v1/users/ME/conversations?view=supportsExtendedHistory%7Cmsnp24Equivalent&startTime=1&targetType=Passport%7CSkype%7CLync%7CThread%7CAgent%7CShortCircuit%7CPSTN%7CFlxt%7CNotificationStream%7CCortanaBot%7CModernBots%7CsecureThreads%7CInviteFree", nil)
	utils.SetRequestHeaders(req)
	q := req.URL.Query()
	q.Set("pageSize", strconv.Itoa(limit))
	req.URL.RawQuery = q.Encode()
	_, _, respBody, err := utils.ExecuteHttpRequestExtended(req)
	if err != nil {
		return nil, err
	}
	//coList := make([]ConversationObject, 0)
	responseObject := ConversationResponse{}
	if err := json.Unmarshal(respBody, &responseObject); err != nil {
		log.Println("Fail to unmarshall data", string(respBody))
		return nil, err
	}
	for _, co := range responseObject.Conversations {
		if co.ThreadProperties.Topic != "" {
			conversations = append(conversations, Conversation{
				Id:         co.Id,
				Type:       co.Type,
				Version:    co.Version,
				TargetLink: co.TargetLink,
				Topic:      co.ThreadProperties.Topic,
			})
		}
	}
	return conversations, nil
}
