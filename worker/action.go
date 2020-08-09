package worker

import (
	"github.com/ndphu/skypebot-go/model"
	_ "image/jpeg"
	"log"
)

func (w *Worker) ProcessMessage(evt *model.MessageEvent) error {
	from := evt.GetFrom()
	threadId := evt.GetThreadId()

	log.Println("Processing message from", from, "on thread", threadId)
	if evt.Type == "EventMessage" && evt.ResourceType == "NewMessage" && evt.Resource.MessageType == "RichText" {
		if w.isMessageFromManager(evt) {
			go w.processManageIM(evt)
			return nil
		}

		if w.isDirectMention(evt) {
			go w.processMention(evt)
			return nil
		}

		if w.isDirectIM(evt) {
			go w.processDirectIM(evt)
			return nil
		}
	}
	return nil
}



