package worker

import (
	"encoding/json"
	"github.com/ndphu/skypebot-go/model"
	"log"
	"net/http"
	"strconv"
)

func (w *Worker) GetConversations(limit int) ([]model.Conversation, error) {
	conversations := make([]model.Conversation, 0)
	req, _ := http.NewRequest("GET", w.baseUrl+"/v1/users/ME/conversations?view=supportsExtendedHistory%7Cmsnp24Equivalent&startTime=1&targetType=Passport%7CSkype%7CLync%7CThread%7CAgent%7CShortCircuit%7CPSTN%7CFlxt%7CNotificationStream%7CCortanaBot%7CModernBots%7CsecureThreads%7CInviteFree", nil)
	w.setRequestHeaders(req)
	q := req.URL.Query()
	q.Set("pageSize", strconv.Itoa(limit))
	req.URL.RawQuery = q.Encode()
	_, _, respBody, err := w.executeHttpRequest(req)
	if err != nil {
		return nil, err
	}
	responseObject := model.SkypeConversationResponse{}
	if err := json.Unmarshal(respBody, &responseObject); err != nil {
		log.Println("Fail to unmarshall data", string(respBody))
		return nil, err
	}
	for _, co := range responseObject.Conversations {
		if co.ThreadProperties.Topic != "" {
			conversations = append(conversations, model.Conversation{
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
