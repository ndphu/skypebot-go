package worker

import (
	"encoding/json"
	"fmt"
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/utils"
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
	fmt.Println(string(respBody))
	responseObject := model.SkypeConversationResponse{}
	if err := json.Unmarshal(respBody, &responseObject); err != nil {
		log.Println("Fail to unmarshal data", string(respBody))
		log.Println("unmarshal error", err)
		return nil, err
	}
	for _, co := range responseObject.Conversations {
		if co.ThreadProperties.LastLeaveAt != "" {
			// ignore conversation that already left
			continue
		}
		threadName := co.ThreadProperties.Topic
		cvType := model.ConversationTypeGroup
		if threadName == "" {
			cvType = model.ConversationTypeDirect
			threadName = co.Id
		}
		conversations = append(conversations, model.Conversation{
			Id:         co.Id,
			Type:       cvType,
			Version:    co.Version,
			TargetLink: co.TargetLink,
			Topic:      threadName,
		})
	}
	return conversations, nil
}

func (w *Worker) LeaveConversation(threadId string) error {
	threadId = completeThreadId(threadId)
	botId := completeUserId(w.skypeId)
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/threads/%s/members/%s", w.baseUrl, threadId, botId), nil)
	w.setRequestHeaders(req)

	return utils.ExecuteWithRetry(func() error {
		if _, _, _, err := w.executeHttpRequest(req); err != nil {
			return err
		}
		return nil
	})
}
