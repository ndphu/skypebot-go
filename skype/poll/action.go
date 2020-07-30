package poll

import (
	"github.com/ndphu/skypebot-go/skype/chat"
	"log"
	"path"
	"strings"
)

func ProcessMessage(evt EventMessage) error {
	threadId, from := parseInfo(evt)

	log.Println("Processing message from", from, "on thread", threadId)
	fromUser := strings.TrimPrefix(from, "8:")

	actions := IsRuleMatched(threadId, fromUser)
	if len(actions) == 0 {
		log.Println("No action taken")
		return nil
	}
	for _, action := range actions {
		TakeAction(evt, action)
	}
	return nil
}

func parseInfo(evt EventMessage) (string, string) {
	threadId := path.Base(evt.Resource.ConversationLink)
	from := path.Base(evt.Resource.From)
	return threadId, from
}

func TakeAction(evt EventMessage, action Action) {
	if action.Type == ActionTypeReact {
		threadId, _ := parseInfo(evt)
		for _, emotion := range action.Data["emotions"].([]interface{}) {
			log.Println("Reacting with emotion", emotion)
			chat.ReactMessage(threadId, evt.Resource.Id, emotion.(string))
		}
	}
}
