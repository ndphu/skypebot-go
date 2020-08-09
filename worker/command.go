package worker

import (
	"github.com/ndphu/skypebot-go/model"
	"log"
	"path"
	"strings"
)

func (w *Worker) isDirectIM(event *model.MessageEvent) bool {
	return path.Base(event.Resource.ConversationLink) == path.Base(event.Resource.From)
}

func (w *Worker) isMessageFromManager(event *model.MessageEvent) bool {
	return w.isDirectIM(event) && contains(w.managers, path.Base(event.Resource.From))
}

func normalizeMessageContent(content string) string {
	commandString := strings.ReplaceAll(content, "-", "")
	commandString = strings.TrimSpace(commandString)
	commandString = strings.ToLower(commandString)
	log.Println("Normalized command:", commandString)
	return commandString
}

